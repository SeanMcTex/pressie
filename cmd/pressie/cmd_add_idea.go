package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/gifts"
	"github.com/SeanMcTex/pressie/internal/store"
)

// cmdAddIdea adds a gift idea for a contact or "anyone" (general ideas).
func cmdAddIdea(args []string) {
	if wantsHelp(args) {
		helpAddIdea()
		return
	}
	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.Item == "" {
		fmt.Fprintln(os.Stderr, "pressie: --item is required")
		os.Exit(1)
	}

	if cfg.For == "" {
		fmt.Fprintln(os.Stderr, "pressie: --for is required (use \"anyone\" for a general idea)")
		os.Exit(1)
	}

	// Build the idea.
	idea := store.Idea{
		ID:     store.NewUUID(),
		Item:   cfg.Item,
		URL:    cfg.URL,
		Tags:   cfg.Tags,
		Status: "open",
		Added:  time.Now().UTC().Format("2006-01-02"),
		Notes:  cfg.Notes,
		Images: []string{},
	}
	if cfg.HasPrice {
		p := cfg.Price
		idea.PriceEstimate = &p
		idea.Currency = "USD"
	}

	if cfg.For == "anyone" {
		addGeneralIdea(cfg, &idea)
	} else {
		addContactIdea(cfg, &idea)
	}
}

// addGeneralIdea appends an idea to _ideas-general.json.
func addGeneralIdea(cfg *config.Config, idea *store.Idea) {
	ideas, err := store.LoadGeneralIdeas(cfg.GiftsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	ideas = append(ideas, *idea)

	if err := store.SaveGeneralIdeas(cfg.GiftsDir, ideas); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Added general idea: %s\n", idea.Item)
}

// addContactIdea resolves the contact, ensures their file exists, appends
// the idea, and saves. Also registers the contact in _index.json if new.
func addContactIdea(cfg *config.Config, idea *store.Idea) {
	idx := requireIndex(cfg.GiftsDir)

	key, name, relPath, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.For, cfg.Visibility)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}
	rejectArchived(idx, key, name)

	cf, err := gifts.EnsureContactFile(cfg.GiftsDir, relPath, key, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	cf.Ideas = append(cf.Ideas, *idea)
	cf.Tags = mergeTags(cf.Tags, idea.Tags)

	absPath := filepath.Join(cfg.GiftsDir, relPath)
	if err := store.SaveContactFile(absPath, cf); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	// Register in index if new, and save index.
	gifts.RegisterContact(idx, key, relPath, cfg.Visibility, cfg.Tags)
	if err := store.SaveIndex(cfg.GiftsDir, idx); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Added idea for %s: %s\n", name, idea.Item)
}

// mergeTags returns the union of existing and new tags, case-insensitive, preserving order.
func mergeTags(existing, additions []string) []string {
	seen := make(map[string]bool, len(existing)+len(additions))
	merged := make([]string, 0, len(existing)+len(additions))
	for _, t := range existing {
		k := strings.ToLower(t)
		if !seen[k] {
			seen[k] = true
			merged = append(merged, t)
		}
	}
	for _, t := range additions {
		k := strings.ToLower(t)
		if !seen[k] {
			seen[k] = true
			merged = append(merged, t)
		}
	}
	return merged
}