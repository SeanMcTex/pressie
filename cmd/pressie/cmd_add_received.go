package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/gifts"
	"github.com/SeanMcTex/pressie/internal/store"
)

// cmdAddReceived logs a gift received from a contact.
func cmdAddReceived(args []string) {
	if wantsHelp(args) {
		helpAddReceived()
		return
	}

	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.From == "" {
		fmt.Fprintln(os.Stderr, "pressie: --from is required")
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

	key, name, relPath, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.From, cfg.Visibility)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	cf, err := gifts.EnsureContactFile(cfg.GiftsDir, relPath, key, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	cf.GiftsReceived = append(cf.GiftsReceived, gift)

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

	fmt.Printf("Logged gift received from %s: %s\n", name, gift.Item)
}