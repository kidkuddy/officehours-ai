import { useEffect, useMemo, useRef, useState } from "react";
import { api, type Dashboard } from "../api/client";
import { useJobPoll } from "../hooks/useJobPoll";
import { ErrorAlert } from "../components/ui";
import StageStepper from "../components/StageStepper";
import { IconArrowRight, IconArrowLeft, IconCheck, IconSpark } from "../components/Icons";

const PLACEHOLDER = `We're an agritech startup in Tunis building soil-moisture sensors and an app that tells smallholder farmers when to irrigate. We have a working prototype on 12 farms, a pilot LOI from a regional co-op, two co-founders (an agronomist and an embedded engineer), and we're looking to raise a pre-seed round...`;

// What the founder is here for — informs the conversation, not the API call.
const INTENTS = [
  "Raise a round",
  "Find my next milestone",
  "Validate the market",
  "Build the team",
  "Get to revenue",
  "Just exploring",
];

const SECTORS = [
  "Agritech", "Fintech", "Healthtech", "SaaS", "Marketplace",
  "Hardware", "Climate", "AI / ML", "Edtech", "Other",
];

// The calm status sequence shown while the diagnoser reads the file.
const DIAG_STEPS = [
  "Reading your description",
  "Mapping market signals",
  "Inferring your maturity stage",
  "Drafting your profile",
  "Preparing your reveal",
];

type Step = "welcome" | "company" | "facts" | "diagnosing" | "reveal";

export default function Onboarding() {
  const job = useJobPoll();

  const [step, setStep] = useState<Step>("welcome");
  const [text, setText] = useState("");
  const [sector, setSector] = useState<string | null>(null);
  const [intent, setIntent] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [diagIdx, setDiagIdx] = useState(0);
  const [result, setResult] = useState<Dashboard | null>(null);
  const textRef = useRef<HTMLTextAreaElement>(null);

  const stepOrder: Step[] = ["welcome", "company", "facts", "diagnosing", "reveal"];
  const progress = (stepOrder.indexOf(step) / (stepOrder.length - 1)) * 100;

  useEffect(() => {
    if (step === "company") textRef.current?.focus();
  }, [step]);

  // Advance the calm status sequence while diagnosing.
  useEffect(() => {
    if (step !== "diagnosing") return;
    const id = window.setInterval(() => {
      setDiagIdx((i) => Math.min(i + 1, DIAG_STEPS.length - 1));
    }, 14000);
    return () => window.clearInterval(id);
  }, [step]);

  // Compose the description with the optional quick facts appended.
  const composed = useMemo(() => {
    const extras: string[] = [];
    if (sector) extras.push(`Sector: ${sector}.`);
    if (intent) extras.push(`What I'm here for: ${intent}.`);
    return extras.length ? `${text.trim()}\n\n${extras.join(" ")}` : text.trim();
  }, [text, sector, intent]);

  const runDiagnosis = async () => {
    if (!text.trim()) return;
    setError(null);
    setDiagIdx(0);
    setStep("diagnosing");
    try {
      const { job_id } = await api.onboarding(composed);
      await job.poll(job_id);
      setDiagIdx(DIAG_STEPS.length - 1);
      // Pull the freshly diagnosed profile for the reveal.
      const dash = await api.dashboard().catch(() => null);
      setResult(dash);
      setStep("reveal");
    } catch (e) {
      setError(e instanceof Error ? e.message : "The diagnosis didn't finish. Try again.");
      setStep("company");
    }
  };

  const wordCount = text.trim() ? text.trim().split(/\s+/).length : 0;

  return (
    <div className="wizard">
      <div className="wizard-top">
        <div className="seal">
          <span className="seal-mark">OH</span>
          <span>
            <span className="seal-word">
              OfficeHours<span className="tld">.ai</span>
            </span>
            <span className="seal-kicker">First Meeting</span>
          </span>
        </div>
        {step !== "welcome" && step !== "reveal" && (
          <div className="wizard-progress">
            <span className="wp-step">
              {Math.min(stepOrder.indexOf(step) + 1, 3)} / 3
            </span>
            <div className="wp-track">
              <div className="wp-fill" style={{ width: `${Math.max(progress, 8)}%` }} />
            </div>
          </div>
        )}
      </div>

      <div className="wizard-body">
        {/* ---- Welcome ---- */}
        {step === "welcome" && (
          <div className="wizard-step" key="welcome">
            <div className="welcome-portrait">OH</div>
            <div className="wizard-eyebrow">Welcome</div>
            <h1>
              Let's open <em>your file.</em>
            </h1>
            <p className="wizard-lead">
              OfficeHours pairs you with specialist advisors who actually read your situation
              before they speak. First, tell us about your company — it takes a minute, and you
              only do it once. We'll build your profile and hand you straight into the practice.
            </p>
            <button className="btn primary lg" onClick={() => setStep("company")}>
              Begin <IconArrowRight style={{ width: 16, height: 16 }} />
            </button>
          </div>
        )}

        {/* ---- Company ---- */}
        {step === "company" && (
          <div className="wizard-step" key="company">
            <div className="wizard-eyebrow">Your company</div>
            <h1>Tell us what you're building.</h1>
            <p className="wizard-lead">
              In your own words: the product, traction, team, and what you need next. The more
              specific you are, the sharper your diagnosis.
            </p>
            {error && <ErrorAlert message={error} />}
            <textarea
              ref={textRef}
              className="wizard-textarea"
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder={PLACEHOLDER}
              aria-label="Company description"
            />
            <div className="wizard-foot">
              <button className="btn ghost" onClick={() => setStep("welcome")}>
                <IconArrowLeft style={{ width: 16, height: 16 }} /> Back
              </button>
              <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
                <span className="wizard-counter">{wordCount} words</span>
                <button className="btn primary" onClick={() => setStep("facts")} disabled={!text.trim()}>
                  Continue <IconArrowRight style={{ width: 16, height: 16 }} />
                </button>
              </div>
            </div>
          </div>
        )}

        {/* ---- Quick facts ---- */}
        {step === "facts" && (
          <div className="wizard-step" key="facts">
            <div className="wizard-eyebrow">A couple of quick facts</div>
            <h1>What's your world, and why now?</h1>
            <p className="wizard-lead">
              Optional, but it helps the advisor calibrate. Skip anything that doesn't fit.
            </p>

            <div style={{ marginBottom: 40 }}>
              <div className="field"><label>Sector</label></div>
              <div className="chip-grid">
                {SECTORS.map((s) => (
                  <button
                    key={s}
                    className={`chip${sector === s ? " selected" : ""}`}
                    onClick={() => setSector(sector === s ? null : s)}
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <div className="field"><label>What you're here for</label></div>
              <div className="chip-grid">
                {INTENTS.map((s) => (
                  <button
                    key={s}
                    className={`chip${intent === s ? " selected" : ""}`}
                    onClick={() => setIntent(intent === s ? null : s)}
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>

            <div className="wizard-foot">
              <button className="btn ghost" onClick={() => setStep("company")}>
                <IconArrowLeft style={{ width: 16, height: 16 }} /> Back
              </button>
              <button className="btn primary" onClick={() => void runDiagnosis()}>
                Start the diagnosis <IconSpark style={{ width: 16, height: 16 }} />
              </button>
            </div>
          </div>
        )}

        {/* ---- Diagnosing ---- */}
        {step === "diagnosing" && (
          <div className="wizard-step" key="diagnosing">
            <div className="diagnosing">
              <div className="diag-orbit" aria-hidden>
                <div className="ring" />
                <div className="ring r2" />
                <div className="core">OH</div>
              </div>
              <div className="diag-title">Reading your file…</div>
              <div className="diag-steps" role="status" aria-live="polite">
                {DIAG_STEPS.map((label, i) => {
                  const state = i < diagIdx ? "done" : i === diagIdx ? "active" : "";
                  return (
                    <div className={`diag-step ${state}`} key={label}>
                      <span className="ds-mark">
                        {i < diagIdx ? (
                          <IconCheck style={{ width: 10, height: 10 }} />
                        ) : i === diagIdx ? (
                          <span className="spinner" />
                        ) : (
                          i + 1
                        )}
                      </span>
                      {label}
                    </div>
                  );
                })}
              </div>
              <p className="diag-note">
                Your advisor is reading the file and building an evidence-based profile. This
                usually takes a minute or two — no need to wait by the screen.
              </p>
            </div>
          </div>
        )}

        {/* ---- Reveal ---- */}
        {step === "reveal" && (
          <div className="wizard-step wide" key="reveal">
            <div className="reveal-stage">
              <div className="reveal-stagebadge">
                <IconCheck style={{ width: 14, height: 14 }} /> Profile ready
              </div>
              <h1 style={{ marginBottom: 10 }}>
                {result?.stage ? (
                  <>You're at <em>{result.stage}.</em></>
                ) : (
                  <>Your profile is <em>ready.</em></>
                )}
              </h1>
              <p className="wizard-lead" style={{ margin: "0 auto 8px", textAlign: "center" }}>
                Here's where you stand on the founder journey. Inside, you'll hold office hours
                with advisors who build on this — and turn it into goals.
              </p>
            </div>

            {result?.stage && (
              <div className="card pad-lg" style={{ marginBottom: 24 }}>
                <StageStepper current={result.stage} />
              </div>
            )}

            <div style={{ display: "flex", justifyContent: "center" }}>
              <button className="btn primary lg" onClick={() => window.location.assign("/dashboard")}>
                Enter the practice <IconArrowRight style={{ width: 16, height: 16 }} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
