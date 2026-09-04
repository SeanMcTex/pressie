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

// cmdEditGift edits fields of an existing logged gift by ID. Searches
// both gifts given and received for the contact (--to/--from/--for all
// resolve the same way; use whichever reads naturally).
func cmdEditGift(args []string) {
	if wantsHelp(args) {
		printEditHelp("gift")
		return
	}

	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.ID == "" {
		fmt.Fprintln(os.Stderr, "pressie: --id <gift-id> is required")
		os.Exit(1)
	}
	who := cfg.For
	if who == "" {
		who = cfg.To
	}
	if who == "" {
		who = cfg.From
	}
	if who == "" {
		fmt.Fprintln(os.Stderr, "pressie: --for <name> (or --to/--from) is required")
		os.Exit(1)
	}
	cfg.For = who

	if !cfg.HasItem && !cfg.HasOccasion && !cfg.HasDate && !cfg.HasPrice && cfg.Notes == "" {
		fmt.Fprintln(os.Stderr, "pressie: nothing to edit — pass at least one of --item, --occasion, --date, --price, --notes")
		os.Exit(1)
	}

	idx := requireIndex(cfg.GiftsDir)

	_, name, relPath, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.For, cfg.Visibility)
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

	edited := ""
	for i := range cf.GiftsGiven {
		if cf.GiftsGiven[i].ID == cfg.ID {
			applyGiftEdit(&cf.GiftsGiven[i], cfg)
			edited = "given"
			break
		}
	}
	if edited == "" {
		for i := range cf.GiftsReceived {
			if cf.GiftsReceived[i].ID == cfg.ID {
				applyGiftEdit(&cf.GiftsReceived[i], cfg)
				edited = "received"
				break
			}
		}
	}
	if edited == "" {
		fmt.Fprintf(os.Stderr, "pressie: no gift with id %s found for %s\n", cfg.ID, name)
		os.Exit(1)
	}

	if err := store.SaveContactFile(absPath, cf); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Edited gift %s for %s.\n", edited, name)

}

func applyGiftEdit(gift *store.Gift, cfg *config.Config) {
	if cfg.HasItem {
		gift.Item = cfg.Item
	}
	if cfg.HasOccasion {
		gift.Occasion = cfg.Occasion
	}
	if cfg.HasDate {
		gift.Date = cfg.Date
	}
	if cfg.HasPrice {
		if cfg.Price < 0 {
			gift.Price = nil
			gift.Currency = ""
		} else {
			gift.Price = &cfg.Price
			if gift.Currency == "" {
				gift.Currency = "USD"
			}
		}
	}
	if cfg.Notes != "" {
		gift.Notes = cfg.Notes
	}
}