# Pressie

**Your agent's gift memory.** An agent-friendly gift tracking CLI with a built-in web UI.

Pressie tracks gift ideas, gifts given, and gifts received — designed to be used directly by you, mediated by an AI agent, or through a web browser. Plain JSON storage, plugin-based contacts, no database, no lock-in.

## Why?

Every existing gift tracker is a walled-garden app. None are designed to work with AI agents, automation, or composable workflows. Pressie is different:

- **Agent-first:** The CLI is the primary interface. An agent can query, add, and manage gifts in one command — no GUI required. Run `pressie -h` for the full command reference.
- **Plain JSON:** Your data is diff-friendly, grep-able, and version-controllable. No database, no server required.
- **Web UI built in:** `pressie serve` starts a web server with a full UI — same data, same binary.
- **Plugin-based contacts:** Don't reinvent contact management. Plugins resolve names to contacts from Google Contacts, Apple Contacts, CSV, or anything else.
- **Thoughtful by design:** Tag-matched general ideas surface for the right people. The tool learns what you've already given and filters out repeats.
- **Open source:** MIT licensed. Your data, your tool, your server.

## Install

```bash
go install github.com/SeanMcTex/pressie/cmd/pressie@latest
```

The binary is placed in `$(go env GOPATH)/bin`. Ensure that's on your `$PATH`.

## Quick Start

```bash
# Initialize a gifts directory
pressie init ~/gifts

# Add a gift idea for someone
pressie add-idea --for "Kris" --item "Letterpress print" --tags art,irish

# Add a general idea (applicable to many people, matched by tags)
pressie add-idea --for anyone --item "Ceramic pour-over set" --tags kitchen,handmade

# See ideas for a person (includes tag-matched general ideas, filters past gifts)
pressie ideas --for "Kris"

# Log a gift you gave (retires matching ideas automatically)
pressie add-given --to "Kris" --item "Custom map" --occasion christmas --date 2025-12-25

# List gift history
pressie list --for "Kris" --direction both

# Set freeform preferences for a person
pressie prefs --for "Kris" --set "Favorite colors: blue, green. Shoe size: 11."

# Delete an idea by ID (find IDs with: pressie ideas --for "Kris")
pressie delete-idea --idea <id> --for "Kris"
```

### Web UI

```bash
# Start the web server (localhost only, no auth)
pressie serve

# With token auth (accessible from your network)
pressie serve --auth-token mysecret

# Custom port
pressie serve --port 8080 --auth-token mysecret
```

Open `http://localhost:7612` in your browser. The web UI shows your contacts list, per-person detail with ideas/gifts tabs, and general ideas — all reading and writing the same JSON files as the CLI.

### Agent Discovery

Pressie includes an `AGENTS.md` file at the repo root that tells AI agents how to use the CLI. Agents can discover all commands, flags, and workflow patterns by reading that file or running `pressie <command> -h`.

## How It Works

Pressie stores everything as JSON files in a directory you control:

```
gifts/
  _index.json              # config: plugins, contact mappings
  _ideas-general.json      # "for anyone" ideas (matched by tags)
  _private/
    kris-mcmains.json      # gifts for Kris — private to you
  _shared/
    chris-mcmains.json     # gifts for Chris — visible to collaborators
```

Each per-contact file contains gifts given, gifts received, ideas, and freeform preferences. General ideas live in a shared file and surface for relevant people via tag matching.

### Smart Idea Matching

When you run `pressie ideas --for "Kris"`:

1. Load Kris's direct ideas from their contact file.
2. Load general ideas whose tags intersect Kris's tags.
3. Filter out ideas that resemble gifts already given (duplicate avoidance).
4. Return the merged list, newest first.

### Contacts Plugins

Pressie doesn't manage contacts — it asks your plugin. A plugin is just an executable that takes a name query and returns JSON:

```bash
$ plugins/google-contacts.sh "Kris"
[{"key":"gc:12345","name":"Kris McMains","email":"kris@example.com"}]
```

Write plugins for any contact source. See [docs/plugin-spec.md](docs/plugin-spec.md) for the full specification.

### Commands

Run `pressie -h` for the full list. Each command also has per-command help: `pressie <command> -h`.

| Command | Description |
|---|---|
| `init` | Initialize a gifts directory |
| `add-idea` | Add a gift idea (for a person or "anyone") |
| `add-given` | Log a gift you gave (retires matching ideas) |
| `add-received` | Log a gift you received |
| `ideas` | Show ideas for a contact (filters past gifts) |
| `list` | List gifts given/received for a contact |
| `prefs` | Read or set freeform preferences for a contact |
| `delete-idea` | Delete an idea by ID |
| `serve` | Start the web server |
| `help` | Show all commands |

## Documentation

- [Brainstorming & Design Notes](docs/brainstorming.md)
- [Plugin Specification](docs/plugin-spec.md)
- [Data Model](docs/data-model.md)
- [Contributing](CONTRIBUTING.md)

## License

MIT — see [LICENSE](LICENSE)