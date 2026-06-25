#!/usr/bin/env bash
#
# rag.sh — end-to-end test of the RAG retrieval pipeline against the running stack.
#
# Exercises `ohctl rag` inside the api container (which talks directly to Postgres):
#   1. Indexes a tmp folder with a known markdown file into a fresh collection
#      (rag-test) and asserts a query for a unique term ranks that file first.
#   2. Queries the already-seeded real collections (kb-product, kb-fundraising,
#      kb-tunisia) and asserts each returns >=1 grounded chunk with content + source.
#
# Requires: docker compose stack up (api + db), `jq` on the host.
# Run from anywhere; resolves the repo root from this script's location.
set -u

# --- locate repo root (dir containing docker-compose.yml) ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT_DIR" || { echo "FATAL: cannot cd to repo root $ROOT_DIR"; exit 2; }

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS+1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL+1)); }

# dce <args...> : run ohctl in the api container, stdin disabled (-T), no TTY.
dce() { docker compose exec -T api "$@"; }

# --- preflight ---
command -v jq >/dev/null 2>&1 || { echo "FATAL: jq not found on host"; exit 2; }
if ! docker compose ps --status running api 2>/dev/null | grep -q api; then
  echo "FATAL: api container is not running (docker compose up -d)"; exit 2
fi
dce ohctl rag --help >/dev/null 2>&1 || { echo "FATAL: ohctl rag not available in api container"; exit 2; }

echo "=== RAG retrieval E2E ==="

# ---------------------------------------------------------------------------
# Case 1: fresh-collection index + retrieval (round-trip)
# ---------------------------------------------------------------------------
# Use a unique nonce so the term cannot collide with any seeded content, and a
# unique collection name so reruns never read stale chunks.
NONCE="zylophonquark$$$(date +%s)"
TMPDIR_C="/tmp/rag-test-$$"
COLL="rag-test"
KNOWN_FILE="known.md"

# Build the known markdown file *inside* the container (the api process reads
# the folder from its own filesystem). /tmp is writable in the container.
dce sh -c "mkdir -p '$TMPDIR_C' && cat > '$TMPDIR_C/$KNOWN_FILE' <<EOF
# Known Test Document

This is a deliberately unique sentinel marker: $NONCE

The $NONCE token appears only in this file and nowhere else in the corpus,
so a full-text query for it must rank this document first.
EOF" || { fail "case1-setup: could not write tmp markdown in container"; }

INDEX_JSON="$(dce ohctl rag index "$TMPDIR_C" --collection "$COLL" 2>/dev/null)"
if echo "$INDEX_JSON" | jq -e '.' >/dev/null 2>&1; then
  FILES="$(echo "$INDEX_JSON" | jq -r '.files // 0')"
  CHUNKS="$(echo "$INDEX_JSON" | jq -r '.chunks // 0')"
  if [ "$FILES" -ge 1 ] && [ "$CHUNKS" -ge 1 ]; then
    pass "case1-index: indexed files=$FILES chunks=$CHUNKS into collection '$COLL'"
  else
    fail "case1-index: expected files>=1 & chunks>=1, got files=$FILES chunks=$CHUNKS ($INDEX_JSON)"
  fi
else
  fail "case1-index: index did not return JSON ($INDEX_JSON)"
fi

Q1_JSON="$(dce ohctl rag query --collection "$COLL" "$NONCE" --k 5 2>/dev/null)"
if echo "$Q1_JSON" | jq -e '.results | length >= 1' >/dev/null 2>&1; then
  TOP_FILE="$(echo "$Q1_JSON" | jq -r '.results[0].filename // ""')"
  TOP_CONTENT="$(echo "$Q1_JSON" | jq -r '.results[0].content // ""')"
  TOP_RANK="$(echo "$Q1_JSON" | jq -r '.results[0].rank // 0')"
  if [ "$TOP_FILE" = "$KNOWN_FILE" ] && echo "$TOP_CONTENT" | grep -q "$NONCE"; then
    pass "case1-query: '$KNOWN_FILE' ranked first (rank=$TOP_RANK) and contains the term"
  else
    fail "case1-query: expected top filename='$KNOWN_FILE' containing '$NONCE', got filename='$TOP_FILE'"
  fi
else
  fail "case1-query: query for '$NONCE' returned no results ($Q1_JSON)"
fi

# cleanup tmp folder in container (best-effort; chunks remain in db but are isolated by collection)
dce rm -rf "$TMPDIR_C" >/dev/null 2>&1 || true

# ---------------------------------------------------------------------------
# Cases 2-4: seeded real collections must return grounded chunks
# Each: >=1 result, with non-empty content AND non-empty source (filename).
# ---------------------------------------------------------------------------
check_seeded() {
  local label="$1" coll="$2" query="$3"
  local out
  out="$(dce ohctl rag query --collection "$coll" "$query" --k 3 2>/dev/null)"
  if ! echo "$out" | jq -e '.results | length >= 1' >/dev/null 2>&1; then
    fail "$label: query '$query' on '$coll' returned 0 results"
    return
  fi
  # require the top hit to have both content and a source filename
  local grounded
  grounded="$(echo "$out" | jq -r '
    [ .results[] | select((.content // "" | length) > 0) | select((.filename // "" | length) > 0) ] | length')"
  if [ "${grounded:-0}" -ge 1 ]; then
    local src cnt
    src="$(echo "$out" | jq -r '.results[0].filename')"
    cnt="$(echo "$out" | jq -r '.results[0].content | length')"
    pass "$label: '$coll' \"$query\" -> grounded chunks=$grounded (top source=$src, content len=$cnt)"
  else
    fail "$label: '$coll' \"$query\" returned results but none had both content+source ($out)"
  fi
}

check_seeded "case2-kb-product"     "kb-product"     "customer discovery"
check_seeded "case3-kb-fundraising" "kb-fundraising" "pitch"
check_seeded "case4-kb-tunisia"     "kb-tunisia"     "Startup Act"

# ---------------------------------------------------------------------------
echo "=== Summary: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] || exit 1
exit 0
