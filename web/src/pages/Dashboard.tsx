import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Dashboard as DashboardData, type Goal } from "../api/client";
import { ErrorAlert, PageLoading, EmptyState, relativeTime } from "../components/ui";
import SignalsPanel from "../components/SignalsPanel";
import StageStepper from "../components/StageStepper";
import { IconArrowRight, IconGoals, IconSpark } from "../components/Icons";

const STAT_LABELS: Record<string, string> = {
  sessions: "Sessions",
  sessions_total: "Sessions",
  concluded_sessions: "Concluded",
  goals: "Goals",
  open_goals: "Open goals",
  goals_open: "Open goals",
  goals_done: "Goals done",
  action_items: "Action items",
  documents: "Documents",
  signals: "Signals",
};

function titleCase(s: string): string {
  return s.replace(/[_-]/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function Dashboard() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [goals, setGoals] = useState<Goal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    Promise.all([api.dashboard(), api.goals().catch(() => [] as Goal[])])
      .then(([d, g]) => {
        if (!active) return;
        setData(d);
        setGoals(g);
      })
      .catch((e) => active && setError(e instanceof Error ? e.message : "We couldn't load your dashboard."))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, []);

  if (loading) return <PageLoading label="Loading your dashboard" />;
  if (error)
    return (
      <div className="content">
        <ErrorAlert message={error} />
      </div>
    );
  if (!data) return null;

  const stats = data.stats || {};
  const statEntries = Object.entries(stats).filter(
    ([, v]) => typeof v === "number" || typeof v === "string",
  );
  const hasProfile = !!data.stage;
  const openGoals = goals.filter((g) => g.status !== "done");

  return (
    <div className="content page-fade-in">
      <div className="page-head">
        <div className="page-header">
          <div className="page-eyebrow">Your practice</div>
          <h1 className="page-title">Dashboard</h1>
          <p className="page-subtitle">Your stage, your Signals across the three frameworks, and the goals in motion.</p>
        </div>
        <Link to="/office-hours" className="btn primary">
          Start office hours <IconArrowRight style={{ width: 16, height: 16 }} />
        </Link>
      </div>

      {!hasProfile ? (
        <EmptyState
          icon={<IconSpark />}
          title="Set up your profile first"
          hint="Describe your company and we'll diagnose your stage and Signals."
          action={<Link to="/onboarding" className="btn primary">Start onboarding</Link>}
        />
      ) : (
        <>
          {/* Maturity stage — the hero */}
          <div className="section-label">Maturity stage</div>
          <div className="card pad-lg" style={{ marginBottom: 8 }}>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 22, gap: 12 }}>
              <span style={{ fontFamily: "var(--font-display)", fontWeight: 600, fontSize: 17 }}>
                The founder journey
              </span>
              <span className="badge brass"><span className="dot-mark" /> {data.stage}</span>
            </div>
            <StageStepper current={data.stage} />
          </div>

          {/* Activity */}
          {statEntries.length > 0 && (
            <>
              <div className="section-label">Activity</div>
              <div className="stat-row" style={{ marginBottom: 8 }}>
                {statEntries.map(([k, v]) => (
                  <div className="stat" key={k}>
                    <div className="stat-val">{String(v)}</div>
                    <div className="stat-label">{STAT_LABELS[k] ?? titleCase(k)}</div>
                  </div>
                ))}
              </div>
            </>
          )}

          {/* Signals — grouped under the three stress-test frameworks */}
          <div className="section-label">Stress-test frameworks</div>
          {data.signals && data.signals.length > 0 ? (
            <SignalsPanel signals={data.signals} />
          ) : (
            <EmptyState
              icon={<IconSpark />}
              title="Signals aren't computed yet"
              hint="Conclude an Office Hours session to generate scored Signals with breakdowns."
              action={<Link to="/office-hours" className="btn primary">Go to Office Hours</Link>}
            />
          )}

          {/* Open goals */}
          <div className="section-label">
            <span>Open goals ({openGoals.length})</span>
            <Link to="/goals" className="linklike" style={{ marginLeft: "auto" }}>View all</Link>
          </div>
          {openGoals.length === 0 ? (
            <div className="alert info">
              No open goals yet — conclude an Office Hours session and the scorer will set them.
            </div>
          ) : (
            <div className="list">
              {openGoals.slice(0, 5).map((g) => (
                <div className="list-row" key={g.id}>
                  <div className="lr-icon"><IconGoals style={{ width: 17, height: 17 }} /></div>
                  <div className="lr-main">
                    <div className="lr-title">{g.title}</div>
                    {g.description && <div className="lr-sub">{g.description}</div>}
                    <div className="lr-sub">Set {relativeTime(g.created_at)}</div>
                  </div>
                  <span className="badge"><span className="dot-mark" /> Open</span>
                </div>
              ))}
              {openGoals.length > 5 && (
                <div style={{ textAlign: "center", padding: "10px 0" }}>
                  <Link to="/goals" className="linklike">
                    +{openGoals.length - 5} more
                  </Link>
                </div>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
}
