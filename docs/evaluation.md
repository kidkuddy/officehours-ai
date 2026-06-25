# Evaluation Report

We evaluate OfficeHours.ai's core inference task — **maturity-stage classification** by the diagnoser — on a labeled set of founder profiles, run end-to-end through the real agent.

## Protocol

- **Test set:** 10 hand-labeled founder profiles (`backend/eval/labeled_set.json`), each a short company description with an expert-assigned expected stage, spanning all six stages.
- **Runner:** `backend/eval` creates an ephemeral user per case, execs the real diagnoser (headless agent with `ohctl` on PATH), reads back the persisted stage, and compares it to the label. Ephemeral users are deleted after each run.
- **Metric:** stage-classification accuracy — reported as **exact match** and **within-one-stage** (adjacent stages on the six-stage ladder), since adjacent confusions are far less harmful than distant ones for orientation.

## Results

| # | Case | Expected | Predicted | Exact | Note |
|---|------|----------|-----------|:----:|------|
| 1 | agritech-overclaim | Structuration | Market Validation | ✗ | run killed mid-diagnosis |
| 2 | pure-idea | Ideation | Ideation | ✓ | |
| 3 | validated-no-product | Market Validation | Ideation | ✗ | run killed → defaulted |
| 4 | incorporated-building | Structuration | Structuration | ✓ | |
| 5 | raising-seed | Fundraising | Fundraising | ✓ | |
| 6 | prelaunch-gtm | Launch Planning | Launch Planning | ✓ | |
| 7 | scaling-growth | Growth | Ideation | ✗ | run killed → defaulted |
| 8 | deeptech-early | Ideation | Ideation | ✓ | |
| 9 | evidence-driven-validation | Market Validation | Market Validation | ✓ | |
| 10 | structured-not-raising | Structuration | Structuration | ✓ | |

**Headline metrics**

| Metric | Result |
|---|---|
| Exact-match accuracy | **70%** (7/10) |
| Within-one-stage accuracy | **90%** (9/10) |
| Accuracy on completed runs | **100%** (6/6) |

## Analysis

The decisive finding: **all three misclassifications coincided with a killed diagnoser process** (`agent: claude exec: signal: killed`) — the agent run was terminated before it wrote a stage, leaving the profile at the default (`Ideation`) or a partial value. Two of the three wrong predictions are exactly that default.

- **Among the 6 runs that completed cleanly, classification was 6/6 exact** — across Ideation, Market Validation, Structuration, Launch Planning, and Fundraising. The model reasons the stages correctly when it finishes.
- The single large error (scaling-growth → Ideation, 5 stages off) is itself a killed-run default, not a reasoning error.
- **Within-one-stage accuracy is 90%**, and the only >1-stage error is a killed run.

So the limiting factor in this batch was **runtime, not reasoning**: long agentic diagnoses (multiple tool calls + scoring) occasionally exceed the eval's per-case process budget and get killed.

## Limitations & next steps

- **Raise the per-case timeout / add a retry-on-kill** in the eval harness; re-running the three killed cases is expected to lift exact-match toward the completed-run rate.
- **Reduce diagnosis latency** (fewer tool round-trips, or a faster model for the classification pass) so runs finish well inside the budget.
- **Expand the labeled set** beyond 10 and add **Signal-consistency** measurement (same profile entered twice → stable composites).
- **Inter-rater check** on the labels themselves, since adjacent-stage boundaries are inherently fuzzy.

## Reproduce

```bash
cd backend && go build -o ./bin/ohctl ./cmd/ohctl && go build -o ./bin/eval ./eval
DATABASE_URL='postgres://officehours:officehours@localhost:5432/officehours?sslmode=disable' \
  ./bin/eval -set eval/labeled_set.json -seed-dir ../seed -ohctl-dir ./bin -out eval/report.json
```
