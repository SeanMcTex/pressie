package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/gifts"
	"github.com/SeanMcTex/pressie/internal/store"
)

// cmdDeleteIdea removes an idea by ID. If --for is given, removes from that
// contact's file and recomputes their tags from remaining ideas. If --for is
// not given, removes from _ideas-general.json.
func cmdDeleteIdea(args []string) {
	if wantsHelp(args) {
		helpDeleteIdea()
		return
	}

	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.Idea == "" {
		fmt.Fprintln(os.Stderr, "pressie: --idea <id> is required")
		os.Exit(1)
	}

	if cfg.For == "" || cfg.For == "anyone" {
		deleteGeneralIdea(cfg)
	} else {
		deleteContactIdea(cfg)
	}
}

// deleteGeneralIdea removes an idea from _ideas-general.json by ID.
func deleteGeneralIdea(cfg *config.Config) {
	ideas, err := store.LoadGeneralIdeas(cfg.GiftsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	found := false
	filtered := make([]store.Idea, 0, len(ideas))
	for _, idea := range ideas {
		if idea.ID == cfg.Idea {
			found = true
			continue
		}
		filtered = append(filtered, idea)
	}

	if !found {
		fmt.Fprintf(os.Stderr, "pressie: no general idea with id %s found\n", cfg.Idea)
		os.Exit(1)
	}

	if err := store.SaveGeneralIdeas(cfg.GiftsDir, filtered); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deleted general idea %s.\n", cfg.Idea)
}

// deleteContactIdea removes an idea from a contact's file by ID, then
// recomputes the contact's tags from all remaining ideas.
func deleteContactIdea(cfg *config.Config) {
	idx := requireIndex(cfg.GiftsDir)

	key, name, relPath, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.For, cfg.Visibility)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	absPath := filepath.Join(cfg.GiftsDir, relPath)
	cf, err := store.LoadContactFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: no contact file found for %s\n", name)
		os.Exit(1)
	}

	found := false
	filtered := make([]store.Idea, 0, len(cf.Ideas))
	for _, idea := range cf.Ideas {
		if idea.ID == cfg.Idea {
			found = true
			continue
		}
		filtered = append(filtered, idea)
	}

	if !found {
		fmt.Fprintf(os.Stderr, "pressie: no idea with id %s found for %s\n", cfg.Idea, name)
		os.Exit(1)
	}

	cf.Ideas = filtered

	// Recompute tags from remaining ideas.
	cf.Tags = recomputeTags(filtered)

	if err := store.SaveContactFile(absPath, cf); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	// Update index tags to match.
	if m, ok := idx.Contacts[key]; ok {
		m.Tags = cf.Tags
		idx.Contacts[key] = m
		if err := store.SaveIndex(cfg.GiftsDir, idx); err != nil {
			fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Deleted idea %s for %s.\n", cfg.Idea, name)
}

// recomputeTags returns the union of all tags across the given ideas,
// case-insensitive, preserving first-seen order.
func recomputeTags(ideas []store.Idea) []string {
	seen := make(map[string]bool)
	tags := make([]string, 0)
	for _, idea := range ideas {
		for _, t := range idea.Tags {
			k := strings.ToLower(t)
			if !seen[k] {
				seen[k] = true
				tags = append(tags, t)
			}
		}
	}
	return tags
}