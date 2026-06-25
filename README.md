# OfficeHours.ai

**Office hours for founders, on demand.**

An AI advisory platform that diagnoses where a startup actually stands, scores it across multiple dimensions, and orients the founder toward concrete next steps grounded in real support and financing programs.

Built for the **AINS Hackathon 2026** — Intelligent Entrepreneurial Orientation Engine, with entrepreneurship in Tunisia and the MENA region as the primary context.

---

## The problem

Tunisian founders are often stronger at pitching than at building. A polished deck attracts praise, praise feeds confirmation bias, and founders end up chasing vanity metrics instead of real validation. Without honest, expert feedback they mistake momentum for progress and spend their limited runway on the wrong things.

The outcome shows in the numbers: of the 1,000+ startups labeled under the Startup Act, only a handful meet VC standards of success.

Existing tools answer questions but do not *assess*. They hand a founder information without ever telling them whether they are actually ready for the step they think they're on.

## The idea

OfficeHours.ai works like real office hours. A founder books a session with a specialist **Advisor** — Product, Go-to-Market, Fundraising, Pitch — and pitches their startup. Each Advisor is an AI agent that reasons from the company's living profile before giving feedback, then leaves the founder with prioritized, trackable action items.

A conversational session is the *interface*. Underneath it sit three engines that do the structural work the brief requires: a diagnostic classifier, an explainable scoring model, and a grounded orientation layer that ties every recommendation to a real program.

## Who it's for

- Early- and growth-stage founders in Tunisia & MENA (pre-seed to early revenue).
- Aspiring founders and students who want to learn entrepreneurship by doing.
- Incubators and support programs that need to advise more founders than their mentors can cover.

---

## Core concepts (vocabulary)

| Term | What it is |
|------|------------|
| **Advisor** | An AI agent specialized in one domain (Product, GTM, Fundraising, Pitch). Reasons over the startup's profile using its Playbooks. |
| **Office Hours** | A working session between a founder and an Advisor. |
| **Dossier** | The startup's living profile: its Data Room, session history, and current Signals. An Advisor reads it before every session. |
| **Data Room** | Founder-uploaded materials — pitch deck, business plan, financials. Ingested as evidence, not just stored. |
| **Signals** | The multi-dimensional scores describing the project. Re-evaluated after each session. |
| **Action Items** | Prioritized next steps produced by a session, each linked to a real program, tracked in the Logbook. |
| **Logbook** | The persistent tracking view: current stage, history, action items, progress over time. |
| **Playbooks** | Admin-managed methodology libraries the Advisors draw on (not exposed in the UI). |

---

## The three engines

### 1. Diagnostic — maturity classification

A founder pitches the Advisor; materials can be uploaded up front, later, or mid-conversation. The system collects evidence through an adaptive exchange and places the startup at a stage in a six-stage taxonomy. Every classification is tied to the specific data points behind it.

**Maturity taxonomy (six stages, grouped into three phases):**

| Phase | Stage | Defining question |
|-------|-------|-------------------|
| **Discover** | 1. Ideation | Is there a clearly defined problem worth solving? |
| **Discover** | 2. Market Validation | Is there evidence real customers want this? |
| **Validate** | 3. Structuration | Is the business structured (legal, team, model)? |
| **Validate** | 4. Launch Planning | Is the offer ready to go to market? |
| **Scale** | 5. Fundraising | Is the startup investable on its fundamentals? |
| **Scale** | 6. Growth | Is growth repeatable without linear cost? |

**Stage-Gate Reviews.** A startup advances only by clearing a gate. To move up, its Signals must meet the gate's criteria — so the system can tell a founder who believes they're fundraising-ready that the evidence still places them at Structuration, and exactly what is missing.

**Advisor unlocking.** Advisors unlock as the work justifies them. A founder starts with the Product Advisor (product-discovery Playbooks only); the Go-to-Market Advisor stays locked until the problem is validated. This keeps founders from skipping ahead.

### 2. Signals — explainable multi-dimensional scoring

Signals are organized by the desirability / feasibility / viability stress-test framework. Each Signal decomposes into weighted sub-criteria with a per-criterion contribution and a plain-language explanation.

| Signal | Sub-criteria | Maps to required score |
|--------|--------------|------------------------|
| **Desirability** | Problem-fit defined · customer validation evidence · market size & competition | Market |
| **Feasibility** | Value-proposition clarity · product readiness · tech intensity & barrier to entry · scalability & manual dependency | Commercial Offer · Innovation · Scalability |
| **Viability** | Revenue model & pricing · unit economics · sustainability & SDG alignment | Green |

**Aggregation rule (not a simple average).** Each composite uses a gated aggregation: when a *fundamental* sub-criterion scores below a floor, it caps the composite rather than being averaged away by strong scores elsewhere. This reflects how a weak fundamental (e.g. no customer validation) genuinely blocks a project regardless of polish in other areas. Weights and floors are part of the scoring methodology draft.

**Anomaly detection.** The model flags contradictory profiles — e.g. high claimed traction with no validation evidence, or high scalability with heavy manual dependency.

### 3. Orientation — grounded roadmap & resources

Each detected gap or low Signal becomes an **Action Item** matched to a real resource. Recommendations are retrieved from a curated knowledge base of national and international support and financing programs (APII, BFPME, BTS, Startup Act mechanisms, ANPE, incubators/accelerators, plus AFD / EU / UNDP funding). Every recommendation carries its source — nothing is invented.

Action Items are structured as trackable goals (SMART) that can carry attachments from the Data Room as context, persist in the Logbook, and feed the next session.

---

## Architecture sketch

```
                ┌─────────────────────────────────────┐
                │            Founder (FR / AR)          │
                └───────────────┬──────────────────────┘
                                │  Office Hours session
                ┌───────────────▼──────────────────────┐
                │              Advisor agent            │
                │     (Product · GTM · Fundraising)     │
                └───┬───────────┬───────────┬───────────┘
                    │           │           │
        ┌───────────▼──┐  ┌─────▼──────┐  ┌─▼──────────────┐
        │  Diagnostic  │  │   Signals  │  │   Orientation  │
        │ (maturity +  │  │  (scoring  │  │  (RAG over the │
        │ stage-gates) │  │ + anomaly) │  │ resource base) │
        └───────┬──────┘  └─────┬──────┘  └────────┬───────┘
                │               │                  │
                └───────┬───────┴──────────────────┘
                        │  shared project profile
                ┌───────▼───────────────────────────────┐
                │   Dossier  =  Data Room + Signals +     │
                │              Logbook (history)          │
                └───────┬───────────────────────┬────────┘
                        │                        │
                ┌───────▼────────┐      ┌────────▼─────────┐
                │   Playbooks    │      │  Knowledge base   │
                │ (admin-managed)│      │ (national + intl  │
                │                │      │   programs)       │
                └────────────────┘      └───────────────────┘
```

The differentiator is integration: a diagnostic gap triggers retrieval of relevant programs; a low Signal surfaces targeted Action Items; the Advisor's feedback references the structured outputs rather than answering from general knowledge.

---

## How it's built (MVP)

The brain is **Claude Code in headless mode** — no API key required by default; the
backend reuses the host's logged-in `claude` subscription (Plan A). The Go backend's
job worker execs `claude -p "<prompt>" --output-format json --dangerously-skip-permissions`
per task. The agent's only hands are a single CLI, **`ohctl`**, which talks directly to
Postgres: it reads context (profile, session, RAG) and persists results (signals, goals,
action items, stage, parcours events). Every agent prompt is a markdown file with
frontmatter under `seed/`.

### Components

| Service | What it is |
|---------|------------|
| **db** | `postgres:16` — all state (users, profiles, signals, sessions, messages, goals, action items, documents, chunks/FTS, events, agent_jobs). |
| **api** | Go HTTP backend on `:8080`. Bundles `ohctl` and the `claude` CLI. Runs the job-worker pool that execs Claude per queued job. Applies the schema (`migrations/0001_init.sql`) on boot. |
| **web** | React + Vite + TS, dark UI (YC-orange × Linear), built and served by nginx on `:3000`; nginx proxies `/api/*` → `api:8080`. |

### Data / agent flow

```
Founder → web (React)
        → POST /api/onboarding ───► agent_jobs(diagnoser) ─► claude + ohctl ─► profile.stage + initial Signals
        → Office Hours chat ──────► agent_jobs(advisor)   ─► claude + ohctl rag query (KB) ─► assistant reply
        → Conclude session ───────► agent_jobs(scorer)    ─► claude + ohctl ─► 5 Signals (subscores+rationale),
                                                                               goals, action items, parcours event
        → Dashboard / Goals / Mon Parcours read the persisted state
        → Data Room upload ───────► agent_jobs(rag_index) ─► chunk + tsvector ─► per-user RAG collection
```

See `docs/architecture.md`, `docs/knowledge-base.md`, `docs/scoring-methodology.md`, and
`docs/evaluation.md` for detail.

### Repository layout

```
backend/   Go module "officehours"
  cmd/api/        HTTP backend + job worker
  cmd/ohctl/      the agent's CLI (and the seed/admin tool)
  internal/{db,models,auth,handlers,agent,rag,config,worker}/
  migrations/0001_init.sql
  eval/           diagnoser stage-classification eval
  Dockerfile
web/         Vite + React + TS, nginx.conf, Dockerfile
seed/{advisors,learn,agents,kb,profiles}/   agent prompts + grounding KB
config/features.yaml
docs/        architecture, knowledge-base, scoring-methodology, evaluation
docker-compose.yml  .env.example  BUILD_SPEC.md
```

---

## Running it locally

### Prerequisites

- **Docker** + **Docker Compose** (v2).
- **A logged-in `claude` CLI on the host (macOS).** Plan A reuses your Claude
  subscription. On macOS the subscription token lives in the **login keychain**, not in
  a file — so a helper exports it into a gitignored `./.secrets/claude-credentials.json`
  that compose mounts into the api container. The api runs as a **non-root** user
  (`claude` refuses `--dangerously-skip-permissions` under root). **Before bringing the
  stack up, run on the host:**

  ```bash
  npm install -g @anthropic-ai/claude-code   # if you don't have it
  claude                                     # log in once (interactive)
  ./scripts/sync-claude-creds.sh             # export the token for the container
  ```

  Re-run `sync-claude-creds.sh` if the token expires. Alternatively set
  `ANTHROPIC_API_KEY` in `.env` to use an API key instead of the subscription.

  > Note: agent turns use the full agent (tool use + Opus) and take ~60–120s each;
  > the UI shows a "thinking" state while a job runs.

### 1. Configure environment

```bash
cp .env.example .env
# defaults work as-is for local docker compose; set JWT_SECRET / ANTHROPIC_API_KEY if desired
```

### 2. Bring the stack up

```bash
docker compose up --build
```

This builds and starts `db`, `api`, and `web`. On first boot the api applies
`migrations/0001_init.sql` automatically (idempotent — it skips if the schema already
exists). The web UI is at **http://localhost:3000**, the API at **http://localhost:8080/api**.

### 3. Seed data and create a login

There is **no public registration** — accounts are created with `ohctl`. Seed the
advisors/concepts and index the knowledge base, and optionally create a ready-to-go
demo founder, in one command:

```bash
# Index the KB into its collections, validate agent defs, and create a demo user
# (founder@demo.officehours.ai / demo1234) with the example agritech company text:
docker compose exec api ohctl seed demo --with-user
```

Or create your own login:

```bash
docker compose exec api ohctl seed user --email you@example.com --password yourpass --name "Your Name"
```

(If you don't pass `--with-user`, run `ohctl seed demo` once to index the KB, then
`ohctl seed user ...` to make a login.)

### 4. Run the demo (critical path)

1. Open **http://localhost:3000**, log in (`founder@demo.officehours.ai` / `demo1234`,
   or your seeded user).
2. **Onboarding** — describe the company (the demo user is pre-filled). Submitting
   enqueues the **diagnoser**, which sets the maturity stage and initial Signals. The
   UI shows a thinking state while the job runs, then refetches.
3. **Office Hours** — pick the **Product Advisor**, start a session, and chat. The
   advisor grounds its answers in the KB via `ohctl rag query`.
4. **Conclude** the session — this enqueues the **scorer**, which updates the 5 Signals
   (each with sub-score breakdowns + rationale), creates **Goals** and **Action Items**
   grounded in real KB programs, and writes a **Mon Parcours** timeline event.
5. **Dashboard** — see the stage, the 5 Signals with sub-score breakdowns, goals, and
   the parcours.
6. **Data Room** — upload a `.md`/`.txt`/`.pdf`; it is chunked and indexed into your
   per-user RAG collection so a later session can cite it.

> The 6 stages: `Ideation`, `Market Validation`, `Structuration`, `Fundraising`,
> `Launch Planning`, `Growth`. The 5 Signals: `Market`, `Commercial Offer`,
> `Innovation`, `Scalability`, `Green`.

### 5. Run the evaluation

The eval feeds a labeled set of ~10 founder profiles through the diagnoser and reports
stage-classification accuracy. It needs the `claude` CLI (logged in) and `ohctl` on
PATH, plus a reachable Postgres. Run it from the host against the compose db:

```bash
cd backend
go build -o ./bin/ohctl ./cmd/ohctl
DATABASE_URL='postgres://officehours:officehours@localhost:5432/officehours?sslmode=disable' \
  go run ./eval -set eval/labeled_set.json -seed-dir ../seed -ohctl-dir ./bin
```

It prints a JSON report (per-case predicted vs expected stage + overall accuracy).
Method and results are written up in `docs/evaluation.md`.

### Testing

Integration / E2E suites live in `tests/` and run against the **live docker
compose stack** (bring it up first). They are idempotent and rerunnable, and
each isolates its fixtures so the suites don't interfere with demo data:

```bash
./tests/cli.sh    # ohctl CLI integration (23 cases)
./tests/api.sh    # HTTP API integration (45 assertions)
./tests/rag.sh    # RAG retrieval E2E (5 cases)
```

See `tests/README.md` for what each suite asserts and prerequisites (`jq`,
`curl`). The full **agent pipeline** (diagnoser/advisor/scorer) smoke is
**manual and Claude-dependent** — there is no automated script for it; the
procedure and current blocker are documented in `tests/README.md`.

---

## Non-functional notes

- **Reliability** — the api boots and serves `/api/health` even before the db is warm;
  the schema is applied idempotently on connect. PDF extraction failures during RAG
  indexing are logged and skipped rather than crashing.
- **Privacy** — founder data lives only in your local Postgres volume; Plan A uses your
  own Claude subscription, no third-party key required.
- **Scalability** — the KB is re-indexable without rebuilding the app (`ohctl seed demo`
  / `ohctl rag index`); new Advisors and concepts are added as new markdown files in
  `seed/advisors` and `seed/learn`.

## Submission docs

- `docs/architecture.md` — components, data flow, the agent/`ohctl`/`claude` pipeline.
- `docs/knowledge-base.md` — sources, formats, key fields, the FTS ingestion pipeline.
- `docs/scoring-methodology.md` — Signal criteria, weights, gated aggregation.
- `docs/evaluation.md` — eval method + results.
- `design/deck-final.html` — the pitch deck.
- `BUILD_SPEC.md` — the frozen build contract.
