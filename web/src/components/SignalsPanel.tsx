import { useEffect, useId, useState } from "react";
import type { Signal } from "../api/client";
import { IconChevron } from "./Icons";

const MAX = 5;

function band(score: number): "good" | "mid" | "weak" {
  if (score >= 3.5) return "good";
  if (score >= 2) return "mid";
  return "weak";
}

// The three stress-test frameworks and the composite Signals that roll up
// into each. Signal names mirror the backend exactly.
type Framework = {
  key: string;
  name: string;
  blurb: string;
  composites: string[];
};

const FRAMEWORKS: Framework[] = [
  {
    key: "desirability",
    name: "Desirability",
    blurb: "Do people want this — is there a real, reachable market?",
    composites: ["Market"],
  },
  {
    key: "feasibility",
    name: "Feasibility",
    blurb: "Can you build and grow it — technically and operationally?",
    composites: ["Innovation", "Scalability"],
  },
  {
    key: "viability",
    name: "Viability",
    blurb: "Does the business hold up — economically and responsibly?",
    composites: ["Commercial Offer", "Green"],
  },
];

function bandWord(b: "good" | "mid" | "weak"): string {
  return b === "good" ? "Strong" : b === "mid" ? "Developing" : "At risk";
}

// One-line read for a framework, keyed off its critical composite.
function frameworkRead(b: "good" | "mid" | "weak", critical: Signal | null): string {
  if (!critical) return "No composite Signals scored yet.";
  if (b === "good") return `Holding up — ${critical.name.toLowerCase()} is the one to keep ahead of.`;
  if (b === "mid") return `Mixed — ${critical.name.toLowerCase()} is the limiting composite to work next.`;
  return `Under pressure — ${critical.name.toLowerCase()} is dragging this framework down.`;
}

// ---- Composite (an individual Signal) — collapsed to a one-line gauge,
// expands to subscore contribution bars + rationale. ----
function Composite({ signal }: { signal: Signal }) {
  const [open, setOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const panelId = useId();
  const pct = Math.max(0, Math.min(100, (signal.score / MAX) * 100));
  const b = band(signal.score);
  const subs = Array.isArray(signal.subscores) ? signal.subscores : [];

  // Sweep the gauges in once the composite is opened.
  useEffect(() => {
    if (!open) {
      setMounted(false);
      return;
    }
    const id = window.setTimeout(() => setMounted(true), 60);
    return () => window.clearTimeout(id);
  }, [open]);

  return (
    <div className={`composite${open ? " open" : ""}`}>
      <button
        className="composite-head"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-controls={panelId}
      >
        <span className="composite-chev" aria-hidden>
          <IconChevron />
        </span>
        <span className="composite-name">
          {signal.name}
          {signal.floor_triggered && <span className="badge brass">Capped</span>}
        </span>
        <span className={`composite-mini-gauge ${b}`} aria-hidden>
          <span style={{ width: `${pct}%` }} />
        </span>
        <span className="composite-readout">
          <span className="composite-score">{signal.score.toFixed(1)}</span>
          <span className="composite-max">/{MAX}</span>
        </span>
      </button>

      <div className="composite-panel" id={panelId} role="region">
        <div className="composite-panel-inner">
          {subs.length > 0 && (
            <div className="subscores">
              {subs.map((s, i) => {
                const sp = Math.max(0, Math.min(100, (s.score / MAX) * 100));
                return (
                  <div className="subscore" key={`${s.criterion}-${i}`}>
                    <div className="subscore-head">
                      <span className="subscore-name">{s.criterion}</span>
                      <span className="subscore-meta">
                        {typeof s.weight === "number" && (
                          <span className="subscore-weight">w {Math.round(s.weight * 100)}%</span>
                        )}
                        <span className="subscore-val">{s.score.toFixed(1)}</span>
                      </span>
                    </div>
                    <div className="subscore-bar">
                      <span style={{ width: mounted ? `${sp}%` : "0%", transitionDelay: `${i * 50}ms` }} />
                    </div>
                  </div>
                );
              })}
            </div>
          )}
          {signal.rationale && <div className="signal-rationale">{signal.rationale}</div>}
          {!subs.length && !signal.rationale && (
            <div className="composite-empty">No breakdown recorded for this Signal yet.</div>
          )}
        </div>
      </div>
    </div>
  );
}

// ---- Framework — top level. Collapsed by default: name + critical roll-up +
// one-line read. Expands to its composite Signal(s). ----
function FrameworkRow({
  fw,
  signals,
  defaultOpen,
}: {
  fw: Framework;
  signals: Signal[];
  defaultOpen?: boolean;
}) {
  const [open, setOpen] = useState(!!defaultOpen);
  const panelId = useId();

  const composites = fw.composites
    .map((name) => signals.find((s) => s.name === name))
    .filter((s): s is Signal => !!s);

  // Honest roll-up: the framework is only as strong as its weakest composite
  // (its critical bottleneck), not a flattering average.
  const critical =
    composites.length > 0
      ? composites.reduce((min, s) => (s.score < min.score ? s : min), composites[0])
      : null;
  const score = critical ? critical.score : 0;
  const b = band(score);
  const hasData = composites.length > 0;

  return (
    <div className={`framework${open ? " open" : ""}`}>
      <button
        className="framework-head"
        onClick={() => hasData && setOpen((v) => !v)}
        aria-expanded={open}
        aria-controls={panelId}
        disabled={!hasData}
      >
        <span className="framework-chev" aria-hidden>
          <IconChevron />
        </span>
        <span className="framework-main">
          <span className="framework-titlerow">
            <span className="framework-name">{fw.name}</span>
            <span className={`framework-band ${b}`}>{hasData ? bandWord(b) : "Not scored"}</span>
          </span>
          <span className="framework-read">
            {hasData ? frameworkRead(b, critical) : fw.blurb}
          </span>
        </span>
        <span className="framework-readout">
          {hasData ? (
            <>
              <span className={`framework-score ${b}`}>{score.toFixed(1)}</span>
              <span className="framework-max">/{MAX}</span>
              <span className="framework-scorelabel">critical · lowest composite</span>
            </>
          ) : (
            <span className="framework-scorelabel">—</span>
          )}
        </span>
      </button>

      <div className="framework-panel" id={panelId} role="region">
        <div className="framework-panel-inner">
          <div className="framework-blurb">{fw.blurb}</div>
          <div className="composite-list">
            {composites.map((s) => (
              <Composite key={s.id ?? s.name} signal={s} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// The full instrument: three frameworks, collapsed by default.
export default function SignalsPanel({ signals }: { signals: Signal[] }) {
  return (
    <div className="frameworks">
      {FRAMEWORKS.map((fw) => (
        <FrameworkRow key={fw.key} fw={fw} signals={signals} />
      ))}
    </div>
  );
}
