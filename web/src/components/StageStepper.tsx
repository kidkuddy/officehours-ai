import { useEffect, useState } from "react";
import { STAGES } from "../lib/constants";
import { IconCheck } from "./Icons";

// Two-letter dossier codes per stage, in journey order.
const STAGE_CODE: Record<string, string> = {
  Ideation: "ID",
  "Market Validation": "MV",
  Structuration: "ST",
  Fundraising: "FR",
  "Launch Planning": "LP",
  Growth: "GR",
};

// The maturity stage rendered as a vertical brass-tracked ladder.
// The track fills up to (and including) the current rung.
export default function StageStepper({ current }: { current: string }) {
  const currentIdx = STAGES.indexOf(current as (typeof STAGES)[number]);
  const [filled, setFilled] = useState(false);

  useEffect(() => {
    const id = window.setTimeout(() => setFilled(true), 150);
    return () => window.clearTimeout(id);
  }, []);

  // Track fill height: from the first node to the current node.
  const total = STAGES.length;
  const fillPct =
    currentIdx >= 0 ? Math.min(100, ((currentIdx + 0.0) / (total - 1)) * 100) : 0;

  return (
    <div className="ladder">
      <div className="ladder-track">
        <div className="ladder-track-fill" style={{ height: filled ? `${fillPct}%` : "0%" }} />
      </div>
      {STAGES.map((stage, i) => {
        const isPassed = currentIdx >= 0 && i < currentIdx;
        const isCurrent = stage === current;
        const cls = isCurrent ? "current" : isPassed ? "passed" : "";
        return (
          <div className={`rung ${cls}`} key={stage}>
            <span className="rung-node">
              {isPassed ? <IconCheck style={{ width: 12, height: 12 }} /> : i + 1}
            </span>
            <span className="rung-body">
              <span className="rung-code">{STAGE_CODE[stage] ?? ""}-{String(i + 1).padStart(2, "0")}</span>
              <span className="rung-name">{stage}</span>
            </span>
            {isCurrent && <span className="rung-now">You are here</span>}
          </div>
        );
      })}
    </div>
  );
}
