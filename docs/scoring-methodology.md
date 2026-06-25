# Scoring Methodology (Draft)

This document defines how OfficeHours.ai turns a startup's profile into **Signals** — explainable, multi-dimensional scores. It is a Round 1 draft: criteria and weights are initial proposals to be calibrated against a labeled test set.

## Principles

1. **Every score is traceable.** A composite decomposes into sub-criteria; each sub-criterion contributes a visible amount; each contribution traces back to specific evidence in the Dossier.
2. **Composites are not averages.** A weak fundamental should not be hidden by strength elsewhere (see *gated aggregation*).
3. **Uncertainty is surfaced.** When evidence for a sub-criterion is missing, the score reflects low confidence rather than a confident guess.

## Signal structure

Signals are grouped by the desirability / feasibility / viability stress-test framework. The five composite scores required by the brief (Market, Commercial Offer, Innovation, Scalability, Green) are expressed as sub-criteria within these three lenses.

### Desirability — *do customers want this?*

| Sub-criterion | Weight | Evidence used |
|---------------|:------:|---------------|
| Problem-fit defined | 0.30 | Problem statement clarity, target segment specificity |
| Customer validation evidence | 0.45 | Interviews, LOIs, paying users, retention/churn |
| Market size & competition | 0.25 | TAM/SAM estimate, competitive landscape |

*Fundamental sub-criterion: Customer validation evidence.*

### Feasibility — *can it be built and delivered?*

| Sub-criterion | Weight | Evidence used |
|---------------|:------:|---------------|
| Value-proposition clarity & differentiation | 0.25 | Offer description, offer↔need alignment |
| Product / service readiness | 0.25 | Maturity of the product, demoability |
| Technology intensity & barrier to entry | 0.20 | Defensibility, IP, technical depth |
| Scalability & manual dependency | 0.30 | Replicability without linear cost, deployment cost |

*Fundamental sub-criterion: Scalability & manual dependency.*

### Viability — *does the business work?*

| Sub-criterion | Weight | Evidence used |
|---------------|:------:|---------------|
| Revenue model & pricing coherence | 0.40 | Pricing strategy, revenue model clarity |
| Unit economics | 0.35 | Margin, CAC/LTV signals where available |
| Sustainability & SDG alignment (Green) | 0.25 | Environmental impact, resource efficiency, SDG fit |

*Fundamental sub-criterion: Revenue model & pricing coherence.*

## Sub-criterion scoring

Each sub-criterion is scored on a **0–5** scale by a rubric (anchored descriptions per level), with a separate **confidence** value (0–1) reflecting how much evidence supports it. Sub-criterion score and confidence are both shown to the founder.

## Gated aggregation

A composite is **not** a weighted mean. The procedure:

1. Compute the weighted mean `m` of the sub-criterion scores.
2. If the **fundamental** sub-criterion for that lens scores below a floor `F` (default `F = 2.0`), cap the composite at a ceiling that scales with the fundamental:
   `composite = min(m, fundamental_score + 0.5)`.
3. Otherwise `composite = m`.

This guarantees that, for example, a project with strong tech and pricing but **no customer validation** cannot present as desirable — the missing fundamental caps the score and the explanation names it.

## Natural-language justification

For each composite, the system emits a short explanation naming: the top positive contributor, the largest gap, and whether a fundamental floor was triggered. Example:

> *Desirability: 2.4 / 5. Capped by weak customer validation (1.0) — strong market sizing (4.0) cannot offset a missing fundamental. Highest-leverage action: collect 5 customer interviews or one signed LOI.*

## Anomaly detection

Contradiction checks run across Signals and flag profiles such as:

- High claimed traction with no validation evidence.
- High scalability score with high manual-dependency evidence.
- Fundraising self-assessment with unviable unit economics.

## Improvement guidance

For each composite, the lowest-scoring high-weight sub-criterion is identified as the highest-leverage gap, and a concrete Action Item is generated and matched to a relevant program in the knowledge base.

## Calibration & evaluation (planned)

- Build a labeled set of startup profiles with expert-assigned stage and rough score bands.
- Tune weights and the floor `F` to maximize maturity-stage classification accuracy and agreement with expert score bands.
- Report Signal consistency: the same profile entered twice should produce stable composites.
