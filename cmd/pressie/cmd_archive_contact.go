package main

import (
	"context"
	"fmt"
	"os"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/gifts"
	"github.com/SeanMcTex/pressie/internal/store"
)

// cmdArchiveContact sets archived=true on a contact's index mapping.
// Archived contacts are hidden from the web UI and excluded from
// resolution for new entries, but their data stays on disk.
func cmdArchiveContact(args []string) {
	setArchived(args, true)
}

// cmdUnarchiveContact clears the archived flag on a contact.
func cmdUnarchiveContact(args []string) {
	setArchived(args, false)
}

func setArchived(args []string, archived bool) {
	if wantsHelp(args) {
		printArchiveHelp(archived)
		return
	}

	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.For == "" || cfg.For == "anyone" {
		fmt.Fprintln(os.Stderr, "pressie: --for <name> is required")
		os.Exit(1)
	}

	idx := requireIndex(cfg.GiftsDir)

	key, name, _, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.For, cfg.Visibility)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	m, ok := idx.Contacts[key]
	if !ok {
		fmt.Fprintf(os.Stderr, "pressie: contact %s is not in the index\n", name)
		os.Exit(1)
	}

	if m.Archived == archived {
		state := "already archived"
		if !archived {
			state = "not archived"
		}
		fmt.Printf("%s is %s.\n", name, state)
		return
	}

	m.Archived = archived
	idx.Contacts[key] = m

	if err := store.SaveIndex(cfg.GiftsDir, idx); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if archived {
		fmt.Printf("Archived %s.\n", name)
	} else {
		fmt.Printf("Unarchived %s.\n", name)
	}
}

func printArchiveHelp(archived bool) {
	verb := "Archive"
	if !archived {
		verb = "Unarchive"
	}
	fmt.Printf(`pressie %s-contact — %s a recipient

Usage:
  pressie %s-contact --for <name>

%s a contact's index mapping. Archived contacts are hidden from the
web interface and cannot be resolved for new entries; their data files
are kept. Use `+"`unarchive-contact`"+` to restore.

Options:
  --for <name>    Contact to %s (required)

Examples:
  pressie %s-contact --for "Kris"
`, verb, verb, verb, verb, verb, verb)
}