import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Profile, type Signal } from "../api/client";
import { ErrorAlert, PageLoading, EmptyState } from "../components/ui";
import SignalsPanel from "../components/SignalsPanel";
import StageStepper from "../components/StageStepper";
import { IconProfile, IconSpark } from "../components/Icons";

type Loaded = Profile & { signals?: Signal[] };

export default function ProfilePage() {
  const [data, setData] = useState<Loaded | null>(null);
  const [signals, setSignals] = useState<Signal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    Promise.all([api.profile(), api.signals().catch(() => [])])
      .then(([p, s]) => {
        if (!active) return;
        setData(p);
        setSignals(p.signals && p.signals.length ? p.signals : s);
      })
      .catch((e) => active && setError(e instanceof Error ? e.message : "We couldn't load your profile."))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, []);

  if (loading) return <PageLoading label="Loading your profile" />;

  return (
    <div className="content page-fade-in">
      <div className="page-head">
        <div className="page-header">
          <div className="page-eyebrow">Founder profile</div>
          <h1 className="page-title">Profile</h1>
          <p className="page-subtitle">Your evidence-based profile, stage, and Signals.</p>
        </div>
        <Link to="/onboarding" className="btn secondary sm">Re-run diagnosis</Link>
      </div>

      {error && <ErrorAlert message={error} />}

      {!data || !data.company_text ? (
        <EmptyState
          icon={<IconProfile />}
          title="No profile yet"
          hint="Describe your company and we'll build your evidence-based profile."
          action={<Link to="/onboarding" className="btn primary">Start onboarding</Link>}
        />
      ) : (
        <>
          <div className="section-label">Maturity stage</div>
          <div className="card pad-lg" style={{ marginBottom: 8 }}>
            <div style={{ marginBottom: 20 }}>
              <span className="badge brass"><span className="dot-mark" /> {data.stage}</span>
            </div>
            <StageStepper current={data.stage} />

            {Array.isArray(data.stage_evidence) && data.stage_evidence.length > 0 && (
              <div style={{ marginTop: 26, paddingTop: 22, borderTop: "1px solid var(--line)" }}>
                <div className="section-label" style={{ margin: "0 0 14px" }}>Evidence</div>
                <ul className="evidence">
                  {data.stage_evidence.map((ev, i) => (
                    <li key={i}>{typeof ev === "string" ? ev : JSON.stringify(ev)}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          <div className="section-label">In your words</div>
          <div className="card pad-lg" style={{ marginBottom: 8 }}>
            <p style={{ whiteSpace: "pre-wrap", margin: 0, fontFamily: "var(--font-display)", fontSize: 16, lineHeight: 1.7, color: "var(--slate)" }}>
              {data.company_text}
            </p>
          </div>

          <div className="section-label">Signals</div>
          {signals.length === 0 ? (
            <EmptyState
              icon={<IconSpark />}
              title="Signals aren't computed yet"
              hint="Hold and conclude an Office Hours session to generate scored Signals."
              action={<Link to="/office-hours" className="btn primary">Go to Office Hours</Link>}
            />
          ) : (
            <SignalsPanel signals={signals} />
          )}
        </>
      )}
    </div>
  );
}
