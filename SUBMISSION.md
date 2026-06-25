# AINS Hackathon 2026 — Final Submission Checklist

Tracking the submission form fields. Status: ✅ ready · 🟡 needs packaging/export · 🔴 not done / needs you.

| # | Form field (required) | Status | Source / action |
|---|---|---|---|
| 1 | **Team Leader name** | 🔴 you | Provide your name. |
| 2 | **Team Name** | 🔴 you | Proposed: "The Jew, the Sad, and the Buggy" — confirm (reconsider for a UN/PNUD panel). |
| 3 | **Team Leader Email** | 🔴 you | Provide your email. |
| 4 | **Work Title** | ✅ | **OfficeHours.ai** |
| 5 | **Concept Presentation** (≤10 MB) | ✅ | Round-1 deck — `OfficeHours-AINS-2026.pdf` (441 KB). |
| 6 | **Pitch deck** (.pptx/pdf, ≤100 MB) | 🟡 | Round-2 15-slide deck — `design/deck-final.html` → export to PDF. |
| 7 | **Architecture diagram** (PDF, ≤10 MB) | 🟡 | `docs/architecture.md` → export the components/data-flow diagram to PDF. |
| 8 | **Explainability layer** (PDF, ≤10 MB) | 🟡 | Assemble a short doc: the Signals instrument (per-criterion breakdown + rationale + "Capped" floor) + `docs/scoring-methodology.md` → PDF. |
| 9 | **Evaluation report** (PDF, ≤10 MB) | 🟡 | Run `backend/eval` against `backend/eval/labeled_set.json`, fill results into `docs/evaluation.md` → PDF. |
| 10 | **Demo video** (video, ≤100 MB, ≤5 min) | 🔴 | Record an end-to-end walkthrough. See script below. |
| 11 | **GitHub Repo Link** | 🔴 | Push the full codebase (repo currently tracks only `README.md`). |

## What's left, grouped
- **You provide:** team leader name, team name, email (#1–3).
- **Export to PDF (I can generate):** pitch deck, architecture diagram, explainability layer, evaluation report (#6–9).
- **Run first:** the evaluation (#9) to get real numbers before exporting.
- **Push:** the repo (#11) — excludes `.secrets/`, `seed/dump/`, `imported__*.md` (already gitignored).
- **Record:** the demo video (#10).

## Demo video — 5-min walkthrough script (draft)
1. **The problem (20s):** Tunisian founders over-pitch and chase vanity metrics; <10 of 1,000+ Startup-Act startups hit VC success.
2. **Onboarding (45s):** log in fresh → the full-screen wizard → describe the company → the "Reading your file…" diagnosis → the stage reveal.
3. **The diagnosis (60s):** dashboard — the stage ladder + Signals grouped by Desirability / Feasibility / Viability; expand one to show sub-criteria contributions, the rationale, and a "Capped" floor (explainability).
4. **Office hours (90s):** start a session — the advisor opens first and leads; ask one thing; show a grounded reply citing a real program; Conclude the session.
5. **The payoff (45s):** dashboard updates (Signals re-scored), Goals + Action Items created, Logbook timeline; mention the Tunisian/intl KB + Learn concepts.
6. **Close (20s):** architecture one-liner (Claude/Gemini agents + ohctl + RAG), what's next.
