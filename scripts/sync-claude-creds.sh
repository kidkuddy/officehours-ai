#!/bin/sh
# Extract the Claude Code subscription token from the macOS keychain into a
# gitignored file that docker-compose mounts into the api container.
# Run this on the host BEFORE `docker compose up` (re-run if the token expires).
set -e
DIR="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "$DIR/.secrets"
security find-generic-password -s "Claude Code-credentials" -w > "$DIR/.secrets/claude-credentials.json"
chmod 600 "$DIR/.secrets/claude-credentials.json"
echo "Wrote $DIR/.secrets/claude-credentials.json"
