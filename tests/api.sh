#!/usr/bin/env bash
# Integration tests for the OfficeHours HTTP API.
# Asserts status codes and response shapes against http://localhost:8080/api.
#
# Usage: tests/api.sh
# Requires: curl, jq, and a running stack (docker compose up).
#
# This script seeds its OWN isolated user (api-test@x.io) so it does not
# depend on or disturb other data. It does NOT require async agent jobs to
# complete (the worker may need Claude) -- it only asserts a job is enqueued
# and reaches a valid status (queued|running|done).
set -euo pipefail

BASE="${BASE:-http://localhost:8080/api}"
EMAIL="api-test@x.io"
PASSWORD="p"
NAME="API"

PASS=0
FAIL=0
FAILED_CASES=()

# --- helpers ---------------------------------------------------------------

# ok NAME -- record a passing case
ok() {
  PASS=$((PASS + 1))
  printf 'PASS  %s\n' "$1"
}

# bad NAME DETAIL -- record a failing case (does not exit; we want full run)
bad() {
  FAIL=$((FAIL + 1))
  FAILED_CASES+=("$1")
  printf 'FAIL  %s -- %s\n' "$1" "$2" >&2
}

# assert_eq NAME EXPECTED ACTUAL
assert_eq() {
  if [[ "$2" == "$3" ]]; then
    ok "$1"
  else
    bad "$1" "expected '$2' got '$3'"
  fi
}

# assert_status NAME EXPECTED_CODE 'curl args...'
# Performs the request capturing body to $BODY and status to $STATUS.
BODY=""
STATUS=""
req() {
  # $@ are curl arguments after -s -o body -w status
  local tmp
  tmp="$(mktemp)"
  STATUS="$(curl -s -o "$tmp" -w '%{http_code}' "$@")"
  BODY="$(cat "$tmp")"
  rm -f "$tmp"
}

# require a command exists
command -v jq  >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "curl is required" >&2; exit 1; }

echo "== OfficeHours API integration tests =="
echo "BASE=$BASE"

# --- 0. seed isolated user -------------------------------------------------
# Idempotent: ohctl seed user returns created:true|false. We don't fail the
# suite if seeding via docker is unavailable AND the user can already log in.
if command -v docker >/dev/null 2>&1; then
  docker compose exec -T api ohctl seed user \
    --email "$EMAIL" --password "$PASSWORD" --name "$NAME" >/dev/null 2>&1 \
    || echo "warn: seed command returned non-zero (user may already exist)" >&2
else
  echo "warn: docker not found; assuming user $EMAIL already exists" >&2
fi

# --- 1. POST /auth/login ---------------------------------------------------
req -X POST "$BASE/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
assert_eq "POST /auth/login status 200" "200" "$STATUS"

TOKEN="$(echo "$BODY" | jq -r '.token // empty')"
if [[ -n "$TOKEN" && "$TOKEN" != "null" ]]; then
  ok "POST /auth/login returns token"
else
  bad "POST /auth/login returns token" "no token in body: $BODY"
  echo "Cannot continue without a token." >&2
  echo "== $PASS passed, $FAIL failed =="
  exit 1
fi
USER_ID="$(echo "$BODY" | jq -r '.user.id // empty')"
assert_eq "POST /auth/login returns user.id" "true" "$([[ -n "$USER_ID" ]] && echo true || echo false)"

AUTH=(-H "Authorization: Bearer $TOKEN")

# --- 2. GET /me ------------------------------------------------------------
req "${AUTH[@]}" "$BASE/me"
assert_eq "GET /me status 200" "200" "$STATUS"
assert_eq "GET /me user.email matches" "$EMAIL" "$(echo "$BODY" | jq -r '.user.email')"
# profile key present (may be null before onboarding)
assert_eq "GET /me has profile key" "true" "$(echo "$BODY" | jq 'has("profile")')"

# --- 3. POST /onboarding ---------------------------------------------------
req -X POST "$BASE/onboarding" "${AUTH[@]}" \
  -H 'Content-Type: application/json' \
  -d '{"text":"We build an AI scheduling assistant for SMB clinics. Pre-seed, 2 founders, no revenue yet."}'
assert_eq "POST /onboarding status 200" "200" "$STATUS"
JOB_ID="$(echo "$BODY" | jq -r '.job_id // empty')"
if [[ -n "$JOB_ID" && "$JOB_ID" != "null" ]]; then
  ok "POST /onboarding returns job_id"
else
  bad "POST /onboarding returns job_id" "no job_id in body: $BODY"
fi

# --- 4. GET /jobs/:id ------------------------------------------------------
# Status must be one of queued|running|done. Do NOT require completion.
if [[ -n "${JOB_ID:-}" && "$JOB_ID" != "null" ]]; then
  req "${AUTH[@]}" "$BASE/jobs/$JOB_ID"
  assert_eq "GET /jobs/:id status 200" "200" "$STATUS"
  JST="$(echo "$BODY" | jq -r '.status')"
  case "$JST" in
    queued|running|done)
      ok "GET /jobs/:id status in {queued,running,done} (got '$JST')" ;;
    *)
      bad "GET /jobs/:id status in {queued,running,done}" "got '$JST'" ;;
  esac
  assert_eq "GET /jobs/:id has output key" "true" "$(echo "$BODY" | jq 'has("output")')"
else
  bad "GET /jobs/:id status 200" "skipped: no job_id from onboarding"
fi

# --- 5. GET /profile -------------------------------------------------------
# Onboarding upserts the profile row, so it must now exist.
req "${AUTH[@]}" "$BASE/profile"
assert_eq "GET /profile status 200" "200" "$STATUS"
assert_eq "GET /profile has profile.stage" "true" \
  "$(echo "$BODY" | jq '.profile | has("stage")')"
assert_eq "GET /profile signals is array" "true" \
  "$(echo "$BODY" | jq '.signals | type == "array"')"

# --- 6. GET /signals -------------------------------------------------------
req "${AUTH[@]}" "$BASE/signals"
assert_eq "GET /signals status 200" "200" "$STATUS"
assert_eq "GET /signals returns array" "true" \
  "$(echo "$BODY" | jq 'type == "array"')"

# --- 7. GET /dashboard -----------------------------------------------------
req "${AUTH[@]}" "$BASE/dashboard"
assert_eq "GET /dashboard status 200" "200" "$STATUS"
assert_eq "GET /dashboard has stage" "true"    "$(echo "$BODY" | jq 'has("stage")')"
assert_eq "GET /dashboard has signals" "true"  "$(echo "$BODY" | jq 'has("signals")')"
assert_eq "GET /dashboard has parcours" "true" "$(echo "$BODY" | jq 'has("parcours")')"
assert_eq "GET /dashboard signals is array" "true"  "$(echo "$BODY" | jq '.signals | type == "array"')"
assert_eq "GET /dashboard parcours is array" "true" "$(echo "$BODY" | jq '.parcours | type == "array"')"

# --- 8. GET /goals ---------------------------------------------------------
req "${AUTH[@]}" "$BASE/goals"
assert_eq "GET /goals status 200" "200" "$STATUS"
assert_eq "GET /goals returns array" "true" "$(echo "$BODY" | jq 'type == "array"')"
# onboarding created a default goal
GOALS_LEN="$(echo "$BODY" | jq 'length')"
assert_eq "GET /goals has default goal (>=1)" "true" \
  "$([[ "$GOALS_LEN" -ge 1 ]] && echo true || echo false)"

# --- 9. GET /advisors (4) --------------------------------------------------
req "${AUTH[@]}" "$BASE/advisors"
assert_eq "GET /advisors status 200" "200" "$STATUS"
assert_eq "GET /advisors returns array" "true" "$(echo "$BODY" | jq 'type == "array"')"
assert_eq "GET /advisors count == 4" "4" "$(echo "$BODY" | jq 'length')"
# pick an advisor_key for session creation below
ADVISOR_KEY="$(echo "$BODY" | jq -r '.[0].key')"
assert_eq "GET /advisors items have key" "true" \
  "$(echo "$BODY" | jq 'all(.[]; has("key"))')"

# --- 10. GET /learn (>=3) --------------------------------------------------
req "${AUTH[@]}" "$BASE/learn"
assert_eq "GET /learn status 200" "200" "$STATUS"
assert_eq "GET /learn returns array" "true" "$(echo "$BODY" | jq 'type == "array"')"
LEARN_LEN="$(echo "$BODY" | jq 'length')"
assert_eq "GET /learn count >= 3" "true" \
  "$([[ "$LEARN_LEN" -ge 3 ]] && echo true || echo false)"

# --- 11. POST /sessions (returns uuid) -------------------------------------
[[ -z "${ADVISOR_KEY:-}" || "$ADVISOR_KEY" == "null" ]] && ADVISOR_KEY="fundraising"
req -X POST "$BASE/sessions" "${AUTH[@]}" \
  -H 'Content-Type: application/json' \
  -d "{\"advisor_key\":\"$ADVISOR_KEY\"}"
assert_eq "POST /sessions status 200" "200" "$STATUS"
SESSION_ID="$(echo "$BODY" | jq -r '.id // empty')"
UUID_RE='^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
if [[ "$SESSION_ID" =~ $UUID_RE ]]; then
  ok "POST /sessions returns uuid id"
else
  bad "POST /sessions returns uuid id" "got '$SESSION_ID'"
fi

# --- 12. GET /sessions -----------------------------------------------------
req "${AUTH[@]}" "$BASE/sessions"
assert_eq "GET /sessions status 200" "200" "$STATUS"
assert_eq "GET /sessions returns array" "true" "$(echo "$BODY" | jq 'type == "array"')"
if [[ -n "${SESSION_ID:-}" ]]; then
  assert_eq "GET /sessions contains created session" "true" \
    "$(echo "$BODY" | jq --arg id "$SESSION_ID" 'any(.[]; .id == $id)')"
fi

# --- 13. GET /sessions/:id -------------------------------------------------
if [[ -n "${SESSION_ID:-}" ]]; then
  req "${AUTH[@]}" "$BASE/sessions/$SESSION_ID"
  assert_eq "GET /sessions/:id status 200" "200" "$STATUS"
  assert_eq "GET /sessions/:id has session.id match" "$SESSION_ID" \
    "$(echo "$BODY" | jq -r '.session.id')"
  assert_eq "GET /sessions/:id messages is array" "true" \
    "$(echo "$BODY" | jq '.messages | type == "array"')"
  assert_eq "GET /sessions/:id action_items is array" "true" \
    "$(echo "$BODY" | jq '.action_items | type == "array"')"
else
  bad "GET /sessions/:id status 200" "skipped: no session id"
fi

# --- 14. POST /sessions/:id/messages (returns job_id) ----------------------
if [[ -n "${SESSION_ID:-}" ]]; then
  req -X POST "$BASE/sessions/$SESSION_ID/messages" "${AUTH[@]}" \
    -H 'Content-Type: application/json' \
    -d '{"content":"How should we think about our seed round size?"}'
  assert_eq "POST /sessions/:id/messages status 200" "200" "$STATUS"
  MSG_JOB="$(echo "$BODY" | jq -r '.job_id // empty')"
  if [[ -n "$MSG_JOB" && "$MSG_JOB" != "null" ]]; then
    ok "POST /sessions/:id/messages returns job_id"
  else
    bad "POST /sessions/:id/messages returns job_id" "no job_id: $BODY"
  fi
else
  bad "POST /sessions/:id/messages status 200" "skipped: no session id"
fi

# --- 15. GET /documents ----------------------------------------------------
req "${AUTH[@]}" "$BASE/documents"
assert_eq "GET /documents status 200" "200" "$STATUS"
assert_eq "GET /documents returns array" "true" "$(echo "$BODY" | jq 'type == "array"')"

# --- summary ---------------------------------------------------------------
echo "== $PASS passed, $FAIL failed =="
if [[ "$FAIL" -gt 0 ]]; then
  printf 'Failed cases:\n'
  printf '  - %s\n' "${FAILED_CASES[@]}"
  exit 1
fi
exit 0
