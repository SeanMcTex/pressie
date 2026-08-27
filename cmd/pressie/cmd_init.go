package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/store"
)

// cmdInit initializes a new gifts directory with the default structure.
func cmdInit(args []string) {
	if wantsHelp(args) {
		helpInit()
		return
	}
	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	dir := cfg.GiftsDir
	if dir == "" {
		dir = config.DefaultGiftsDir()
	}

	// Refuse to clobber an existing installation.
	if _, err := os.Stat(filepath.Join(dir, "_index.json")); err == nil {
		fmt.Fprintf(os.Stderr, "pressie: already initialized at %s\n", dir)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Join(dir, "_private"), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Join(dir, "_shared"), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if err := store.SaveIndex(dir, &store.IndexFile{Version: 1, Plugins: []store.PluginConfig{}}); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if err := store.SaveGeneralIdeas(dir, []store.Idea{}); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Initialized pressie at %s\n", dir)
}