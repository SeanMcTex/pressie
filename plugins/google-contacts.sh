#!/bin/bash
# Google Contacts plugin for Pressie
# Wraps `gog` CLI to search Google Contacts.
# Requires gog to be installed and authenticated.
set -euo pipefail

QUERY="$1"

# gog contacts search returns JSON; adapt to Pressie plugin format.
# Adjust this script based on gog's actual output format.
RESULT=$(gog contacts search "$QUERY" --json 2>/dev/null || echo '[]')

# Pass through — gog already returns JSON array of contacts.
# If gog's format differs from Pressie's expected schema, transform here.
echo "$RESULT"