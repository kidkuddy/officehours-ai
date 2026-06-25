import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Goal } from "../api/client";
import { ErrorAlert, PageLoading, EmptyState, relativeTime } from "../components/ui";
import { IconGoals, IconCheck } from "../components/Icons";

export default function Goals() {
  const [goals, setGoals] = useState<Goal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    api
      .goals()
      .then((g) => active && setGoals(g))
      .catch((e) => active && setError(e instanceof Error ? e.message : "We couldn't load your goals."))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, []);

  const { open, done } = useMemo(
    () => ({
      open: goals.filter((g) => g.status !== "done"),
      done: goals.filter((g) => g.status === "done"),
    }),
    [goals],
  );

  if (loading) return <PageLoading label="Loading your goals" />;

  const Row = ({ g }: { g: Goal }) => {
    const isDone = g.status === "done";
    return (
      <div className="list-row">
        <div
          className="lr-icon"
          style={isDone ? { background: "var(--band-good-soft)", color: "var(--green-text)" } : undefined}
        >
          {isDone ? <IconCheck style={{ width: 16, height: 16 }} /> : <IconGoals style={{ width: 17, height: 17 }} />}
        </div>
        <div className="lr-main">
          <div className="lr-title" style={isDone ? { color: "var(--slate-2)", textDecoration: "line-through" } : undefined}>
            {g.title}
          </div>
          {g.description && <div className="lr-sub">{g.description}</div>}
          <div className="lr-sub">
            {isDone && g.done_at ? `Done ${relativeTime(g.done_at)}` : `Set ${relativeTime(g.created_at)}`}
          </div>
        </div>
        {isDone ? (
          <span className="badge green"><span className="dot-mark" /> Done</span>
        ) : (
          <span className="badge"><span className="dot-mark" /> Open</span>
        )}
      </div>
    );
  };

  return (
    <div className="content page-fade-in">
      <div className="page-header">
        <div className="page-eyebrow">Outcomes</div>
        <h1 className="page-title">Goals</h1>
        <p className="page-subtitle">Goals are set when you conclude an Office Hours session.</p>
      </div>

      {error && <ErrorAlert message={error} />}

      {goals.length === 0 ? (
        <EmptyState
          icon={<IconGoals />}
          title="No goals yet"
          hint="Conclude an Office Hours session and the scorer will set goals for you."
          action={<Link to="/office-hours" className="btn primary">Go to Office Hours</Link>}
        />
      ) : (
        <>
          <div className="section-label">Open ({open.length})</div>
          {open.length === 0 ? (
            <div className="alert info">No open goals — strong work.</div>
          ) : (
            <div className="list">
              {open.map((g) => (
                <Row key={g.id} g={g} />
              ))}
            </div>
          )}

          {done.length > 0 && (
            <>
              <div className="section-label">Completed ({done.length})</div>
              <div className="list">
                {done.map((g) => (
                  <Row key={g.id} g={g} />
                ))}
              </div>
            </>
          )}
        </>
      )}
    </div>
  );
}
