package main

import (
	"fmt"
)

// wantsHelp returns true if args contains -h or --help.
func wantsHelp(args []string) bool {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}

// printHelp prints a per-command help block to stdout.
func printHelp(name, summary, usage, flags, example string) {
	fmt.Printf(`pressie %s — %s

Usage:
  pressie %s%s

Flags:
%s
`, name, summary, name, usage, flags)

	if example != "" {
		fmt.Printf("Example:\n  %s\n\n", example)
	} else {
		fmt.Println()
	}
}

// printHelpStub prints a not-yet-implemented help block.
func printHelpStub(name, summary, plannedFlags string) {
	fmt.Printf(`pressie %s — %s

Not yet implemented.

Planned flags:
%s
`, name, summary, plannedFlags)
	fmt.Println()
}

// --- Per-command help text ---

func helpInit() {
	printHelp("init",
		"Initialize a gifts directory",
		" [path]",
		`  [path]                  Target directory (default: ./gifts or ~/.pressie)
  --private                (ignored, init always creates both)
  --shared                 (ignored, init always creates both)`,
		`pressie init ~/gifts`,
	)
}

func helpAddIdea() {
	printHelp("add-idea",
		"Add a gift idea for a person or \"anyone\"",
		" --for <name> --item <text> [flags]",
		`  --for <name>             Contact name, key, or "anyone" for a general idea (required)
  --item <text>            Gift idea description (required)
  --tags <a,b,c>           Comma-separated tags (enables tag-matched suggestions)
  --url <url>              Link to product or inspiration
  --price <number>         Rough price estimate
  --notes <text>           Free-text notes
  --private                Store in private directory (default)
  --shared                 Store in shared directory (visible to collaborators)`,
		`pressie add-idea --for "Kris" --item "Letterpress print" --tags art,irish`,
	)
}

func helpAddGiven() {
	printHelp("add-given",
		"Log a gift you gave to someone",
		" --to <name> --item <text> [flags]",
		`  --to <name>              Recipient name or key (required)
  --item <text>            Gift description (required)
  --idea <id>              Retire a specific idea by ID (see: pressie ideas)
  --occasion <text>       Occasion: christmas, birthday, etc.
  --date <YYYY-MM-DD>      Date of the gift (default: today)
  --price <number>         Amount spent
  --notes <text>           Free-text notes
  --private                Store in private directory (default)
  --shared                 Store in shared directory`,
		`pressie add-given --to "Kris" --item "Custom map" --occasion christmas --date 2025-12-25`,
	)
}
func helpAddReceived() {
	printHelp("add-received",
		"Log a gift you received from someone",
		" --from <name> --item <text> [flags]",
		`  --from <name>            Giver name or key (required)
  --item <text>            Gift description (required)
  --occasion <text>       Occasion: christmas, birthday, etc.
  --date <YYYY-MM-DD>      Date received (default: today)
  --price <number>         Estimated value
  --notes <text>           Free-text notes
  --private                Store privately (default)
  --shared                 Store in shared directory`,
		`pressie add-received --from "Sam" --item "DADGAD poster" --occasion christmas --date 2025-12-25`,
	)
}

func helpIdeas() {
	printHelp("ideas",
		"Show gift ideas for a contact, including tag-matched general ideas",
		" --for <name> [flags]",
		`  --for <name>             Contact name, key, or "anyone" for all general ideas
  --status <status>        Filter: open (default), purchased, archived
                           Default filters out gifts already given (duplicate avoidance)
  --private                Look in private directory (default)
  --shared                 Look in shared directory`,
		`pressie ideas --for "Kris"`,
	)
}

func helpList() {
	printHelp("list",
		"List gifts given and/or received for a contact",
		" --for <name> [flags]",
		`  --for <name>             Contact name or key (required)
  --direction <dir>        Filter: given, received, or both (default: both)
  --year <YYYY>            Filter by year
  --private                Look in private directory (default)
  --shared                 Look in shared directory`,
		`pressie list --for "Kris" --direction both`,
	)
}

func helpAssign() {
	printHelpStub("assign",
		"Assign a general idea to a specific person",
		`  <idea-id>                ID of the general idea to assign
  --for <name>             Contact to assign the idea to
  --private                Store in private directory (default)
  --shared                 Store in shared directory`,
	)
}

func helpSearch() {
	printHelpStub("search",
		"Search contacts via configured plugins",
		`  <query>                  Name or partial name to search for`,
	)
}

func helpResolve() {
	printHelpStub("resolve",
		"Resolve a name to a contact key and file path",
		`  <name>                   Name to resolve
  --private                Look in private directory (default)
  --shared                 Look in shared directory`,
	)
}

func helpStats() {
	printHelpStub("stats",
		"Show gift statistics (spending, counts)",
		`  --year <YYYY>            Filter by year
  --direction <dir>        Filter: given, received, or both`,
	)
}

func helpSync() {
	printHelpStub("sync",
		"Sync the gifts directory with git remote (pull + push)",
		`  (no flags)`,
	)
}

func helpPlugins() {
	printHelpStub("plugins",
		"List configured contact plugins",
		`  (no flags)`,
	)
}

func helpPrefs() {
	printHelp("prefs",
		"Read or set freeform preferences for a contact",
		" --for <name> [--set <text>] [flags]",
		`  --for <name>             Contact name or key (required)
  --set <text>             Set preferences (replaces existing). Without this, reads current preferences.
  --private                Look in private directory (default)
  --shared                 Look in shared directory`,
		`pressie prefs --for "Kris" --set "Favorite colors: blue, green. Shoe size: 11."`,
	)
}