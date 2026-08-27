# Pressie

**Your agent's gift memory.** An agent-friendly gift tracking CLI.

Pressie tracks gift ideas, gifts given, and gifts received — designed to be used directly by you or mediated by an AI agent. Plain JSON storage, plugin-based contacts, git sync for multiplayer.

## Why?

Every existing gift tracker is a walled-garden app. None are designed to work with AI agents, automation, or composable workflows. Pressie is different:

- **Agent-first:** The CLI is the primary interface. An agent can query, add, and manage gifts in one command — no GUI required.
- **Plain JSON:** Your data is diff-friendly, grep-able, and version-controllable. No database, no lock-in.
- **Plugin-based contacts:** Don't reinvent contact management. Plugins resolve names to contacts from Google Contacts, Apple Contacts, CSV, or anything else.
- **Multiplayer via git:** Private/shared partition means you and your partner can collaborate without spoiling surprises.
- **Open source:** MIT licensed. Your data, your tool, your server.

## Install

```bash
brew install pressie
```

*Or build from source:*

```bash
go install github.com/SeanMcTex/pressie/cmd/pressie@latest
```

## Quick Start

```bash
# Initialize a gifts directory
pressie init ~/gifts

# Add a gift idea for someone
pressie add-idea --for "Kris" --item "Letterpress print" --tags art,irish

# Add a general idea (not assigned to anyone yet)
pressie add-idea --for anyone --item "Ceramic pour-over set" --tags kitchen,handmade --shared

# Log a gift you gave
pressie add-given --to "Kris" --item "Custom map" --occasion christmas --date 2025-12-25

# See ideas for a person (includes tag-matched general ideas)
pressie ideas --for "Kris"

# List gift history
pressie list --for "Kris" --direction both

# Sync with remote (git pull + push)
pressie sync
```

## How It Works

Pressie stores everything as JSON files in a directory you control:

```
gifts/
  _index.json              # config: plugins, contact mappings
  _ideas-general.json      # "for anyone" ideas
  _private/
    kris-mcmains.json      # gifts for Kris — private to you
  _shared/
    chris-mcmains.json     # gifts for Chris — visible to partner
```

Each per-contact file contains gifts given, gifts received, and ideas. General ideas live in a shared file and surface for relevant people via tag matching.

### Contacts Plugins

Pressie doesn't manage contacts — it asks your plugin. A plugin is just an executable that takes a name query and returns JSON:

```bash
$ plugins/google-contacts.sh "Kris"
[{"key":"gc:12345","name":"Kris McMains","email":"kris@example.com"}]
```

Write plugins for any contact source. See [docs/plugin-spec.md](docs/plugin-spec.md) for the full specification.

### Multiplayer

Use git for sync. `pressie sync` does a pull + push. Private files stay private; shared files are visible to all collaborators. See [docs/data-model.md](docs/data-model.md) for the visibility model.

## Documentation

- [Brainstorming & Design Notes](docs/brainstorming.md)
- [Plugin Specification](docs/plugin-spec.md)
- [Data Model](docs/data-model.md)

## License

MIT — see [LICENSE](LICENSE)