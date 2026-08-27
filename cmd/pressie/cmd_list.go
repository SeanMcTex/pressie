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

// cmdList lists gifts given and/or received for a contact.
func cmdList(args []string) {
	if wantsHelp(args) {
		helpList()
		return
	}

	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.For == "" {
		fmt.Fprintln(os.Stderr, "pressie: --for is required")
		os.Exit(1)
	}

	idx, err := store.LoadIndex(cfg.GiftsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	_, name, relPath, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.For, cfg.Visibility)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	absPath := filepath.Join(cfg.GiftsDir, relPath)
	cf, err := store.LoadContactFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: no gift history found for %s\n", name)
		os.Exit(1)
	}

	direction := cfg.Direction
	if direction == "" {
		direction = "both"
	}

	year := cfg.Year

	// Collect and filter gifts.
	var given, received []store.Gift
	if direction == "given" || direction == "both" {
		given = filterGifts(cf.GiftsGiven, year)
	}
	if direction == "received" || direction == "both" {
		received = filterGifts(cf.GiftsReceived, year)
	}

	total := len(given) + len(received)
	if total == 0 {
		fmt.Printf("No gifts found for %s", name)
		if year != "" {
			fmt.Printf(" in %s", year)
		}
		fmt.Println(".")
		return
	}

	fmt.Printf("Gifts for %s", name)
	if year != "" {
		fmt.Printf(" (%s)", year)
	}
	fmt.Println(":")

	if len(given) > 0 {
		fmt.Printf("Given (%d):\n\n", len(given))
		for _, g := range given {
			printGift(g)
		}
	}

	if len(received) > 0 {
		if len(given) > 0 {
			fmt.Println()
		}
		fmt.Printf("Received (%d):\n\n", len(received))
		for _, g := range received {
			printGift(g)
		}
	}
}

// filterGifts returns gifts matching the given year (empty year = all).
func filterGifts(gifts []store.Gift, year string) []store.Gift {
	if year == "" {
		return gifts
	}
	filtered := make([]store.Gift, 0, len(gifts))
	for _, g := range gifts {
		if strings.HasPrefix(g.Date, year) {
			filtered = append(filtered, g)
		}
	}
	return filtered
}

// printGift prints a single gift in a human-readable format.
func printGift(g store.Gift) {
	fmt.Printf("  %s\n", g.Item)
	fmt.Printf("    id: %s\n", g.ID)
	if g.Occasion != "" {
		fmt.Printf("    occasion: %s\n", g.Occasion)
	}
	fmt.Printf("    date: %s\n", g.Date)
	if g.Price != nil {
		currency := g.Currency
		if currency == "" {
			currency = "USD"
		}
		fmt.Printf("    price: %s %.0f\n", currency, *g.Price)
	}
	if g.Notes != "" {
		fmt.Printf("    notes: %s\n", g.Notes)
	}
	fmt.Println()
}