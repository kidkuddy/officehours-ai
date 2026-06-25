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
  updated_at timestamptz default now(),
  onboarded_at timestamptz                         -- set once on first onboarding submit
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
