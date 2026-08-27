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
	case "delete-idea":
		cmdDeleteIdea(args)
	case "list":
		cmdList(args)
	case "ideas":
		cmdIdeas(args)
	case "prefs":
		cmdPrefs(args)
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
	case "serve":
		cmdServe(args)
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
	fmt.Fprintf(os.Stderr, `pressie %s — your agent's gift memory

Usage:
  pressie <command> [options]
  pressie <command> -h        Per-command help

Commands (implemented):
  init [path]              Initialize a gifts directory
  add-idea                 Add a gift idea (for a person or "anyone")
  delete-idea              Delete an idea by ID
  add-given                Log a gift you gave (retires matching ideas)
  add-received             Log a gift you received
  ideas                    Show ideas for a contact (filters past gifts)
  list                     List gifts given/received for a contact
  prefs                    Read or set freeform preferences for a contact
  serve                    Start the web server

Commands (not yet implemented):
  assign <idea-id>         Assign a general idea to a specific person
  search <query>           Search contacts via plugin
  resolve <name>           Resolve a name to a contact key
  stats                    Show statistics (spending, counts)
  sync                     Git pull + push
  plugins                  List configured plugins

Quick start:
  pressie init ~/gifts
  pressie add-idea --for "Kris" --item "Letterpress print" --tags art,irish
  pressie ideas --for "Kris"
  pressie add-given --to "Kris" --item "Letterpress print" --occasion christmas
  pressie ideas --for "Kris"          # "Letterpress print" now filtered out

Required flags by command:
  init                    [path]
  add-idea                --for <name> --item <text>
  delete-idea             --idea <id>  (add --for <name> for per-contact)
  add-given               --to <name> --item <text>
  add-received            --from <name> --item <text>
  ideas                   --for <name>  (or "anyone" for general ideas)
  list                    --for <name>
  prefs                   --for <name>  (add --set <text> to write)

Common flags:
  --for <name>             Target contact (by name or key)
  --to <name>              Recipient (for add-given)
  --from <name>            Giver (for add-received)
  --item <text>            Gift item description
  --occasion <text>        Occasion (christmas, birthday, etc.)
  --date <YYYY-MM-DD>      Date of the gift (default: today)
  --url <url>              Link to product or inspiration
  --tags <a,b,c>           Comma-separated tags
  --price <number>         Price or price estimate
  --notes <text>           Free-text notes
  --private                Store in private directory (default)
  --shared                 Store in shared directory (visible to collaborators)
  --status <status>        Filter: open, purchased, archived
  --direction <dir>        Filter: given, received, or both
  --year <YYYY>            Filter by year

Environment:
  PRESSIE_DIR              Path to gifts directory (default: ./gifts or ~/.pressie)

Run `+"`pressie <command> -h`"+` for detailed help on any command.

`, version)
}