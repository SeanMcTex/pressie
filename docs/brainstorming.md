# Pressie — Brainstorming Notes

*August 27, 2026 — Sharky & Lorcan*

## Problem

Tracking gift ideas, what you've bought for people, and what they've bought for you is frustrating. Existing options:

- **Dedicated gift apps** (GiftList, GiftPlanner, Who Gave Me What, Gift Tracking) — walled gardens, most focus on the giver side, weak or no contacts integration, not agent-friendly.
- **Personal CRMs** (Monica, Dex, YourPond) — Monica doesn't actually have gift tracking (verified). Others claim it but unverified. All are SaaS or self-hosted web apps, none are agent-first.
- **Apple Notes + Shortcuts** — works but unstructured, no query power, no plugin ecosystem.

**The gap:** No gift tracker is designed to be agent-mediated, composable, or open. They're all apps. We want a tool.

## Vision

A CLI tool, agent-friendly first, human-friendly second. Plain JSON storage. Plugin system for contacts. Open source from day one.

**Agent-first** means: the primary interface is CLI/JSON. An agent (OpenClaw, Claude, whatever) can query and mutate gift data in one command. A human *can* use it directly, but the sweet spot is agent-mediated — "add a gift idea for Blair" from iMessage, and the agent runs the command.

## Design Decisions

### Storage

- **Plain JSON files, no database.** Diff-friendly, grep-able, version-controllable.
- **One file per contact.** Agents can read just what they need. Clean git diffs.
- **Private/shared partition.** Solves the "Kris can see her own gifts" problem.

```
gifts/
  _index.json              # config: plugins, sync settings, contact→file mapping
  _ideas-general.json      # unassigned "for anyone" ideas
  _private/
    kris-mcmains.json      # gifts for Kris — Sean only
  _shared/
    chris-mcmains.json     # gifts for Team Hollywood — both can see/edit
    ideas-general.json     # shared general ideas
```

### Plugin System

Plugins are executable scripts that resolve names to contact records. The tool never touches contacts directly.

**Contract:**
- Input: query string (name, email, or unique key)
- Output: JSON array of matches
- Exit code: 0 on success, non-zero on error

```json
[{ "key": "gc:12345", "name": "Kris McMains", "email": "kris@...", "birthday": "..." }]
```

Ship with: Google Contacts plugin (wrapping `gog`), manual plugin (slugified name as key).

### Data Model

Per-contact file:
```json
{
  "contact_key": "google-contacts:12345",
  "name": "Kris McMains",
  "tags": ["art", "irish", "kitchen", "garden"],
  "gifts_given": [
    {
      "id": "uuid",
      "date": "2025-12-25",
      "occasion": "christmas",
      "item": "Custom map of our first date location",
      "price": 85,
      "notes": "Framed, she loved it",
      "source": "manual"
    }
  ],
  "gifts_received": [
    {
      "id": "uuid",
      "date": "2025-12-25",
      "occasion": "christmas",
      "item": "DADGAD chord poster",
      "notes": "Homemade, hung in music room",
      "source": "manual"
    }
  ],
  "ideas": [
    {
      "id": "uuid",
      "item": "Letterpress print of an Irish poem",
      "url": "https://...",
      "price_estimate": 40,
      "tags": ["art", "irish"],
      "status": "open",
      "added": "2026-08-20",
      "notes": "Saw at the Guinness spot, she mentioned liking the style"
    }
  ]
}
```

General ideas file (`_ideas-general.json`):
```json
[
  {
    "id": "uuid",
    "item": "Handmade ceramic pour-over coffee set",
    "url": "https://...",
    "price_estimate": 60,
    "tags": ["kitchen", "handmade"],
    "status": "open",
    "added": "2026-08-27",
    "notes": "Saw at the market, universally appealing",
    "assigned_to": null
  }
]
```

### Tag-Based Idea Matching

When querying ideas for a person, return:
1. Their direct per-contact ideas
2. General ideas whose tags intersect the contact's tags

This lets you capture "saw this, good for someone" without knowing who yet.

### CLI Commands

```
pressie add-given --to "Kris" --item "..." --occasion christmas --date 2025-12-25 [--private|--shared]
pressie add-received --from "Blair" --item "..." --occasion birthday [--private|--shared]
pressie add-idea --for "Kris" --item "..." --url "..." --tags art,irish [--private|--shared]
pressie add-idea --for "anyone" --item "..." --tags kitchen,handmade --shared
pressie list --for "Kris" --direction given|received|both
pressie ideas --for "Kris" [--status open|archived|purchased]
pressie assign <idea-id> --to "Kris" [--private]
pressie search "Kris"              # fuzzy contact lookup via plugin
pressie stats --year 2025          # total spent per person, etc.
pressie resolve "Kris"             # resolve name → contact_key via plugin
pressie sync                       # git pull + push
pressie plugins list               # show configured plugins
pressie init [path]                # initialize a gifts directory
```

### Image Support (v1.3+)

Store `images` array of relative paths in idea/gift objects. CLI copies images to `images/` directory. Not essential for v1 but data model accommodates it.

```json
{
  "images": ["images/uuid-1.jpg"]
}
```

### Multiplayer

**Problem:** Kris doesn't have a personal agent. Needs a way to view and add gifts without touching a CLI.

**Solution: Surfaces on top of the same data.**

1. **CLI (v1)** — for agent users and terminal humans.
2. **Web UI + thin API server (v1.2)** — runs on home server. Reads/writes same JSON files. Caddy reverse proxy. Kris gets a URL. Simple pages: ideas browser, add idea, per-person view, add gift.
3. **iOS Shortcut (v1.3)** — hits the web API. "Add gift idea" from Siri or share sheet. Lowest-friction capture for non-technical users.
4. **iMessage bot (optional, if Kris gets an agent)** — CLI just works.

**Architecture:**
```
                    JSON files (source of truth)
                         |
              -----------+-----------
              |                     |
         CLI tool               Web API server
         (for agents/           (thin, reads/writes
          terminal users)        same JSON files)
              |                     |
         OpenClaw              Web UI (Kris)
         (you)                 iOS Shortcuts (Kris)
              
         Git sync ←→ Git remote ←→ Kris's machine
```

### Sync

Git-based. `pressie sync` = `git pull && git push`. Private repo. Both people work against the same repo. Private/shared partition handles access control at the file level.

Cloud sync (iCloud Drive, Syncthing) as an alternative for non-git users.

## Language: Go

- Single binary, no runtime dependency
- Cross-compile trivially
- `brew install` distribution
- Plugin system (executable scripts) is language-agnostic anyway
- Good HTTP server ecosystem for the web UI later

## Name: Pressie

Checked: no conflicts on GitHub (a few zero-star repos), npm, Homebrew, Go pkg.dev, crates.io, Docker Hub. Clear.

Irish/British slang for "present" — fits the vibe.

## Open Source

- License: MIT
- Public from day one
- Plugin spec properly documented
- Community contributions welcome but not required for v1

## v1 Scope

1. Core data model with private/shared partition
2. `add-given`, `add-received`, `add-idea`, `list`, `ideas`, `assign`
3. Google Contacts plugin + manual plugin
4. General ideas list with tag-based matching
5. `sync` via git
6. `_index.json` config
7. `init` command

## Post-v1

- v1.2: Web UI + API server
- v1.3: Image support, iOS Shortcuts
- v1.5: `stats`, `suggest` (agent-bait command)
- v2: Tauri app? Plugin marketplace?

## Repo Structure

```
pressie/
  README.md
  LICENSE
  go.mod
  go.sum
  cmd/pressie/           # main CLI entry
  internal/
    store/               # JSON file read/write
    contacts/            # plugin interface + built-in plugins
    gifts/               # gift/idea CRUD logic
    sync/                # git sync logic
    config/              # _index.json handling
  plugins/
    google-contacts.sh
    manual.sh
    apple-contacts.scpt
  web/                   # web UI + API server (v1.2)
  docs/
    brainstorming.md     # this file
    plugin-spec.md       # how to write a contacts plugin
    data-model.md
  examples/
    sample-gifts-dir/
      _index.json
      _ideas-general.json
      _private/
        example.json
```