# Architecture

This document describes the OfficeHours.ai MVP: its components, how data flows
through the system, and the **claude → ohctl → job-worker** pipeline that is the
core of the product. It is written against the frozen contract in
[`BUILD_SPEC.md`](../BUILD_SPEC.md); the names used here (stages, Signal names,
`ohctl` commands, API routes, DB tables) are taken verbatim from that spec.

## 1. Components

The system is three containers orchestrated by `docker-compose`.

| Service | Tech | Responsibility |
|---------|------|----------------|
| `db` | `postgres:16` (+ `pgcrypto`) | System of record. Holds users, profiles, the 5 Signals, sessions/messages, goals, action items, documents/chunks (with FTS `tsvector`), the Mon Parcours `events` timeline, and the `agent_jobs` queue. Volume `dbdata`. |
| `api` | Go (chi / net/http) | HTTP API on `:8080`, JWT auth, the **job worker** goroutine pool, and the bundled binaries `ohctl` and `claude`. Mounts the host's `~/.claude` read-only (Plan A subscription auth), an `uploads` volume, plus `./seed` and `./config` read-only. |
| `web` | React + Vite + TS, served by nginx | Dark UI (YC × Linear). nginx serves the SPA and proxies `/api/*` → `api:8080`. |

Inside the `api` service there are several logical modules
(`internal/{db,models,auth,handlers,agent,rag,config}`), but the two that matter
for understanding the runtime are:

- **HTTP handlers** — translate REST calls into DB writes and *enqueue jobs*.
  Handlers never call Claude directly; they only insert rows into `agent_jobs`.
- **Job worker** — a pool of goroutines that poll `agent_jobs`
  (`queued → running → done/error`), render a prompt per job type, and exec the
  `claude` CLI. The agent's *hands* are the `ohctl` binary, which talks straight
  to Postgres.

This separation is deliberate: the HTTP path is always fast and synchronous
(insert a row, return a `job_id`), while the slow, non-deterministic LLM work
happens asynchronously and is observable through `GET /api/jobs/:id`.

## 2. The brain: Claude in headless mode

The AI is **Claude Code running headless** (Plan A — reuse the host's logged-in
Claude subscription, no API key required). The `api` image installs Node and
`@anthropic-ai/claude-code`, so `claude` is on `PATH`.

For each job, the worker creates a temporary working directory, puts `ohctl` on
`PATH`, and execs:

```
claude -p "<rendered prompt>" \
  --output-format json \
  --dangerously-skip-permissions \
  --add-dir <workdir>
```

- **Auth**: the mounted `/root/.claude` subscription is used by default. If
  `ANTHROPIC_API_KEY` is set, the API key is used instead.
- **Output**: `--output-format json` is parsed; the assistant's *final text* is
  the chat reply and is stored as an `assistant` message (for advisor/learn
  jobs).
- **Side effects**: the prompt tells the agent it has `ohctl` and *must* use it
  to read context (profile, signals, prior messages, RAG) and to persist results
  (signals, goals, action items, events). The agent does not write to Postgres
  directly — it shells out to `ohctl`, which is the only sanctioned database
  surface for agents.

> Prerequisite documented in the README: run `claude` login on the **host**
> first, so the mounted `~/.claude` carries valid credentials.

## 3. ohctl: the agent's CLI

`ohctl` (cobra, `backend/cmd/ohctl`) connects to `DATABASE_URL` and **prints JSON
to stdout** for every command. It is used by agents (via Bash) and by
humans/seed scripts. It is the contract boundary between "the agent reasoning"
and "the database." The full command surface is frozen in BUILD_SPEC §3; the
groups are:

- **read context**: `profile get`, `signal list`, `goal list`, `session get`,
  `rag query`.
- **persist results**: `profile set-stage`, `signal set`, `goal create` /
  `goal done`, `action-item create`, `session message`, `session conclude`,
  `event add`.
- **ingestion / setup**: `rag index`, `seed user`, `seed demo`.

Because both the worker and operators go through the same JSON CLI, an agent run
is reproducible from a shell: anything the agent did, a human can do and inspect.

## 4. Job types

Four job `type`s flow through `agent_jobs`, each backed by an agent definition
(markdown + frontmatter) under `/seed`:

| Job type | Trigger | Agent def | What it does |
|----------|---------|-----------|--------------|
| `diagnoser` | `POST /api/onboarding` | `/seed/agents/diagnoser.md` | Reads the founder's company text, classifies the maturity **stage** (one of the 6) with evidence, and seeds initial Signals. |
| `advisor` | `POST /api/sessions/:id/messages` (office_hours) | `/seed/advisors/*.md` | Specialist reply grounded in `ohctl rag query`. Stores the assistant message. |
| `learn` | `POST /api/sessions/:id/messages` (learn) | `/seed/learn/*.md` | Concept tutor reply grounded in a topic KB collection. Stores the assistant message. |
| `scorer` | `POST /api/sessions/:id/conclude` | `/seed/agents/scorer.md` | Recomputes the **5 composite Signals** (subscores + rationale, gated aggregation), creates Goals + Action Items grounded in KB programs, writes a Mon Parcours `event`, may close the default goal. |

The 6 stages (`Ideation`, `Market Validation`, `Structuration`, `Fundraising`,
`Launch Planning`, `Growth`) and the 5 Signal names (`Market`,
`Commercial Offer`, `Innovation`, `Scalability`, `Green`) are exact strings from
the schema and must not drift.

## 5. Data flow — the critical path

The demo-critical path (BUILD_SPEC §12) chains the job types together:

1. **Onboarding.** Founder describes the company once. `POST /api/onboarding`
   upserts the `profiles` row, creates the default goal "Start an Office Hours
   session", and enqueues a `diagnoser` job. The diagnoser sets `stage` +
   `stage_evidence` and writes initial Signals via `ohctl`.
2. **Office Hours.** Founder opens a session with the Product Advisor and chats.
   Each message enqueues an `advisor` job that reads the dossier + Signals,
   runs `ohctl rag query` against the advisor's KB collection, and stores a
   grounded reply.
3. **Conclude.** `POST /api/sessions/:id/conclude` enqueues a `scorer` job. The
   scorer recomputes the 5 Signals with sub-score breakdowns and rationale
   (gated aggregation — a weak fundamental caps the composite), creates Goals
   and Action Items each matched to a real KB program (`program_ref`), and
   appends a `session` event to Mon Parcours.
4. **Dashboard.** `GET /api/dashboard` returns the stage, the 5 Signals with
   sub-score breakdowns, stats, goals, and the parcours timeline.
5. **Data Room.** `POST /api/documents` stores an uploaded file and enqueues RAG
   indexing into the founder's per-user collection; a later session can cite it.

Every step that changes state also (directly or via the scorer) appends to the
`events` table, which is what powers the Mon Parcours timeline.

## 6. Pipeline diagram

```
                                 HOST
                       ┌───────────────────────┐
                       │  ~/.claude (logged in) │  Plan A subscription auth
                       └───────────┬───────────┘
                                   │ mounted :ro
┌──────────┐   /api/*    ┌─────────▼──────────────────────────────────────┐
│   web    │────────────▶│                   api  (Go)                      │
│ (nginx + │  proxied    │                                                 │
│  React)  │◀────────────│  HTTP handlers ── insert ──▶ agent_jobs (queued)│
└──────────┘   JSON      │       │                          ▲              │
                         │       │ return {job_id}          │ poll         │
   GET /jobs/:id ◀───────┤       ▼                          │              │
   (poll status)         │  ┌──────────────── job worker pool ──────────┐  │
                         │  │  1. claim job (queued→running)             │  │
                         │  │  2. render prompt for type                 │  │
                         │  │  3. exec:  claude -p "<prompt>"            │  │
                         │  │            --output-format json           │  │
                         │  │            --dangerously-skip-permissions │  │
                         │  │            --add-dir <workdir>            │  │
                         │  │                  │   ▲                     │  │
                         │  │         agent uses│   │final text          │  │
                         │  │            ohctl  │   │= chat reply         │  │
                         │  │                  ▼   │                     │  │
                         │  │     ┌──── ohctl (JSON CLI) ────┐           │  │
                         │  │     │ read: profile/signal/    │           │  │
                         │  │     │       session/rag query  │           │  │
                         │  │     │ write: signal set, goal, │           │  │
                         │  │     │        action-item,      │           │  │
                         │  │     │        session message,  │           │  │
                         │  │     │        event add         │           │  │
                         │  │     └────────────┬─────────────┘           │  │
                         │  │  4. store output, mark done/error          │  │
                         │  └────────────────────┼──────────────────────┘  │
                         └───────────────────────┼─────────────────────────┘
                                                 │ SQL
                                       ┌─────────▼─────────┐
                                       │   db  (postgres)  │
                                       │  users, profiles, │
                                       │  signals, sessions│
                                       │  messages, goals, │
                                       │  action_items,    │
                                       │  documents/chunks │
                                       │  events, agent_jobs│
                                       └───────────────────┘
```

The same diagram as a sequence, for the conclude → scorer case:

```
web → api:        POST /api/sessions/:id/conclude
api → db:         insert agent_jobs(type=scorer, status=queued)
api → web:        { job_id }
worker → db:      claim job (queued → running)
worker → claude:  claude -p "<scorer prompt>" --output-format json ...
claude → ohctl:   ohctl profile get / signal list / session get / rag query
claude → ohctl:   ohctl signal set (x5) / goal create / action-item create / event add
ohctl → db:       SQL writes
claude → worker:  final JSON (assistant text + summary)
worker → db:      agent_jobs.output, status = done
web → api:        GET /api/jobs/:id  (polled) → done
web → api:        GET /api/dashboard → updated stage + 5 Signals + goals + parcours
```

## 7. Design notes & invariants

- **One database surface for agents.** Agents never run raw SQL; they only call
  `ohctl`. This keeps writes validated, JSON-shaped, and auditable.
- **Idempotent Signals.** `signals` has `unique(user_id, name)`, so re-running
  the scorer updates the five rows in place rather than duplicating them.
- **Resumable sessions.** A session is addressed by its UUID; chat history lives
  in `messages`, so a session can be reopened and continued with full context.
- **Async by default.** Anything that calls Claude is a job; the UI shows a
  thinking state and polls `GET /api/jobs/:id`. The HTTP layer stays responsive.
- **Graceful degradation.** PDF text extraction failures are logged and skipped
  (BUILD_SPEC §3); a missing fundamental in scoring is surfaced, not hidden
  (see [`scoring-methodology.md`](scoring-methodology.md)).
