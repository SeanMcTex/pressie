# Version History

## 0.1.0 — initial release

First tagged release of pressie: gift tracking as plain JSON files.

- `init` — create a gifts directory
- `add-idea` / `delete-idea` — gift ideas for a person or general ("anyone")
- `add-given` / `add-received` — log gifts given and received; `add-given` retires matching open ideas
- `ideas` — list ideas with duplicate avoidance against past gifts; tag-matched general idea suggestions
- `list` — gifts given/received by contact, with direction/year filters
- `prefs` — freeform preferences per contact
- Contact resolution: exact slug → fuzzy → plugins → manual fallback
- Web interface: JSON API + SPA with auth (login form, Bearer token, cookie sessions)
- Plugin system for external contact sources

## Unreleased

### Added

- **Archive recipients** — `pressie archive-contact` / `unarchive-contact`, plus archive/unarchive endpoints and a confirm-dialog button in the web UI. Archived contacts are hidden from the contacts list and blocked from new entries; data files are untouched and `list`/`ideas` still work.
- **Edit existing entries** — `pressie edit-idea` and `pressie edit-gift` by ID (only passed flags change); web `PUT` endpoints for contact ideas, general ideas, and gifts, with a pre-filled edit dialog on every card.
- **`prefs --append`** — add a line to existing preferences without reading/rewriting them; `--set` still replaces.
- **`ideas --tag <tag>`** — case-insensitive tag filter for both per-contact and general idea listings.
- **Access key in URL** — opening the web UI with `?token=<key>` authenticates, sets the session cookie, and redirects to the clean path, so browser sessions don't need the login form.
- **Given-ideas filter in web** — the Ideas tab hides ideas already marked given by default (matching the CLI), with a "show them" toggle.

### Changed

- **"Purchased" is now "given"** — ideas marked as given use `status: "given"`. The store transparently normalizes the legacy `"purchased"` value on load, and `ideas --status purchased` remains accepted, so existing data files and scripts keep working. The web button reads "✓ Mark given".
- Contacts in the web list are sorted alphabetically (was map order).
- Embedded static assets send `Cache-Control: no-cache` so browser sessions pick up UI updates after a binary rebuild.