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

// cmdPrefs reads or sets freeform preferences for a contact.
// Without --set, prints the current preferences. With --set, replaces them.
func cmdPrefs(args []string) {
	if wantsHelp(args) {
		helpPrefs()
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

	idx := requireIndex(cfg.GiftsDir)

	key, name, relPath, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.For, cfg.Visibility)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	absPath := filepath.Join(cfg.GiftsDir, relPath)
	cf, err := gifts.EnsureContactFile(cfg.GiftsDir, relPath, key, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.HasPrefs {
		// Set mode: replace preferences.
		cf.Preferences = cfg.Prefs
		if err := store.SaveContactFile(absPath, cf); err != nil {
			fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
			os.Exit(1)
		}
		gifts.RegisterContact(idx, key, relPath, cfg.Visibility, nil)
		if err := store.SaveIndex(cfg.GiftsDir, idx); err != nil {
			fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Preferences set for %s.\n", name)
	} else {
		// Read mode: print current preferences.
		if cf.Preferences == "" {
			fmt.Printf("No preferences set for %s.\n", name)
			fmt.Printf("Set with: pressie prefs --for \"%s\" --set \"...\"\n", name)
			return
		}
		fmt.Printf("Preferences for %s:\n\n%s\n", name, cf.Preferences)
	}
}