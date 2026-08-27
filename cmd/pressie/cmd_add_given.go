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

// cmdAddGiven logs a gift given to a contact.
// If the gift item matches an open idea for that contact, the idea's
// status is flipped to "purchased" so it stops surfacing in ideas queries.
func cmdAddGiven(args []string) {
	if wantsHelp(args) {
		helpAddGiven()
		return
	}
	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.To == "" {
		fmt.Fprintln(os.Stderr, "pressie: --to is required")
		os.Exit(1)
	}
	if cfg.Item == "" {
		fmt.Fprintln(os.Stderr, "pressie: --item is required")
		os.Exit(1)
	}

	date := cfg.Date
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	gift := store.Gift{
		ID:       store.NewUUID(),
		Date:     date,
		Occasion: cfg.Occasion,
		Item:     cfg.Item,
		Notes:    cfg.Notes,
		Images:   []string{},
		Source:   "manual",
		Added:    time.Now().UTC().Format(time.RFC3339),
	}
	if cfg.HasPrice {
		p := cfg.Price
		gift.Price = &p
		gift.Currency = "USD"
	}

	idx, err := store.LoadIndex(cfg.GiftsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	key, name, relPath, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.To, cfg.Visibility)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	cf, err := gifts.EnsureContactFile(cfg.GiftsDir, relPath, key, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	cf.GiftsGiven = append(cf.GiftsGiven, gift)

	// Retire matching open ideas.
	retired := 0
	if cfg.Idea != "" {
		// Precise retirement by idea ID.
		for i := range cf.Ideas {
			if cf.Ideas[i].ID == cfg.Idea && cf.Ideas[i].Status == "open" {
				cf.Ideas[i].Status = "purchased"
				retired++
			}
		}
		if retired == 0 {
			fmt.Fprintf(os.Stderr, "pressie: no open idea with id %s found for %s\n", cfg.Idea, name)
			os.Exit(1)
		}
	} else {
		// Fuzzy retirement by item text match.
		for i := range cf.Ideas {
			if cf.Ideas[i].Status == "open" && itemsMatch(cf.Ideas[i].Item, gift.Item) {
				cf.Ideas[i].Status = "purchased"
				retired++
			}
		}
	}

	absPath := filepath.Join(cfg.GiftsDir, relPath)
	if err := store.SaveContactFile(absPath, cf); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	gifts.RegisterContact(idx, key, relPath, cfg.Visibility, nil)
	if err := store.SaveIndex(cfg.GiftsDir, idx); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Logged gift for %s: %s\n", name, gift.Item)
	if retired > 0 {
		fmt.Printf("Retired %d idea(s) matching this gift.\n", retired)
	}
}

// itemsMatch returns true if two item descriptions are the same gift,
// using normalized exact or substring matching (case-insensitive).
func itemsMatch(a, b string) bool {
	na := normalizeItem(a)
	nb := normalizeItem(b)
	if na == nb {
		return true
	}
	if na == "" || nb == "" {
		return false
	}
	return strings.Contains(na, nb) || strings.Contains(nb, na)
}

// normalizeItem lowercases, trims, and collapses whitespace.
func normalizeItem(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}