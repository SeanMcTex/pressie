package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SeanMcTex/pressie/internal/store"
)

// requireIndex loads _index.json from the gifts directory, exiting with a
// helpful message if the directory hasn't been initialized yet.
func requireIndex(giftsDir string) *store.IndexFile {
	idx, err := store.LoadIndex(giftsDir)
	if err != nil {
		if _, statErr := os.Stat(filepath.Join(giftsDir, "_index.json")); os.IsNotExist(statErr) {
			fmt.Fprintf(os.Stderr, "pressie: no gifts directory found at %s\n", giftsDir)
			fmt.Fprintf(os.Stderr, "Run `pressie init %s` first.\n", giftsDir)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}
	return idx
}

// rejectArchived exits if the given contact key maps to an archived contact.
// Call after ResolveContact for commands that create or list active data.
func rejectArchived(idx *store.IndexFile, key, name string) {
	if m, ok := idx.Contacts[key]; ok && m.Archived {
		fmt.Fprintf(os.Stderr, "pressie: %s is archived (use `pressie unarchive-contact --for %s` to restore)\n", name, key)
		os.Exit(1)
	}
}