# Agent Instructions

This file guides AI agents using Pressie as a gift-tracking tool.

## What is Pressie?

Pressie is a CLI for tracking gift ideas, gifts given, and gifts received. Data is stored as plain JSON files in a directory you control. No database, no server, no lock-in.

## Setup

```bash
pressie init ~/gifts        # create the gifts directory (do this once)
```

This creates `_index.json`, `_ideas-general.json`, `_private/`, and `_shared/`.

Set `PRESSIE_DIR` to point pressie at your gifts directory, or run from a directory containing a `gifts/` folder:

```bash
export PRESSIE_DIR=~/gifts
```

## Commands (implemented)

### init
```bash
pressie init [path]
```
Creates a gifts directory. Refuses to clobber an existing one.

### add-idea
```bash
pressie add-idea --for "Kris" --item "Letterpress print" --tags art,irish
pressie add-idea --for anyone --item "Ceramic pour-over set" --tags kitchen,handmade --shared
```
Required: `--for` (name or `"anyone"`), `--item`.
Optional: `--tags`, `--url`, `--price`, `--notes`, `--private` (default), `--shared`.

Tags on per-contact ideas also merge into the contact's tag profile, which drives tag-matched general idea suggestions.

### add-given
```bash
pressie add-given --to "Kris" --item "Custom map" --occasion christmas --date 2025-12-25 --price 85
pressie add-given --to "Kris" --item "Letterpress print" --idea <idea-id>
```
Required: `--to`, `--item`.
Optional: `--idea <id>` (retire a specific idea by ID — see `pressie ideas`), `--occasion`, `--date` (default: today), `--price`, `--notes`, `--private`/`--shared`.

Without `--idea`, pressie fuzzy-matches the gift item against open ideas and retires matches (prevents duplicate suggestions). With `--idea`, it retires exactly that idea by ID.

### add-received
```bash
pressie add-received --from "Sam" --item "DADGAD poster" --occasion christmas --date 2025-12-25
```
Required: `--from` (the giver), `--item`.
Optional: `--occasion`, `--date`, `--price`, `--notes`, `--private`/`--shared`.

### ideas
```bash
pressie ideas --for "Kris"
pressie ideas --for anyone
pressie ideas --for "Kris" --status purchased
```
Required: `--for` (name or `"anyone"`).
Optional: `--status` (open/purchased/archived, default: open), `--private`/`--shared`.

Output includes idea IDs (`id:` line) which can be passed to `add-given --idea <id>`.

Default behavior filters out ideas that resemble gifts already given to that person (duplicate avoidance). Use `--status purchased` to see retired ideas.

### list
```bash
pressie list --for "Kris"
pressie list --for "Kris" --direction given --year 2025
```
Required: `--for`.
Optional: `--direction` (given/received/both, default: both), `--year`, `--private`/`--shared`.

## Commands (not yet implemented)

`assign`, `search`, `resolve`, `stats`, `sync`, `plugins` — run `pressie <command> -h` for planned flags.

## How idea matching works

When you run `ideas --for "Kris"`:

1. Load Kris's direct ideas from their contact file.
2. Load general ideas (`--for anyone`) whose tags intersect Kris's tags.
3. Filter out ideas that resemble gifts already given (fuzzy text match).
4. Return the merged list, newest first.

This means: add ideas with tags that describe the person's interests (`art`, `irish`, `music`), and general ideas with matching tags will surface automatically.

## Contact resolution

Pressie resolves names in this order:
1. Exact slug match in the index (case-insensitive).
2. Fuzzy match: stored name contains the query, or vice versa.
3. Configured contact plugins (if any).
4. Manual fallback: slugify the name, create a new contact.

If a name is ambiguous (matches multiple contacts), pressie errors and lists the matches. Use the full name to disambiguate.

## Data model

All data is JSON files:
- `_index.json` — config, plugin list, contact-to-file mappings.
- `_ideas-general.json` — general (unassigned) ideas.
- `_private/<slug>.json` — per-contact gifts and ideas (private to you).
- `_shared/<slug>.json` — per-contact gifts and ideas (visible to collaborators).

## Agent workflow examples

**Track a gift idea and later give it:**
```bash
pressie add-idea --for "Kris" --item "Letterpress print" --tags art,irish
# ... time passes ...
pressie ideas --for "Kris"                    # see the idea + its ID
pressie add-given --to "Kris" --item "Letterpress print" --idea <id>  # retire it precisely
pressie ideas --for "Kris"                    # idea no longer shows
```

**Suggest a general idea for multiple people:**
```bash
pressie add-idea --for anyone --item "Irish wool scarf" --tags irish,warm
# Surfaces for anyone whose contact tags include "irish" or "warm"
pressie ideas --for "Kris"    # Kris has tag "irish" → scarf appears
```

**Check what you've given someone:**
```bash
pressie list --for "Kris" --direction given --year 2025
```

## Help

```bash
pressie -h                    # top-level help
pressie <command> -h          # per-command help with flags and examples
```