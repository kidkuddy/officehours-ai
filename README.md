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

## Proposed project structure

```
officehours-ai/
├── README.md
├── docs/
│   ├── architecture.md           # system components & data flow
│   ├── scoring-methodology.md     # Signal criteria, weights, aggregation (draft)
│   ├── maturity-taxonomy.md       # six stages, gate criteria
│   └── knowledge-base.md          # sources, fields, ingestion notes
├── design/
│   ├── deck.html                  # concept deck (source)
│   └── OfficeHours-AINS-2026.pdf  # concept deck (export)
├── advisors/                      # advisor agents + playbook bindings
├── engines/
│   ├── diagnostic/                # intake + maturity classification
│   ├── signals/                   # scoring model
│   └── orientation/               # RAG + roadmap generation
└── knowledge-base/                # curated program catalogue + index
```

## Evaluation plan

- **Test set** — a small labeled set of startup profiles with known maturity stages.
- **Metric** — maturity-stage classification accuracy, plus Signal consistency across repeated inputs of the same profile.
- **Edge cases** — incomplete profiles surface uncertainty rather than guessing.

## Non-functional notes

- **Responsiveness** — sessions return feedback within seconds for realistic profile sizes.
- **Reliability** — missing or dirty inputs are handled by surfacing uncertainty, not crashing.
- **Privacy** — founder financials and project data are sensitive; they are masked/anonymized in any shared or evaluation context.
- **Scalability** — the knowledge base is updatable without rebuilding the retrieval index; new Advisors are added as new Playbook bindings.

## Scope

This repository is the **Round 1 concept submission**: problem framing, technical direction, maturity taxonomy, scoring methodology draft, and architecture sketch. It is not yet a working prototype. The concept deck is delivered separately as a PDF.
