# Evaluation

This document defines how we evaluate OfficeHours.ai's core inference task —
**maturity-stage classification** by the diagnoser — and the protocol for
running it. Results are marked **to be filled after running `backend/eval`**.

It is written against the frozen contract in [`BUILD_SPEC.md`](../BUILD_SPEC.md)
§9 (evaluation) and §2 (the 6 stages), and complements
[`scoring-methodology.md`](scoring-methodology.md).

## 1. What we evaluate

The demo-critical, most-objective output is the diagnoser's **stage**
classification: given a founder's company description, predict one of the six
frozen stages —

`Ideation` · `Market Validation` · `Structuration` · `Fundraising` ·
`Launch Planning` · `Growth`

Stage is a single categorical label per profile, which makes it directly
measurable against an expert-assigned ground truth. (The 5 Signals are
multi-dimensional and partly subjective; their evaluation is via *consistency*
and *agreement with expert score bands*, treated as secondary — see §6.)

## 2. The labeled set

A small, hand-labeled set of **~8–10 profiles** lives with the eval runner
(`backend/eval/`). Each item is a realistic founder description plus its
expert-assigned expected stage. The set deliberately spans all six stages and
includes at least one **adversarial** case — the demo founder
(`seed/profiles/example-agritech.md`) who *claims* fundraising-ready while the
reality is `Structuration`. The set is meant to catch exactly the failure mode
the product exists to fix: founders mistaking momentum for progress.

| Field | Meaning |
|-------|---------|
| `id` | stable identifier for the profile |
| `text` | the company description fed to the diagnoser |
| `expected_stage` | expert ground-truth stage (one of the 6) |
| `notes` | optional: why this stage, what makes it tricky |

## 3. Method

The runner feeds each profile through the **diagnoser** exactly as production
does — the same agent definition (`seed/agents/diagnoser.md`), the same
`claude` headless exec, the same `ohctl` surface (BUILD_SPEC §1, §3). This means
the eval measures the real pipeline, not a stub.

For each profile:

1. Create (or reset) an isolated user/profile.
2. Run the diagnoser job over `text`.
3. Read back the predicted stage via `ohctl profile get --user <uuid>`.
4. Compare `predicted_stage` to `expected_stage`.

Primary metric: **stage-classification accuracy** = correct / total.

Secondary diagnostics:

- **Confusion matrix** over the six stages — to see *which* stages get confused
  (e.g. adjacent-stage drift vs. catastrophic jumps).
- **Off-by-one rate** — predictions one stage above/below truth, since adjacent
  stages are genuinely fuzzy and a near miss is less harmful than a far one.
- **Over-claim catch rate** — on the adversarial case(s), did the diagnoser
  resist the founder's self-assessment and land on the real (lower) stage?

## 4. Protocol (how to run)

Prerequisite: the stack is up and `claude` is logged in on the host
(BUILD_SPEC §1). Then run the eval target in `backend/eval/` (see the repo's
README/Makefile for the exact invocation produced by the backend agent). The
runner:

1. Loads the labeled set.
2. For each item, provisions a clean profile and runs the diagnoser.
3. Collects predicted vs. expected stages.
4. Prints accuracy, the confusion matrix, the off-by-one rate, and the
   over-claim catch rate.

Reproducibility notes:

- Use a fresh DB (or per-run users) so prior state never leaks between items.
- LLM output is non-deterministic; report results as the mean over **N runs**
  (default N = 3) and note the spread, rather than a single point estimate.

## 5. Results

> **To be filled after running `backend/eval`.**

| Metric | Value |
|--------|-------|
| Profiles evaluated | _TBD_ |
| Runs (N) | _TBD_ |
| Stage-classification accuracy | _TBD_ |
| Off-by-one rate | _TBD_ |
| Over-claim catch rate (adversarial cases) | _TBD_ |

Confusion matrix (rows = expected, cols = predicted): _TBD._

Qualitative observations (common confusions, failure modes, fixes applied):
_TBD._

## 6. Secondary: Signal evaluation (planned)

Beyond stage accuracy, the scoring model is assessed (per
[`scoring-methodology.md`](scoring-methodology.md)) on:

- **Consistency** — the same profile entered twice should yield stable composite
  Signals (low variance across repeated scorer runs).
- **Agreement with expert score bands** — composites fall within expert-assigned
  rough bands per profile.
- **Fundamental-floor behaviour** — profiles missing a fundamental (e.g. no
  customer validation) must be *capped* and the explanation must *name* the
  missing fundamental, never silently averaged up.

These are tracked as the calibration target for tuning sub-criterion weights and
the floor `F`; their numeric results are also **to be filled after running the
eval** with an expanded labeled set.
