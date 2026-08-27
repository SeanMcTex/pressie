package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		cmdInit(args)
	case "add-given":
		cmdAddGiven(args)
	case "add-received":
		cmdAddReceived(args)
	case "add-idea":
		cmdAddIdea(args)
	case "list":
		cmdList(args)
	case "ideas":
		cmdIdeas(args)
	case "assign":
		cmdAssign(args)
	case "search":
		cmdSearch(args)
	case "resolve":
		cmdResolve(args)
	case "stats":
		cmdStats(args)
	case "sync":
		cmdSync(args)
	case "plugins":
		cmdPlugins(args)
	case "version", "--version", "-v":
		fmt.Printf("pressie %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `pressie %s — agent-friendly gift tracking

Usage:
  pressie <command> [options]

Commands:
  init [path]              Initialize a gifts directory
  add-given                Log a gift you gave
  add-received             Log a gift you received
  add-idea                 Add a gift idea (for a person or "anyone")
  list                     List gifts for a contact
  ideas                    Show ideas for a contact (includes tag-matched general)
  assign <idea-id>         Assign a general idea to a specific person
  search <query>           Search contacts via plugin
  resolve <name>           Resolve a name to a contact key
  stats                    Show statistics (spending, counts)
  sync                     Git pull + push
  plugins                  List configured plugins
  version                  Show version
  help                     Show this help

Options:
  --for <name>             Target contact (by name or key)
  --to <name>              Recipient (for add-given)
  --from <name>            Giver (for add-received)
  --item <text>            Gift item description
  --occasion <text>        Occasion (christmas, birthday, etc.)
  --date <YYYY-MM-DD>      Date of the gift
  --url <url>              Link to product or inspiration
  --tags <a,b,c>           Comma-separated tags
  --price <number>         Price or price estimate
  --notes <text>           Free-text notes
  --private                Store in private directory (default)
  --shared                 Store in shared directory
  --status <status>        Filter by status (open, purchased, archived)
  --direction <dir>        Filter: given, received, or both
  --year <YYYY>            Filter by year (for stats)

Environment:
  PRESSIE_DIR              Path to gifts directory (default: ./gifts or ~/.pressie)

`, version)
}