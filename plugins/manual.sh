#!/bin/bash
# Manual contacts plugin for Pressie
# No external dependencies. Uses slugified name as the key.
# Useful as a fallback when no contacts backend is configured.
set -euo pipefail

QUERY="$1"

# Simple slugify: lowercase, replace spaces with hyphens, strip non-alphanumeric
SLUG=$(echo "$QUERY" | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g' | sed 's/[^a-z0-9-]//g')

# Return a single match with the slug as key
echo "[{\"key\":\"$SLUG\",\"name\":\"$QUERY\"}]"