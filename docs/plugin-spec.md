# Pressie Plugin Specification

Plugins let Pressie resolve names to contact records without coupling to any specific contacts backend. A plugin is simply an executable that receives a query and returns JSON.

## Contract

### Input

A plugin receives a single argument: the query string. This may be a name ("Kris McMains"), an email ("kris@example.com"), or a unique key ("gc:12345").

### Output

A plugin must output a JSON array to stdout. Each element is a contact match:

```json
[
  {
    "key": "gc:12345",
    "name": "Kris McMains",
    "email": "kris@example.com",
    "birthday": "1985-03-15",
    "metadata": {}
  }
]
```

### Required Fields

| Field   | Type   | Required | Description |
|---------|--------|----------|-------------|
| `key`   | string | yes      | Unique identifier for this contact within this plugin's namespace. Must be stable across queries. |
| `name`  | string | yes      | Display name. |

### Optional Fields

| Field      | Type   | Description |
|------------|--------|-------------|
| `email`    | string | Primary email address. |
| `birthday` | string | ISO date (YYYY-MM-DD) if known. |
| `metadata` | object | Arbitrary key-value pairs (phone, address, tags, etc.). Pressie preserves but does not interpret these. |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Success (may return empty array if no matches) |
| 1    | Plugin error (network failure, auth issue, etc.) |
| 2    | Invalid arguments |

Pressie reads stdout only on exit code 0. On non-zero exit, Pressie reports the plugin error to the user.

## Configuration

Plugins are registered in `_index.json`:

```json
{
  "plugins": [
    {
      "name": "google-contacts",
      "command": "plugins/google-contacts.sh",
      "timeout_ms": 5000
    },
    {
      "name": "manual",
      "command": "plugins/manual.sh",
      "timeout_ms": 1000
    }
  ]
}
```

### Fields

| Field        | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `name`       | string | yes      | Unique plugin identifier. Used as the namespace prefix for contact keys. |
| `command`    | string | yes      | Path to the executable. Relative to the gifts directory root. |
| `timeout_ms` | int    | no       | Max execution time in milliseconds. Default: 10000. |

## Resolution Order

When resolving a name, Pressie queries plugins in the order listed in `_index.json`. The first plugin that returns a match wins. If multiple plugins return matches, Pressie presents all options to the user (or, in agent mode, returns all matches for the agent to disambiguate).

## Key Namespacing

Contact keys are namespaced by plugin name: `plugin-name:contact-key`. This ensures keys from different plugins never collide. For example, `google-contacts:12345` vs `manual:kris-mcmains`.

## Writing a Plugin

A plugin can be written in any language. The only requirements:

1. It must be executable.
2. It must accept a query string as its first argument.
3. It must output a JSON array to stdout on success.
4. It must exit 0 on success, non-zero on failure.

### Example: Shell Script (Google Contacts via gog)

```bash
#!/bin/bash
set -euo pipefail
QUERY="$1"
# Use gog to search contacts
RESULT=$(gog contacts search "$QUERY" --json 2>/dev/null || echo '[]')
echo "$RESULT"
```

### Example: Python Script (CSV-based)

```python
#!/usr/bin/env python3
import csv, json, sys, os

contacts_file = os.path.expanduser("~/.contacts.csv")
query = sys.argv[1].lower()

matches = []
with open(contacts_file) as f:
    for row in csv.DictReader(f):
        if query in row.get("name", "").lower() or query in row.get("email", "").lower():
            matches.append({
                "key": row["id"],
                "name": row["name"],
                "email": row.get("email", ""),
            })
print(json.dumps(matches))
```

### Example: Apple Contacts via AppleScript

```applescript
#!/usr/bin/env osascript
on run argv
    set query to item 1 of argv
    tell application "Contacts"
        set matches to every person whose name contains query
        set output to "["
        set firstItem to true
        repeat with p in matches
            if not firstItem then set output to output & ","
            set output to output & "{\"key\":\"" & id of p & "\",\"name\":\"" & name of p & "\"}"
            set firstItem to false
        end repeat
        set output to output & "]"
        return output
    end tell
end run
```