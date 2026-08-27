# Pressie Data Model

## Directory Structure

```
gifts/
  _index.json              # config: plugins, sync settings
  _ideas-general.json      # unassigned "for anyone" ideas (shared)
  _private/
    <contact-slug>.json    # gifts/ideas for a contact — private to you
  _shared/
    <contact-slug>.json    # gifts/ideas for a contact — visible to all
    ideas-general.json     # shared general ideas (if not using top-level)
  images/                  # image storage (v1.3+)
    <id>.jpg
```

## _index.json

```json
{
  "version": 1,
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
  ],
  "sync": {
    "type": "git",
    "remote": "git@github.com:SeanMcTex/pressie-data.git",
    "branch": "main"
  },
  "contacts": {
    "google-contacts:12345": {
      "file": "_private/kris-mcmains.json",
      "visibility": "private",
      "tags": ["art", "irish", "kitchen", "garden"]
    },
    "manual:blair-raker": {
      "file": "_shared/blair-raker.json",
      "visibility": "shared",
      "tags": ["music", "irish", "teaching"]
    }
  }
}
```

## Per-Contact File

```json
{
  "contact_key": "google-contacts:12345",
  "name": "Kris McMains",
  "tags": ["art", "irish", "kitchen", "garden"],
  "gifts_given": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "date": "2025-12-25",
      "occasion": "christmas",
      "item": "Custom map of our first date location",
      "price": 85,
      "currency": "USD",
      "notes": "Framed, she loved it",
      "images": [],
      "source": "manual",
      "added": "2025-12-25T10:30:00Z"
    }
  ],
  "gifts_received": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "date": "2025-12-25",
      "occasion": "christmas",
      "item": "DADGAD chord poster",
      "price": null,
      "currency": null,
      "notes": "Homemade, hung in music room",
      "images": [],
      "source": "manual",
      "added": "2025-12-25T10:35:00Z"
    }
  ],
  "ideas": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "item": "Letterpress print of an Irish poem",
      "url": "https://example.com/print",
      "price_estimate": 40,
      "currency": "USD",
      "tags": ["art", "irish"],
      "status": "open",
      "added": "2026-08-20",
      "notes": "Saw at the Guinness spot, she mentioned liking the style",
      "images": []
    }
  ]
}
```

## General Ideas File

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440003",
    "item": "Handmade ceramic pour-over coffee set",
    "url": "https://example.com/ceramic",
    "price_estimate": 60,
    "currency": "USD",
    "tags": ["kitchen", "handmade"],
    "status": "open",
    "added": "2026-08-27",
    "notes": "Saw at the market, universally appealing",
    "assigned_to": null,
    "images": []
  }
]
```

## Field Reference

### Gift (given or received)

| Field      | Type   | Required | Description |
|------------|--------|----------|-------------|
| `id`       | string | yes      | UUID v4. |
| `date`     | string | yes      | ISO date (YYYY-MM-DD). |
| `occasion` | string | yes      | Free text: "christmas", "birthday", "just because", etc. |
| `item`     | string | yes      | Description of the gift. |
| `price`    | number | no       | Amount spent. null if unknown or not tracked. |
| `currency` | string | no       | ISO 4217 code. null if no price. |
| `notes`    | string | no       | Free-text notes. |
| `images`   | array  | no       | Array of relative paths to image files. Empty array if none. |
| `source`   | string | no       | How this entry was added: "manual", "agent", "import". Default: "manual". |
| `added`    | string | no       | ISO timestamp when the record was created. |
| `thanked`  | string | no       | Thank-you status for received gifts: "pending" (default for received), "sent", or omitted. null/omitted for gifts given (not applicable). |

### Idea

| Field            | Type   | Required | Description |
|------------------|--------|----------|-------------|
| `id`             | string | yes      | UUID v4. |
| `item`           | string | yes      | Description of the gift idea. |
| `url`            | string | no       | Link to the product or inspiration. |
| `price_estimate` | number | no       | Rough cost estimate. |
| `currency`       | string | no       | ISO 4217 code. |
| `tags`           | array  | no       | Array of tag strings for matching. |
| `status`         | string | yes      | One of: "open", "purchased", "archived". Default: "open". |
| `added`          | string | yes      | ISO date when the idea was recorded. |
| `notes`          | string | no       | Free-text notes. |
| `images`         | array  | no       | Array of relative paths to image files. |
| `assigned_to`    | string | no       | Contact key if assigned to a specific person. null for general ideas. |

## Tag-Based Idea Matching

When querying `pressie ideas --for "Kris"`:

1. Load Kris's contact file → get her `tags` array.
2. Load all her direct ideas (from her contact file).
3. Load `_ideas-general.json` → filter where any tag in the idea's `tags` array intersects Kris's `tags` array.
4. Filter out any idea whose `item` text matches or closely resembles a gift in Kris's `gifts_given` (duplicate avoidance — see below).
5. Merge and return, sorted by `added` date (newest first).

If the contact has no `tags` defined, return only their direct ideas (no general matching).

## Purchased Tracking & Duplicate Avoidance

### Purchased Tracking

When a gift is logged via `add-given`, it's appended to `gifts_given` in the contact's file. If the gift item matches an existing open idea for that contact, the idea's `status` is set to `purchased` so it no longer surfaces in `ideas` queries.

### Duplicate Avoidance

The `ideas` command filters out suggestions that resemble past gifts given to the same person. This prevents repeating a gift you've already given.

Matching logic:
- Normalize item text: lowercase, trim, collapse whitespace.
- Exact match: if a gift's `item` equals an idea's `item`, suppress the idea.
- Substring match: if one normalized item is a substring of the other, suppress the idea.

This is intentionally fuzzy — it errs on the side of hiding potential repeats. The user can still see purchased/archived ideas via `pressie ideas --status purchased` if they want the full history.

## Visibility Rules

- **Private** files (`_private/`): only visible on the local machine. Not synced to shared remotes (or synced to a private remote only).
- **Shared** files (`_shared/`): synced to the shared remote, visible to all collaborators.
- The `_index.json` `contacts` section records the visibility of each contact. New contacts default to private.
- `_ideas-general.json` at the top level is shared by default. A `_private/ideas-general.json` can exist for private general ideas.