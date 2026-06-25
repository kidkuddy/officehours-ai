# Tests

Integration / E2E test suites for OfficeHours.ai. These run **against a live
docker compose stack**, not against mocks — they drive the real `ohctl` CLI, the
real HTTP API, and the real Postgres-backed RAG index.

## Prerequisites

1. The stack must be **up and healthy**:

   ```bash
   cp .env.example .env          # if you haven't already
   docker compose up --build -d  # db + api + web all Up
   docker compose ps             # confirm ains-api / ains-db / ains-web are Up
   ```

2. Host tooling used by the scripts:
   - `docker compose` (the scripts shell into the `api` container).
   - `jq` (JSON assertions on the host).
   - `curl` (HTTP suite).

3. The KB seed is **not** required for the suites — each script seeds its own
   isolated fixtures. The RAG suite does query the three pre-seeded KB
   collections (`kb-product`, `kb-fundraising`, `kb-tunisia`); if you ran
   `ohctl seed demo` (see the root README) they are present. A fresh
   `docker compose up` followed by `ohctl seed demo` guarantees them.

All three scripts are `set -euo pipefail`, **idempotent**, and **rerunnable** —
each isolates its work to a dedicated test user / per-run-tagged records / a
per-run RAG collection so suites don't interfere with each other or with demo
data.

## Running the suites

From the repo root:

```bash
./tests/cli.sh    # ohctl CLI integration (23 cases)
./tests/api.sh    # HTTP API integration (45 assertions / 15 endpoint groups)
./tests/rag.sh    # RAG retrieval E2E (5 cases)
```

Each script exits `0` on success and non-zero on the first failed assertion.

---

## `tests/cli.sh` — ohctl CLI integration

Drives the agent's CLI via `docker compose exec -T api ohctl ...` against the
live stack. Each command is one case; for every case it asserts the **exit
code**, that **stdout is valid JSON** (piped through `jq`), and that the
**expected keys/values** are present. Mutating commands are additionally
verified by a follow-up `get`/`list` confirming the DB side-effect.

Isolation: a dedicated user `cli-test@x.io`, a per-run goal / action-item /
event tagged with a unique run id, and a per-run RAG collection.

What it asserts (23 cases):

- `seed user` — creates/upserts the test user (idempotent by email).
- `profile get` / `profile set-stage` — sets stage `Structuration`; follow-up
  `profile get` confirms the DB side-effect. A negative case asserts an
  **invalid stage** is rejected with non-zero exit + `{"error":...}`.
- `signal set` / `signal list` — sets a `Market` signal (score 4); list
  confirms it. A negative case asserts an **invalid signal name** is rejected.
- `goal create` / `goal list` / `goal done` — open → done lifecycle, each step
  verified by a follow-up list.
- `session get` / `session message` / `session conclude` — the session itself
  is created via the HTTP API (`POST /api/sessions` with a JWT from
  `POST /api/auth/login`, since `ohctl` exposes no `session create`), then
  exercised with `ohctl`. Messages and `concluded` status are verified by
  follow-up `session get`.
- `action-item create` — created against the session; surfaced via
  `session get`.
- `event add` — payload round-trips.
- `rag index` / `rag query` — a known `.md` with a unique run marker is written
  **inside the api container** (`ohctl rag index <folder>` reads the container
  FS), indexed into a per-run collection, and queried back; the indexed doc is
  returned.

Valid stage / signal names are taken from `backend/internal/models/models.go`.

## `tests/api.sh` — HTTP API integration

Hits the API at `http://localhost:8080/api` with an isolated seeded user
(`api-test@x.io`) and a JWT from `POST /auth/login`. 45 assertions across these
endpoint groups:

- `POST /auth/login` → token + `user.id`.
- `GET /me` → user + profile.
- `POST /onboarding` → `job_id` (upserts profile, creates default goal).
- `GET /jobs/:id` → status in `{queued,running,done}`. **The async onboarding
  job is only observed in `queued`; completion is NOT required** (the worker
  needs Claude — see the agent smoke note below).
- `GET /profile`, `GET /signals`, `GET /dashboard` (stage + signals + parcours).
- `GET /goals` (default onboarding goal present).
- `GET /advisors` → exactly 4 (product, gtm, fundraising, pitch).
- `GET /learn` → ≥ 3 concepts.
- `POST /sessions`, `GET /sessions`, `GET /sessions/:id`,
  `POST /sessions/:id/messages` (→ advisor `job_id`).
- `GET /documents`.

## `tests/rag.sh` — RAG retrieval E2E

Drives `ohctl rag` inside the `api` container (which talks directly to
Postgres); parses JSON with host `jq`. 5 cases:

1. **Fresh-collection round-trip** — writes a known `.md` with a unique nonce
   into a tmp folder inside the api container, indexes it into a fresh
   `rag-test` collection (files=1, chunks=1), queries the nonce, and asserts
   `known.md` ranks first and its content contains the term.
2–4. Queries the three pre-seeded real collections and asserts each returns
   ≥ 1 grounded chunk (content **and** source filename non-empty):
   - `kb-product` / "customer discovery"
   - `kb-fundraising` / "pitch"
   - `kb-tunisia` / "Startup Act"

Retrieval uses Postgres FTS (`ts_rank` / `plainto_tsquery`), not vector
embeddings.

---

## Agent pipeline smoke (manual / Claude-dependent)

There is **no automated script** for the full agent pipeline (diagnoser /
advisor / scorer), because it execs the host `claude` CLI and depends on the
runtime auth/exec environment. The pipeline **mechanics** (enqueue → worker pool
→ job status transition with error captured on `GET /jobs/:id`) are verified
manually, but the `claude` exec step currently fails — see the **BLOCKED** note
below. To smoke it manually once the blocker is resolved:

```bash
# Seed a user, log in, onboard, poll the job to a terminal status.
docker compose exec api ohctl seed user --email agent-test@x.io --password p --name AGENT
TOKEN=$(curl -s localhost:8080/api/auth/login -d '{"email":"agent-test@x.io","password":"p"}' | jq -r .token)
JOB=$(curl -s localhost:8080/api/onboarding -H "Authorization: Bearer $TOKEN" \
  -d '{"company_text":"<describe the company>"}' | jq -r .job_id)
curl -s localhost:8080/api/jobs/$JOB -H "Authorization: Bearer $TOKEN" | jq .
```

A successful run sets `profile.stage` + initial Signals; a blocked run reports
the exec error in the job's `error` field.

> **BLOCKED — agent jobs fail at the `claude` exec step.** Both the diagnoser
> and advisor jobs fail with:
> `--dangerously-skip-permissions cannot be used with root/sudo privileges for
> security reasons`.
> The api container runs as `uid=0 (root)` (`backend/Dockerfile` sets no
> non-root `USER`), and `backend/internal/agent/claude.go` always passes
> `--dangerously-skip-permissions`, which the `claude` CLI refuses under root.
> The mounted subscription credentials are present and this is **not** an auth
> problem — purely the root-execution guard.
>
> **Fixes (any one):**
> 1. Run the api container as a **non-root user** — add a non-root `USER` in
>    `backend/Dockerfile` and `chown` the mounted `/root/.claude` (or mount it
>    at that user's home) so `claude` accepts the flag.
> 2. **Drop `--dangerously-skip-permissions`** for the root case and use an
>    allowed-tools config instead.
> 3. Stopgap: set `ANTHROPIC_API_KEY` in `.env` — but the flag is rejected
>    regardless of auth method, so the root guard must still be resolved.
>
> A secondary warning (`missing /root/.claude.json`) can be cleared by restoring
> it from `/root/.claude/backups/.claude.json.backup.*`.
