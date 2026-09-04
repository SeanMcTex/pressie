package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/gifts"
	"github.com/SeanMcTex/pressie/internal/store"
)

// cmdEditIdea edits fields of an existing idea by ID. Works on both
// per-contact ideas (--for <name>) and general ideas (--for anyone or
// omitted). Only explicitly passed flags change.
func cmdEditIdea(args []string) {
	if wantsHelp(args) {
		printEditHelp("idea")
		return
	}

	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.ID == "" {
		fmt.Fprintln(os.Stderr, "pressie: --id <idea-id> is required")
		os.Exit(1)
	}
	if !cfg.HasItem && !cfg.HasURL && !cfg.HasTags && !cfg.HasPrice && !cfg.HasDate {
		fmt.Fprintln(os.Stderr, "pressie: nothing to edit — pass at least one of --item, --url, --tags, --price, --date")
		os.Exit(1)
	}

	applyIdeaEdit := func(idea *store.Idea) {
		if cfg.HasItem {
			idea.Item = cfg.Item
		}
		if cfg.HasURL {
			idea.URL = cfg.URL
		}
		if cfg.HasTags {
			idea.Tags = cfg.Tags
		}
		if cfg.HasPrice {
			if cfg.Price < 0 {
				idea.PriceEstimate = nil
				idea.Currency = ""
			} else {
				idea.PriceEstimate = &cfg.Price
				if idea.Currency == "" {
					idea.Currency = "USD"
				}
			}
		}
		if cfg.HasDate {
			idea.Added = cfg.Date
		}
	}

	if cfg.For == "" || cfg.For == "anyone" {
		editGeneralIdea(cfg, applyIdeaEdit)
	} else {
		editContactIdea(cfg, applyIdeaEdit)
	}
}

func editGeneralIdea(cfg *config.Config, apply func(*store.Idea)) {
	ideas, err := store.LoadGeneralIdeas(cfg.GiftsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	edited := false
	for i := range ideas {
		if ideas[i].ID == cfg.ID {
			apply(&ideas[i])
			fmt.Printf("Edited general idea: %s\n", ideas[i].Item)
			edited = true
			break
		}
	}
	if !edited {
		fmt.Fprintf(os.Stderr, "pressie: no general idea with id %s\n", cfg.ID)
		os.Exit(1)
	}

	if err := store.SaveGeneralIdeas(cfg.GiftsDir, ideas); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}
}

func editContactIdea(cfg *config.Config, apply func(*store.Idea)) {
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

	edited := false
	for i := range cf.Ideas {
		if cf.Ideas[i].ID == cfg.ID {
			apply(&cf.Ideas[i])
			fmt.Printf("Edited idea for %s: %s\n", name, cf.Ideas[i].Item)
			edited = true
			break
		}
	}
	if !edited {
		fmt.Fprintf(os.Stderr, "pressie: no idea with id %s found for %s\n", cfg.ID, name)
		os.Exit(1)
	}

	// Recompute tags so the tag profile reflects the edited idea.
	if cf.Tags != nil || cfg.HasTags {
		cf.Tags = recomputeTags(cf.Ideas)
	}

	if err := store.SaveContactFile(absPath, cf); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	// Keep index tags in sync.
	if m, ok := idx.Contacts[key]; ok {
		m.Tags = cf.Tags
		idx.Contacts[key] = m
		if err := store.SaveIndex(cfg.GiftsDir, idx); err != nil {
			fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
			os.Exit(1)
		}
	}
}

func printEditHelp(kind string) {
	fmt.Printf(`pressie edit-%s — edit an existing %s

Usage:
  pressie edit-%s --id <id> [--item <text>] [--tags <a,b,c>] [--url <url>] [--price <number|-1>] [--date <YYYY-MM-DD>]
  pressie edit-%s --id <id> --for <name> ...   # per-contact

Only the flags you pass are changed. Use --price -1 to clear the price.
For per-contact ideas, pass --for <name>; omit it (or use "anyone")
to edit a general idea.

Examples:
  pressie edit-idea --id 4e72649a --item "Letterpress print, A4"
  pressie edit-idea --id 4e72649a --for Kris --price 85
`, kind, kind, kind, kind)
}