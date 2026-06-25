# Knowledge Base

The knowledge base (KB) is the **grounding corpus** for every agent in
OfficeHours.ai. Advisors, concept tutors, and the scorer do not answer from the
model's parametric memory alone — they retrieve relevant KB passages via
`ohctl rag query` and cite real programs. This document describes the sources,
the file format and key fields, the full-text ingestion pipeline, and coverage.

It is written against the frozen contract in [`BUILD_SPEC.md`](../BUILD_SPEC.md)
(§3 RAG, §5 agent definitions, §8 seed data). Names and behaviours here match
that spec exactly.

## 1. Sources

The corpus lives under `seed/kb/<collection>/*.md` and contains **≥30 real
programs total** spread across topic collections. The programs are the Tunisian
and international support/financing landscape a founder actually has to navigate:

- **National financing & support**: APII (Agence de Promotion de l'Industrie et
  de l'Innovation), BFPME (Banque de Financement des PME), BTS (Banque
  Tunisienne de Solidarité), the **Startup Act** label and its benefits, ANPE,
  and related public instruments.
- **Incubators & accelerators**: Tunisian and regional programs that take
  early-stage teams (cohorts, equity-free grants, mentorship).
- **International funding**: AFD (Agence Française de Développement), EU
  instruments, UNDP and other development-finance programs available to
  Tunisia/MENA founders.

Each program is one markdown file so it can be cited atomically and updated
without touching its neighbours.

## 2. Collections

Documents and chunks are partitioned by a `collection` string. There are two
classes of collection:

| Class | `user_id` | Examples | Who reads it |
|-------|-----------|----------|--------------|
| **Shared KB** | `null` | `kb-fundraising`, `kb-product`, `kb-tunisia`, `kb-gtm`, … | All users. Advisors/concept tutors query the collection named in their frontmatter. The scorer queries KB to match Action Items to programs. |
| **Per-user Data Room** | a user UUID | the founder's uploaded deck/BP/financials | Only that founder's sessions. Used to cite the founder's own documents. |

The collection a given agent may query is fixed in its markdown frontmatter
(`collection: kb-product` for the Product Advisor, etc.). The same FTS pipeline
indexes both classes; the only difference is whether `user_id` is set.

## 3. File format & key fields

### Program KB entries (`seed/kb/<collection>/*.md`)

Each program file is plain markdown carrying, at minimum, the fields the spec
requires for grounding:

- **title** — the program's name.
- **what it is** — a one/two-line description of the instrument.
- **who qualifies** — eligibility (stage, sector, legal form, geography).
- **how to apply** — the practical application path.
- **source URL** — the authoritative link, so a citation is verifiable.

These five fields are what an Action Item's `program_ref` points at: when the
scorer turns a gap into an Action Item, it names the program and attaches the
source so the advice is traceable.

### Agent definitions (`seed/{advisors,learn,agents}/*.md`)

Agents are markdown with YAML frontmatter (BUILD_SPEC §5):

```
---
key: product
name: Product Advisor
kind: advisor            # advisor | concept | system
collection: kb-product   # KB collection this agent may query
enabled: true
order: 1
description: Sharpens problem definition and product discovery.
---
<system prompt body…>
```

- `kind: advisor` — Office Hours specialists: `product`, `gtm`, `fundraising`,
  `pitch`.
- `kind: concept` — Learn tutors: `lean-startup`, `investment`, `gtm-basics`
  (≥3), surfaced when `learn.enabled` in `config/features.yaml`.
- `kind: system` — `diagnoser` and `scorer` under `seed/agents/`.

The `collection` field is the link between an agent and the KB: it is the exact
collection the agent passes to `ohctl rag query`.

## 4. The FTS ingestion pipeline

OfficeHours.ai uses **PostgreSQL full-text search** (no external vector store).
Ingestion runs through `ohctl rag index <folder> --collection <name> [--user
<uuid>]`. `ohctl seed demo` indexes every `seed/kb/*` folder into its matching
collection.

The pipeline, per file:

1. **Read & accept.** Read `.md` / `.txt` directly. For `.pdf`, extract text via
   a Go PDF library (`github.com/ledongthuc/pdf`); on extraction failure, **log a
   warning and skip** the file rather than aborting the batch. Accepted types
   for the Data Room are configured in `config/features.yaml`
   (`data_room.accept: [".md", ".txt", ".pdf"]`).
2. **Create the document row.** Insert into `documents`
   (`user_id`, `collection`, `filename`, `mime`, `path`). `user_id` is `null`
   for shared KB, set for a Data Room upload.
3. **Chunk.** Split the text into **~800-character, paragraph-aware** chunks so a
   retrieved passage is coherent (the splitter respects paragraph boundaries
   instead of cutting mid-sentence).
4. **Store chunks with a tsvector.** Insert each chunk into `chunks`
   (`document_id`, `collection`, `user_id`, `ord`, `content`) and set
   `tsv = to_tsvector('english', content)`.
5. **Index.** A GIN index `chunks_tsv_idx` over `tsv` (plus `chunks_collection_idx`
   on `collection`) keeps retrieval fast.

### Retrieval (`ohctl rag query`)

```
ohctl rag query --collection <name> "<q>" [--k 5] [--user <uuid>]
```

The query is:

```sql
SELECT content
FROM chunks
WHERE collection = $collection
  [AND user_id = $user]            -- when --user is given
  AND tsv @@ plainto_tsquery('english', $q)
ORDER BY ts_rank(tsv, plainto_tsquery('english', $q)) DESC
LIMIT $k;                          -- default k = 5
```

Results are returned as JSON (like every `ohctl` command), so the agent can
read them from stdout and quote/cite them. Filtering by `collection` keeps an
advisor scoped to its domain; adding `--user` scopes retrieval to a founder's
own Data Room.

### Ingestion flow

```
seed/kb/<collection>/*.md ──┐
Data Room upload (POST) ─────┤
                            ▼
                  ohctl rag index <folder> --collection <name> [--user]
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
       read/accept      chunk (~800c,    to_tsvector('english')
       (.md/.txt/.pdf)  paragraph-aware)        │
            │               │                   ▼
            └──── documents ─┴──── chunks (tsv, gin index) ── Postgres
                                          ▲
                                          │ tsv @@ plainto_tsquery + ts_rank
                          ohctl rag query --collection … "<q>" --k 5
                                          │
                                     agent (Claude) cites programs
```

## 5. Coverage notes

- **Breadth.** ≥30 real programs across the financing/support landscape
  (national: APII, BFPME, BTS, Startup Act, ANPE; ecosystem: incubators &
  accelerators; international: AFD, EU, UNDP). This is enough to ground every
  advisor domain and to give the scorer concrete programs to attach to Action
  Items.
- **Grounding guarantee.** Because Action Items carry a `program_ref` back to a
  KB file (with its source URL), advice is checkable rather than hallucinated.
- **Per-user augmentation.** A founder's Data Room (deck, BP, financials) is
  indexed into a per-user collection, so a session can cite the founder's own
  evidence alongside public programs.
- **Known limitations.**
  - FTS is lexical, not semantic — synonyms/paraphrase can miss. The ~800-char
    paragraph-aware chunking and `plainto_tsquery` mitigate this for the MVP;
    embeddings/hybrid search are a post-MVP step (see
    [`evaluation.md`](evaluation.md) and `next steps` in the deck).
  - The English text-search configuration is used for all content; French-only
    program text still indexes but ranks less precisely.
  - Programs change (deadlines, amounts). KB files are point-in-time snapshots
    with source URLs; keeping them current is an operational task, not a code
    one.
