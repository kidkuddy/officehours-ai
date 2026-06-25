#!/usr/bin/env bash
# Integration tests for the ohctl CLI against the running docker compose stack.
#
# Strategy:
#   - All ohctl calls run inside the api container:  docker compose exec -T api ohctl ...
#   - We operate on our OWN isolated user (cli-test@x.io) so we never touch
#     other suites' data.
#   - Sessions have no `ohctl ... create`, so the one session we exercise is
#     created via the HTTP API (POST /api/sessions) using a JWT we obtain by
#     logging in as our isolated user.
#   - Every command is one test case. For each we assert:
#       1. exit code == 0 (or != 0 for negative cases),
#       2. stdout is valid JSON,
#       3. expected keys / values are present,
#     and where it mutates state we verify the DB side-effect with a follow-up
#     get/list.
#
# Rerunnable: seeding is idempotent (ohctl seed user is upsert-by-email; rag
#   index re-chunks into the collection). Goals/signals/etc. accumulate but the
#   assertions don't depend on exact counts, only on the presence of the rows we
#   just created (matched by the unique id/title returned by the create call).
#
# Requires on the host: docker, jq, curl, python3.

set -euo pipefail

# --- config -----------------------------------------------------------------
EMAIL="cli-test@x.io"
PASSWORD="p"
NAME="CLI"
API="http://localhost:8080/api"
RUN_TAG="$(date +%s)-$$"               # unique-ish suffix for this run
COLLECTION="cli-test-${RUN_TAG}"
RAG_DIR="/tmp/ohctl-ragtest-${RUN_TAG}" # path INSIDE the api container
RAG_MARKER="quasar nebula gizmo ${RUN_TAG}"

OHCTL() { docker compose exec -T api ohctl "$@"; }

# --- harness ----------------------------------------------------------------
PASS=0
FAIL=0
declare -a RESULTS

ok()   { PASS=$((PASS+1)); RESULTS+=("PASS | $1"); printf 'PASS | %s\n' "$1"; }
bad()  { FAIL=$((FAIL+1)); RESULTS+=("FAIL | $1 :: $2"); printf 'FAIL | %s :: %s\n' "$1" "$2"; }

# is_json <string> -> 0 if valid JSON
is_json() { printf '%s' "$1" | jq -e . >/dev/null 2>&1; }

# jq_get <json> <filter> -> prints value (raw)
jq_get() { printf '%s' "$1" | jq -r "$2"; }

# Run a command that must exit 0 and emit valid JSON; checks a jq predicate.
# usage: case_ok <name> <jq-predicate> -- <command...>
case_ok() {
  local name="$1" pred="$2"; shift 3   # drop name, pred, "--"
  local out rc
  set +e
  out="$("$@" 2>&1)"; rc=$?
  set -e
  if [[ $rc -ne 0 ]]; then bad "$name" "exit=$rc out=${out:0:300}"; LAST_OUT=""; return; fi
  if ! is_json "$out"; then bad "$name" "stdout not JSON: ${out:0:300}"; LAST_OUT=""; return; fi
  if ! printf '%s' "$out" | jq -e "$pred" >/dev/null 2>&1; then
    bad "$name" "predicate '$pred' failed on: ${out:0:300}"; LAST_OUT="$out"; return
  fi
  ok "$name"
  LAST_OUT="$out"
}

# Negative case: command must exit non-zero AND emit JSON with an "error" key.
# usage: case_err <name> -- <command...>
case_err() {
  local name="$1"; shift 2  # drop name, "--"
  local out rc
  set +e
  out="$("$@" 2>&1)"; rc=$?
  set -e
  if [[ $rc -eq 0 ]]; then bad "$name" "expected non-zero exit, got 0: ${out:0:200}"; return; fi
  if ! is_json "$out"; then bad "$name" "stdout not JSON: ${out:0:200}"; return; fi
  if ! printf '%s' "$out" | jq -e 'has("error")' >/dev/null 2>&1; then
    bad "$name" "no error key: ${out:0:200}"; return
  fi
  ok "$name"
}

# ============================================================================
# 1. seed user (idempotent) -> capture user id
# ============================================================================
case_ok "seed user" \
  '.email=="'"$EMAIL"'" and .name=="'"$NAME"'" and (.id|type=="string") and (.id|length>0)' \
  -- OHCTL seed user --email "$EMAIL" --password "$PASSWORD" --name "$NAME"
UID_="$(jq_get "${LAST_OUT}" '.id')"
if [[ -z "$UID_" || "$UID_" == "null" ]]; then
  echo "FATAL: could not resolve user id; aborting." >&2
  exit 1
fi
echo "    user_id=$UID_"

# ============================================================================
# 2. profile get -> belongs to our user, has a stage
# ============================================================================
case_ok "profile get" \
  '.user_id=="'"$UID_"'" and has("stage")' \
  -- OHCTL profile get --user "$UID_"

# ============================================================================
# 3. profile set-stage (Structuration) + get reflects it (DB side-effect)
# ============================================================================
case_ok "profile set-stage" \
  '.ok==true and .stage=="Structuration"' \
  -- OHCTL profile set-stage --user "$UID_" --stage "Structuration" \
       --evidence '[{"note":"cli test '"$RUN_TAG"'"}]'

case_ok "profile get reflects stage" \
  '.user_id=="'"$UID_"'" and .stage=="Structuration"' \
  -- OHCTL profile get --user "$UID_"

# ============================================================================
# 4. signal set (Market) + signal list shows it (DB side-effect)
# ============================================================================
case_ok "signal set" \
  '.user_id=="'"$UID_"'" and .name=="Market" and .score==4 and (.id|length>0)' \
  -- OHCTL signal set --user "$UID_" --name "Market" --score 4 \
       --subscores '[{"k":"tam","v":3}]' --rationale "cli ${RUN_TAG}"

case_ok "signal list shows it" \
  '(.signals|type=="array") and ([.signals[]|select(.name=="Market" and .score==4)]|length>=1)' \
  -- OHCTL signal list --user "$UID_"

# ============================================================================
# 5. goal create -> capture id; goal list shows open; goal done flips status
# ============================================================================
GOAL_TITLE="cli-goal-${RUN_TAG}"
case_ok "goal create" \
  '.user_id=="'"$UID_"'" and .title=="'"$GOAL_TITLE"'" and .status=="open" and (.id|length>0)' \
  -- OHCTL goal create --user "$UID_" --title "$GOAL_TITLE" --desc "created by cli.sh"
GOAL_ID="$(jq_get "${LAST_OUT}" '.id')"
echo "    goal_id=$GOAL_ID"

case_ok "goal list shows open goal" \
  '[.goals[]|select(.id=="'"$GOAL_ID"'" and .status=="open")]|length==1' \
  -- OHCTL goal list --user "$UID_"

case_ok "goal done" \
  '.ok==true and .id=="'"$GOAL_ID"'" and .status=="done"' \
  -- OHCTL goal done --id "$GOAL_ID"

case_ok "goal list reflects done (DB side-effect)" \
  '[.goals[]|select(.id=="'"$GOAL_ID"'" and .status=="done")]|length==1' \
  -- OHCTL goal list --user "$UID_"

# ============================================================================
# 6. session: create via API (no ohctl create), then ohctl get/message/conclude
# ============================================================================
TOKEN="$(curl -s -X POST "$API/auth/login" -H 'Content-Type: application/json' \
          -d '{"email":"'"$EMAIL"'","password":"'"$PASSWORD"'"}' \
          | jq -r '.token // empty')"
if [[ -z "$TOKEN" ]]; then
  bad "api login (session setup)" "no token from /auth/login"
  SESSION_ID=""
else
  ok "api login (session setup)"
  SESS_JSON="$(curl -s -X POST "$API/sessions" -H "Authorization: Bearer $TOKEN" \
                -H 'Content-Type: application/json' \
                -d '{"advisor_key":"general","kind":"office_hours"}')"
  if is_json "$SESS_JSON" && [[ "$(jq -r '.id // empty' <<<"$SESS_JSON")" != "" ]]; then
    SESSION_ID="$(jq -r '.id' <<<"$SESS_JSON")"
    ok "api create session"
    echo "    session_id=$SESSION_ID"
  else
    bad "api create session" "${SESS_JSON:0:200}"
    SESSION_ID=""
  fi
fi

if [[ -n "${SESSION_ID:-}" ]]; then
  # action-item create bound to this session
  AI_TITLE="cli-action-${RUN_TAG}"
  case_ok "action-item create" \
    '.user_id=="'"$UID_"'" and .title=="'"$AI_TITLE"'" and .horizon=="short" and .status=="open" and (.id|length>0)' \
    -- OHCTL action-item create --user "$UID_" --session "$SESSION_ID" \
         --title "$AI_TITLE" --horizon short --rationale "cli ${RUN_TAG}"

  # session get -> meta + messages + action_items; our action item present
  case_ok "session get" \
    '.session.id=="'"$SESSION_ID"'" and (.messages|type=="array") and ([.action_items[]|select(.title=="'"$AI_TITLE"'")]|length>=1)' \
    -- OHCTL session get "$SESSION_ID"

  # session message append
  case_ok "session message" \
    '.session_id=="'"$SESSION_ID"'" and .role=="user" and .content=="hello from cli '"$RUN_TAG"'" and (.id|length>0)' \
    -- OHCTL session message "$SESSION_ID" --role user --content "hello from cli ${RUN_TAG}"

  # session get reflects the appended message (DB side-effect)
  case_ok "session get reflects message" \
    '[.messages[]|select(.role=="user" and .content=="hello from cli '"$RUN_TAG"'")]|length>=1' \
    -- OHCTL session get "$SESSION_ID"

  # session conclude
  case_ok "session conclude" \
    '.ok==true and .id=="'"$SESSION_ID"'" and .status=="concluded"' \
    -- OHCTL session conclude "$SESSION_ID" --outcomes "wrapped up by cli ${RUN_TAG}"

  # session get reflects concluded status (DB side-effect)
  case_ok "session get reflects concluded" \
    '.session.id=="'"$SESSION_ID"'" and .session.status=="concluded"' \
    -- OHCTL session get "$SESSION_ID"
else
  for n in "action-item create" "session get" "session message" \
           "session get reflects message" "session conclude" \
           "session get reflects concluded"; do
    bad "$n" "skipped: no session id"
  done
fi

# ============================================================================
# 7. event add
# ============================================================================
case_ok "event add" \
  '.user_id=="'"$UID_"'" and .kind=="cli_test" and (.id|length>0) and (.payload.tag=="'"$RUN_TAG"'")' \
  -- OHCTL event add --user "$UID_" --kind cli_test --payload '{"tag":"'"$RUN_TAG"'"}'

# ============================================================================
# 8. rag index a tmp folder with a known .md, then rag query returns it
# ============================================================================
# Create the folder + a known .md INSIDE the api container (its own filesystem).
docker compose exec -T api sh -c \
  "mkdir -p '$RAG_DIR' && printf '# CLI RAG doc\n\nThis document mentions %s as a unique phrase.\n' '$RAG_MARKER' > '$RAG_DIR/known.md'"

case_ok "rag index" \
  '.collection=="'"$COLLECTION"'" and .files>=1 and .chunks>=1 and (.skipped|type=="array")' \
  -- OHCTL rag index "$RAG_DIR" --collection "$COLLECTION" --user "$UID_"

case_ok "rag query returns indexed doc" \
  '.collection=="'"$COLLECTION"'" and ([.results[]|select(.filename=="known.md" and (.content|test("'"$RUN_TAG"'")))]|length>=1)' \
  -- OHCTL rag query "$RAG_MARKER" --collection "$COLLECTION" --user "$UID_"

# ============================================================================
# 9. negative cases: validation must fail with JSON error + non-zero exit
# ============================================================================
case_err "profile set-stage rejects invalid stage" \
  -- OHCTL profile set-stage --user "$UID_" --stage "NotAStage" --evidence '[]'

case_err "signal set rejects invalid name" \
  -- OHCTL signal set --user "$UID_" --name "Bogus" --score 1 --subscores '[]'

# ============================================================================
# summary
# ============================================================================
echo
echo "================ ohctl CLI integration summary ================"
for line in "${RESULTS[@]}"; do echo "$line"; done
echo "---------------------------------------------------------------"
echo "PASS=$PASS  FAIL=$FAIL"
echo "==============================================================="

[[ $FAIL -eq 0 ]]
