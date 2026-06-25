# Explainability Layer

OfficeHours.ai is built so that **every output is traceable**: a founder (and a judge) can see *why* the system produced a given maturity stage, *what evidence* supports each score, and *how* each number was computed. Explainability is not a panel bolted on at the end — it is the shape of the data the engines produce.

## 1. Diagnosis is evidence-linked, not a verdict

The diagnoser never returns a bare stage. It writes the stage **with the specific data points** that justify it, persisted on the profile (`stage_evidence`). Example for an agri-tech founder who *claimed* "ready to raise a seed round":

| Field | Value |
|---|---|
| Self-declared intent | Fundraising-ready |
| System stage | **Market Validation** |
| Evidence | "3 pilot farms but no signed contracts"; "no revenue yet"; "no validated pricing"; "IoT sensor built (product exists)"; "no repeatable acquisition channel" |

The gap between the claim and the evidence is visible, and the founder can see exactly which observations moved them down the ladder.

## 2. Signals decompose into weighted, contribution-level detail

Each of the 5 composite Signals (Market, Commercial Offer, Innovation, Scalability, Green) is stored with its **sub-criteria, each with a weight and a contribution**, plus a plain-language rationale and a `floor_triggered` flag. The UI surfaces this progressively: the founder first sees the three stress-test frameworks (Desirability / Feasibility / Viability), expands one to its composite Signals, then expands a Signal to its sub-criteria bars and rationale.

Example — the **Market** Signal:

```
Market  2.0 / 5   ⚑ CAPPED
  ├─ Customer validation evidence   weight 0.45   score 1.0   contribution 0.45   ← fundamental, below floor
  ├─ Market size & competition      weight 0.25   score 4.0   contribution 1.00
  └─ Revenue model clarity          weight 0.30   score 2.0   contribution 0.60
  rationale: "Strong market sizing cannot offset missing customer validation —
              the fundamental is below the floor, so the composite is capped."
```

The number `2.0` is not asserted; it is *derived and shown*: the weighted mean is gated by the weakest fundamental (see §3), and the contribution of every criterion is on screen.

## 3. The scoring math is transparent and defensible

Composite scores are **not simple averages** (full method in [`scoring-methodology.md`](scoring-methodology.md)). Each lens has a designated *fundamental* sub-criterion; when it scores below a floor `F = 2.0`, the composite is capped:

```
composite = min(weighted_mean, fundamental_score + 0.5)
```

This encodes a real domain truth — a startup with no customer validation cannot present as "desirable" no matter how polished its deck — and the UI shows the `CAPPED` badge plus the rationale naming the limiting criterion, so the cap is never a black box.

## 4. Recommendations are grounded, never invented

Every Action Item is matched to a **real resource retrieved from the knowledge base** (`program_ref`), and the conversational advisors are instructed to name programs only from `ohctl rag query` results — "never name a program from memory; if retrieval returns nothing, say so." A recommendation that cannot be traced to a knowledge-base item or a diagnostic output is treated as a failure, not output. (KB sources and ingestion in [`knowledge-base.md`](knowledge-base.md).)

## 5. Where to see it in the product

- **Dashboard → Signals**: the three frameworks → composites → sub-criteria contribution bars + rationale + `CAPPED` badge (click to expand each level).
- **Profile**: maturity stage with the evidence behind it.
- **Office Hours**: advisor replies reference the founder's actual Signals/stage and cite KB programs by name.
- **Logbook**: every stage change and signal update is recorded as a timestamped event.

## 6. Uncertainty is surfaced, not hidden

Incomplete profiles produce lower-confidence sub-scores rather than confident guesses, and the diagnoser states what evidence is missing. The system tells a founder what it *doesn't* yet know about their company instead of inventing a value.
