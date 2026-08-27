package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/gifts"
	"github.com/SeanMcTex/pressie/internal/store"
)

// cmdIdeas shows ideas for a contact, including tag-matched general ideas.
// Open ideas that resemble past gifts given are filtered out (duplicate avoidance),
// unless --status is set to show purchased/archived ideas.
func cmdIdeas(args []string) {
	if wantsHelp(args) {
		helpIdeas()
		return
	}
	cfg, err := config.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	if cfg.For == "" || cfg.For == "anyone" {
		listGeneralIdeas(cfg)
		return
	}

	listContactIdeas(cfg)
}

// listGeneralIdeas prints all general (unassigned) ideas, optionally filtered by status.
func listGeneralIdeas(cfg *config.Config) {
	ideas, err := store.LoadGeneralIdeas(cfg.GiftsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	ideas = filterByStatus(ideas, cfg.Status)

	if len(ideas) == 0 {
		fmt.Println("No general ideas yet.")
		fmt.Println("Add one with: pressie add-idea --for anyone --item \"...\"")
		return
	}

	sort.Slice(ideas, func(i, j int) bool {
		return ideas[i].Added > ideas[j].Added
	})

	fmt.Printf("General ideas (%d):\n\n", len(ideas))
	for _, idea := range ideas {
		printIdea(idea)
	}
}

// listContactIdeas loads a contact's direct ideas, merges tag-matched general
// ideas, filters out duplicates from gifts_given, and prints the combined list.
func listContactIdeas(cfg *config.Config) {
	idx := requireIndex(cfg.GiftsDir)

	_, name, relPath, err := gifts.ResolveContact(context.Background(), cfg.GiftsDir, idx, cfg.For, cfg.Visibility)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}

	absPath := filepath.Join(cfg.GiftsDir, relPath)
	cf, err := store.LoadContactFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressie: no ideas found for %s\n", name)
		os.Exit(1)
	}

	// Start with direct ideas.
	combined := make([]store.Idea, len(cf.Ideas))
	copy(combined, cf.Ideas)

	// Merge tag-matched general ideas (only open ones, unless --status overrides).
	statusFilter := cfg.Status
	if statusFilter == "" {
		statusFilter = "open"
	}

	if statusFilter == "open" {
		general, _ := store.LoadGeneralIdeas(cfg.GiftsDir)
		contactTags := cf.Tags
		if len(contactTags) > 0 {
			for _, gi := range general {
				if gi.Status != "open" {
					continue
				}
				if tagsIntersect(gi.Tags, contactTags) {
					combined = append(combined, gi)
				}
			}
		}
	}

	// Filter by status (default: open).
	combined = filterByStatus(combined, statusFilter)

	// Duplicate avoidance: filter out ideas that resemble past gifts given.
	// Only applies when showing open ideas (the default). If the user
	// explicitly requests --status purchased or archived, show everything.
	if cfg.Status == "" {
		combined = filterDuplicates(combined, cf.GiftsGiven)
	}

	if len(combined) == 0 {
		fmt.Printf("No ideas for %s yet.\n", name)
		fmt.Printf("Add one with: pressie add-idea --for \"%s\" --item \"...\"\n", name)
		return
	}

	// Sort newest first.
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Added > combined[j].Added
	})

	fmt.Printf("Ideas for %s (%d):\n\n", name, len(combined))
	for _, idea := range combined {
		printIdea(idea)
	}
}

// filterByStatus returns only ideas matching the given status.
// Empty status returns all ideas.
func filterByStatus(ideas []store.Idea, status string) []store.Idea {
	if status == "" {
		return ideas
	}
	filtered := make([]store.Idea, 0, len(ideas))
	for _, idea := range ideas {
		if idea.Status == status {
			filtered = append(filtered, idea)
		}
	}
	return filtered
}

// filterDuplicates removes ideas whose item text resembles any gift in giftsGiven.
// Uses the same normalizeItem + itemsMatch logic as add-given.
func filterDuplicates(ideas []store.Idea, giftsGiven []store.Gift) []store.Idea {
	if len(giftsGiven) == 0 {
		return ideas
	}
	filtered := make([]store.Idea, 0, len(ideas))
	for _, idea := range ideas {
		isDup := false
		for _, g := range giftsGiven {
			if itemsMatch(idea.Item, g.Item) {
				isDup = true
				break
			}
		}
		if !isDup {
			filtered = append(filtered, idea)
		}
	}
	return filtered
}

// printIdea prints a single idea in a human-readable format.
func printIdea(idea store.Idea) {
	fmt.Printf("  %s\n", idea.Item)
	fmt.Printf("    id: %s\n", idea.ID)
	if idea.URL != "" {
		fmt.Printf("    url: %s\n", idea.URL)
	}
	if idea.PriceEstimate != nil {
		currency := idea.Currency
		if currency == "" {
			currency = "USD"
		}
		fmt.Printf("    est: %s %.0f\n", currency, *idea.PriceEstimate)
	}
	if len(idea.Tags) > 0 {
		fmt.Printf("    tags: %s\n", strings.Join(idea.Tags, ", "))
	}
	if idea.Notes != "" {
		fmt.Printf("    notes: %s\n", idea.Notes)
	}
	fmt.Printf("    added: %s  status: %s\n", idea.Added, idea.Status)
	fmt.Println()
}

// tagsIntersect returns true if any tag in a appears in b (case-insensitive).
func tagsIntersect(a, b []string) bool {
	bset := make(map[string]bool, len(b))
	for _, t := range b {
		bset[strings.ToLower(t)] = true
	}
	for _, t := range a {
		if bset[strings.ToLower(t)] {
			return true
		}
	}
	return false
}