# OfficeHours.ai — BUILD SPEC (frozen contract)

This is the single source of truth for the MVP build. Every agent builds against this. Do not invent alternative schemas, routes, or formats — if something is missing, follow the conventions here.

## 0. What we're building

A founder logs in, describes their company once (onboarding), and the system stands up an evidence-based profile. The founder then holds **Office Hours** chat sessions with specialist **Advisors** (Claude Code agents). A session can be **concluded**, which runs a **scorer** agent that updates the founder's **Signals** (5 composite scores), creates **Goals** and **Action Items**, and appends to a **Mon Parcours** timeline. Files uploaded to the **Data Room** are indexed (RAG, full-text) into a per-user collection. **Learn** offers concept tutors (also agents) grounded in topic KB collections.

The AI is **Claude Code in headless mode** (Plan A: reuse the host's logged-in Claude subscription, no API key). The backend's job worker execs `claude -p`. The agent's hands are a single CLI binary, **`ohctl`**, which talks directly to Postgres.

## 1. Services (docker-compose)

- `db` — `postgres:16`, volume `dbdata`, healthcheck, `POSTGRES_*` from `.env`.
- `api` — Go backend (HTTP) + bundled `ohctl` + the `claude` CLI. Mounts `${HOME}/.claude:/root/.claude:ro` (Plan A creds) and a `uploads` volume. Env: `DATABASE_URL`, `JWT_SECRET`, `ANTHROPIC_API_KEY` (optional fallback), `UPLOAD_DIR=/data/uploads`, `SEED_DIR=/seed`, `CONFIG_DIR=/config`. Mounts `./seed:/seed:ro` and `./config:/config:ro`.
- `web` — React (Vite + TS) built and served by nginx; nginx proxies `/api/*` → `api:8080`.

`.env.example` must list every variable. Default ports: api `8080`, web `3000`, db `5432`.

### Claude exec (the brain)
The api image installs Node + `@anthropic-ai/claude-code` so `claude` is on PATH. The job worker runs, per task, in a temp working dir with `ohctl` on PATH:
```
claude -p "<rendered prompt>" --output-format json --dangerously-skip-permissions --add-dir <workdir>
```
Auth: mounted `/root/.claude` (subscription). If `ANTHROPIC_API_KEY` is set, that is used instead. The prompt tells the agent it has `ohctl` and must use it to read context and persist results. The assistant's final text is the chat reply (captured and stored). README must document: "run `claude` login on the host first."

## 2. Database schema (Postgres) — `backend/migrations/0001_init.sql`

Use `gen_random_uuid()` (enable `pgcrypto`). Timestamps `timestamptz default now()`.

```sql
create extension if not exists pgcrypto;

create table users (
  id uuid primary key default gen_random_uuid(),
  email text unique not null,
  password_hash text not null,
  name text not null,
  created_at timestamptz default now()
);

create table profiles (
  user_id uuid primary key references users(id) on delete cascade,
  company_text text not null default '',
  stage text not null default 'Ideation',          -- one of the 6 stages
  stage_evidence jsonb not null default '[]',
  created_at timestamptz default now(),
  updated_at timestamptz default now()
);

-- 5 composite Signals per user (Market, Commercial Offer, Innovation, Scalability, Green)
create table signals (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references users(id) on delete cascade,
  name text not null,
  score numeric(3,1) not null default 0,            -- 0.0 .. 5.0
  subscores jsonb not null default '[]',            -- [{criterion,weight,score,contribution}]
  rationale text not null default '',
  floor_triggered boolean not null default false,
  updated_at timestamptz default now(),
  unique(user_id, name)
);

create table sessions (
  id uuid primary key default gen_random_uuid(),    -- resumable by this UUID
  user_id uuid references users(id) on delete cascade,
  kind text not null default 'office_hours',        -- 'office_hours' | 'learn'
  advisor_key text not null,                         -- advisor or concept key
  title text not null default '',
  status text not null default 'active',             -- 'active' | 'concluded'
  outcomes text not null default '',                 -- written on conclude
  created_at timestamptz default now(),
  concluded_at timestamptz
);

create table messages (
  id uuid primary key default gen_random_uuid(),
  session_id uuid references sessions(id) on delete cascade,
  role text not null,                                -- 'user' | 'assistant' | 'system'
  content text not null,
  created_at timestamptz default now()
);

create table goals (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references users(id) on delete cascade,
  title text not null,
  description text not null default '',
  status text not null default 'open',               -- 'open' | 'done'
  source_session_id uuid references sessions(id),
  created_at timestamptz default now(),
  done_at timestamptz
);

create table action_items (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references users(id) on delete cascade,
  session_id uuid references sessions(id),
  title text not null,
  horizon text not null default 'short',             -- 'immediate'|'short'|'medium'
  rationale text not null default '',
  program_ref text not null default '',              -- KB source for grounding
  status text not null default 'open',
  created_at timestamptz default now()
);

create table documents (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references users(id) on delete cascade,  -- null => shared KB
  collection text not null,
  filename text not null,
  mime text not null default '',
  path text not null default '',
  created_at timestamptz default now()
);

create table chunks (
  id uuid primary key default gen_random_uuid(),
  document_id uuid references documents(id) on delete cascade,
  collection text not null,
  user_id uuid,                                      -- null for shared KB
  ord int not null default 0,
  content text not null,
  tsv tsvector
);
create index chunks_tsv_idx on chunks using gin(tsv);
create index chunks_collection_idx on chunks(collection);

create table events (                                -- Mon Parcours timeline
  id uuid primary key default gen_random_uuid(),
  user_id uuid references users(id) on delete cascade,
  kind text not null,                                -- 'stage_change'|'signal_update'|'session'|'goal'|...
  payload jsonb not null default '{}',
  created_at timestamptz default now()
);

create table agent_jobs (
  id uuid primary key default gen_random_uuid(),
  type text not null,                                -- 'advisor'|'diagnoser'|'scorer'|'learn'
  user_id uuid references users(id) on delete cascade,
  session_id uuid references sessions(id),
  status text not null default 'queued',             -- 'queued'|'running'|'done'|'error'
  input jsonb not null default '{}',
  output jsonb not null default '{}',
  error text not null default '',
  created_at timestamptz default now(),
  started_at timestamptz,
  finished_at timestamptz
);
```

The 6 maturity stages (exact strings): `Ideation`, `Market Validation`, `Structuration`, `Fundraising`, `Launch Planning`, `Growth`.
The 5 Signal names (exact strings): `Market`, `Commercial Offer`, `Innovation`, `Scalability`, `Green`.

## 3. `ohctl` — the agent's CLI (`backend/cmd/ohctl`, cobra)

Connects to `DATABASE_URL`. **Every command prints JSON to stdout.** Used by agents (via Bash) and by humans/seed.

```
ohctl seed user      --email --password --name            # create a login
ohctl seed demo      [--with-user]                         # advisors+learn from /seed md, index /seed/kb, optional demo user
ohctl rag index <folder> --collection <name> [--user <uuid>]   # chunk md/txt/pdf, store + tsv
ohctl rag query  --collection <name> "<q>" [--k 5] [--user <uuid>]
ohctl profile get        --user <uuid>
ohctl profile set-stage  --user <uuid> --stage "<stage>" --evidence '<json-array>'
ohctl signal set         --user <uuid> --name "Market" --score 2.4 --subscores '<json>' --rationale "..." [--floor]
ohctl signal list        --user <uuid>
ohctl goal create        --user <uuid> --title "..." [--desc "..."] [--session <uuid>]
ohctl goal done          --id <uuid>
ohctl goal list          --user <uuid>
ohctl action-item create --user <uuid> --session <uuid> --title "..." --horizon short --rationale "..." --program-ref "..."
ohctl session get        <uuid>                            # meta + messages + action items
ohctl session message    <uuid> --role assistant --content "..."
ohctl session conclude   <uuid> --outcomes "..."
ohctl event add          --user <uuid> --kind "..." --payload '<json>'
```

RAG indexing: split files into ~800-char chunks (paragraph-aware), `tsv = to_tsvector('english', content)`. Query: `tsv @@ plainto_tsquery('english', $q)` ordered by `ts_rank` desc, limit k; filter by collection (and user if given). PDF text via a Go lib (e.g. `github.com/ledongthuc/pdf`); on failure, skip with a logged warning.

## 4. HTTP API (`backend/cmd/api`) — base path `/api`

Auth = JWT in `Authorization: Bearer` (or httpOnly cookie). No public register endpoint — accounts are created via `ohctl seed user`.

```
POST /api/auth/login            {email,password} -> {token, user}
GET  /api/me                    -> {user, profile}
POST /api/onboarding            {text} -> creates/updates profile, enqueues diagnoser job -> {job_id}
GET  /api/profile               -> profile + signals (preview)
GET  /api/signals               -> [signals]
GET  /api/dashboard             -> {stage, signals, stats, parcours:[events]}
GET  /api/goals                 -> [goals]
GET  /api/advisors              -> [{key,name,description}]  (from /seed advisor md, enabled)
GET  /api/learn                 -> [{key,name,description}]  (from /seed learn md + features.yaml)
POST /api/sessions              {advisor_key, kind} -> {id(uuid), ...}
GET  /api/sessions              [?kind=] -> [sessions]
GET  /api/sessions/:id          -> {session, messages, action_items}
POST /api/sessions/:id/messages {content} -> appends user msg, enqueues advisor/learn job -> {job_id}
POST /api/sessions/:id/conclude -> enqueues scorer job -> {job_id}
POST /api/documents             multipart(file) -> stores file, enqueues rag index -> {document}
GET  /api/documents             -> [documents]
GET  /api/jobs/:id              -> {status, output, error}
```
Job worker: a goroutine pool polling `agent_jobs` (queued→running→done/error). Each job type renders its prompt and execs claude (see §1), with the agent using `ohctl`. The advisor/learn job stores the assistant reply as a message. Default goal "Start an Office Hours session" is created at onboarding; the scorer may mark it done.

## 5. Agent definitions — markdown with frontmatter (`/seed/advisors/*.md`, `/seed/learn/*.md`)

```
---
key: product
name: Product Advisor
kind: advisor            # advisor | concept
collection: kb-product   # KB collection this agent may query via ohctl rag query
enabled: true
order: 1
description: Sharpens problem definition and product discovery.
---
<system prompt body, markdown. Tells the agent its role, that it has `ohctl`,
which subcommands to use, to ground answers in `ohctl rag query`, and (for the
scorer pass) how to set the 5 Signals with subscores + rationale, create goals
and action items, and write a stage with evidence.>
```
Advisors to seed: `product`, `gtm` (Go-to-Market), `fundraising`, `pitch`. Concepts to seed (Learn): `lean-startup`, `investment`, `gtm-basics` (at least 3). All `enabled: true`.

The **diagnoser** and **scorer** are agents too, defined at `/seed/agents/diagnoser.md` and `/seed/agents/scorer.md` (kind: `system`). The scorer must produce the 5 composite Signals using the methodology in `docs/scoring-methodology.md` (gated aggregation — a weak fundamental caps the composite). It writes everything via `ohctl`.

## 6. Config (`/config/features.yaml`)

```yaml
learn:
  enabled: true
data_room:
  enabled: true
  accept: [".md", ".txt", ".pdf"]
advisors:
  unlock_all: true        # MVP: all advisors available
```

## 7. Frontend (`web/`, React + Vite + TS, dark)

Pages (react-router): `/` landing, `/login`, `/register` (static page: "Accounts are created via `ohctl seed user --email ... --password ... --name ...`"), `/onboarding`, `/profile` (preview only), `/office-hours` (advisor picker + session history list), `/office-hours/:id` (chat), `/data-room`, `/learn`, `/learn/:key` (chat), `/goals`, `/dashboard`.

Chat: markdown rendering (tables/code), file upload control, polls `GET /jobs/:id` then refetches messages. "Conclude session" button. Show a thinking state while a job runs.

### Design tokens (dark, YC × Linear) — `web/src/styles/tokens.css`
```
--bg:#0a0a0b; --surface:#131316; --surface-2:#1a1a1f; --border:#26262c;
--text:#ededf0; --muted:#8b8b94; --accent:#FB651E; --accent-soft:#2a1a12;
--radius:10px; --font:'Inter',system-ui,sans-serif;
```
Linear-like structure (sidebar nav, dense, subtle borders, calm surfaces), YC-orange accent and Inter to match the deck. No purple. Keep it clean and intentional.

## 8. Seed data & examples

- `/seed/advisors/*.md`, `/seed/learn/*.md`, `/seed/agents/{diagnoser,scorer}.md` (per §5).
- `/seed/kb/<collection>/*.md` — **≥30 real programs total** across collections (`kb-fundraising`, `kb-product`, `kb-tunisia`, ...): APII, BFPME, BTS, Startup Act, ANPE, incubators/accelerators, AFD/EU/UNDP funding. Each file: title, what it is, who qualifies, how to apply, source URL. These are the grounding corpus.
- `/seed/profiles/example-agritech.md` — the demo founder (claims fundraising-ready; reality = Structuration).
- `ohctl seed demo` indexes all `/seed/kb/*` into their collections.

## 9. Evaluation (`backend/eval/` or `scripts/eval/`)

A small labeled set (~8–10 profiles with expected stage) + a runner that feeds each through the diagnoser and reports stage-classification accuracy. Output a short `docs/evaluation.md` with method + results.

## 10. Submission docs (`docs/`)

`architecture.md` (components + data flow + the agent/ohctl/claude pipeline + a diagram), `knowledge-base.md` (sources, formats, key fields, ingestion = the FTS pipeline, coverage notes), `scoring-methodology.md` (exists — keep/extend), `evaluation.md`. Plus extend the pitch deck to ≤15 slides at `design/deck-final.html` (reuse `deck.html` styles): add value proposition, demo walkthrough, limitations, next steps.

## 11. Repo layout

```
backend/  (go module "officehours")
  cmd/api/  cmd/ohctl/
  internal/{db,models,auth,handlers,agent,rag,config}/
  migrations/0001_init.sql
  eval/
  Dockerfile
web/        (vite react ts)  Dockerfile  nginx.conf
seed/{advisors,learn,agents,kb,profiles}/
config/features.yaml
docs/
docker-compose.yml  .env.example  README.md  BUILD_SPEC.md
```

## 12. Critical path (must work for the demo video)

onboarding (diagnoser sets stage + initial Signals) → open Office Hours with Product Advisor → chat → **Conclude** (scorer updates the 5 Signals with breakdowns + rationale, creates goals + action items grounded in KB programs, writes a parcours event) → Dashboard shows stage, 5 Signals with sub-score breakdowns, goals, parcours → Data Room upload indexes a file and a session can cite it. Learn and deep Mon Parcours polish are secondary.

## 13. Conventions

- Go: standard layout, `database/sql` + `pgx` or `lib/pq`, no heavy frameworks (chi or net/http). `go build ./...` must pass.
- Everything containerized; `docker compose up` must build. README documents setup + the `claude` login prerequisite + the seed command.
- Don't break the frozen names in §2 (stages, signal names), the `ohctl` surface in §3, or the API in §4 — the other agents depend on them exactly.
