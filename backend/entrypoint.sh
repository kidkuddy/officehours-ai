#!/bin/sh
# Runs as root: stage the mounted (read-only) Claude subscription creds into the
# non-root user's writable home, fix ownership of writable dirs, then drop privileges.
# claude refuses --dangerously-skip-permissions under root, so the api (and the
# `claude` child it execs) must run as a non-root user.
set -e

mkdir -p /home/node/.claude
# Prefer the dedicated subscription token (extracted from the macOS keychain by
# scripts/sync-claude-creds.sh); fall back to the mounted ~/.claude dir.
if [ -f /claude-cred.json ]; then
  cp /claude-cred.json /home/node/.claude/.credentials.json
elif [ -d /host-claude ] && [ -f /host-claude/.credentials.json ]; then
  cp -n /host-claude/.credentials.json /home/node/.claude/ 2>/dev/null || true
fi
if [ -d /host-claude ] && [ -f /host-claude/settings.json ]; then
  cp -n /host-claude/settings.json /home/node/.claude/ 2>/dev/null || true
fi

# Gemini (Vertex AI) auth via Google ADC. The host's gcloud config is mounted
# read-only at /host-gcloud; stage it into the node user's writable home so the
# gemini CLI can read Application Default Credentials. Run on the host first:
#   gcloud auth application-default login
mkdir -p /home/node/.config/gcloud
if [ -d /host-gcloud ]; then
  cp -a /host-gcloud/. /home/node/.config/gcloud/ 2>/dev/null || true
fi
# Point the Google libraries / gemini CLI at the staged config + ADC file.
export CLOUDSDK_CONFIG=/home/node/.config/gcloud
if [ -f /home/node/.config/gcloud/application_default_credentials.json ]; then
  export GOOGLE_APPLICATION_CREDENTIALS=/home/node/.config/gcloud/application_default_credentials.json
fi

chown -R node:node /home/node /data 2>/dev/null || true

# Preserve the exported env (CLOUDSDK_CONFIG / GOOGLE_APPLICATION_CREDENTIALS)
# when dropping to the node user.
exec gosu node env \
  CLOUDSDK_CONFIG="${CLOUDSDK_CONFIG}" \
  GOOGLE_APPLICATION_CREDENTIALS="${GOOGLE_APPLICATION_CREDENTIALS}" \
  "$@"
