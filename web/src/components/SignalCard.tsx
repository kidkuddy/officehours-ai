import { useEffect, useState } from "react";
import type { Signal } from "../api/client";

const MAX = 5;

function band(score: number): "good" | "mid" | "weak" {
  if (score >= 3.5) return "good";
  if (score >= 2) return "mid";
  return "weak";
}

// A single Signal rendered as an instrument readout: a gauge with ticks,
// weighted sub-criteria bars, and the advisor's rationale set in serif.
export default function SignalCard({ signal, delay = 0 }: { signal: Signal; delay?: number }) {
  const [mounted, setMounted] = useState(false);
  const pct = Math.max(0, Math.min(100, (signal.score / MAX) * 100));
  const b = band(signal.score);
  const subs = Array.isArray(signal.subscores) ? signal.subscores : [];

  useEffect(() => {
    // Delay the fill so the gauges sweep in after the card reveals.
    const id = window.setTimeout(() => setMounted(true), 120 + delay);
    return () => window.clearTimeout(id);
  }, [delay]);

  return (
    <div className="signal reveal" style={{ animationDelay: `${delay}ms` }}>
      <div className="signal-head">
        <div className="signal-name">
          {signal.name}
          {signal.floor_triggered && <span className="badge brass">Capped</span>}
        </div>
        <div className="signal-readout">
          <span className="signal-score">{signal.score.toFixed(1)}</span>
          <span className="signal-max">/{MAX}</span>
        </div>
      </div>

      <div className="gauge">
        <div className={`gauge-fill ${b}`} style={{ width: mounted ? `${pct}%` : "0%" }} />
      </div>
      <div className="gauge-ticks" aria-hidden>
        <span>0</span><span>1</span><span>2</span><span>3</span><span>4</span><span>5</span>
      </div>

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
                  <span style={{ width: mounted ? `${sp}%` : "0%", transitionDelay: `${i * 60}ms` }} />
                </div>
              </div>
            );
          })}
        </div>
      )}

      {signal.rationale && <div className="signal-rationale">{signal.rationale}</div>}
    </div>
  );
}
