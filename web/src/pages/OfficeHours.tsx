import { useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { api, type AdvisorRef, type Session } from "../api/client";
import { ErrorAlert, PageLoading, EmptyState, Spinner, initials, relativeTime } from "../components/ui";
import { IconArrowRight, IconOfficeHours } from "../components/Icons";

export default function OfficeHours() {
  const navigate = useNavigate();
  const [advisors, setAdvisors] = useState<AdvisorRef[]>([]);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [starting, setStarting] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    Promise.all([api.advisors(), api.listSessions("office_hours").catch(() => [])])
      .then(([a, s]) => {
        if (!active) return;
        setAdvisors(a);
        setSessions(s);
      })
      .catch((e) => setError(e instanceof Error ? e.message : "We couldn't load the advisors."))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, []);

  const start = async (key: string) => {
    if (starting) return;
    setError(null);
    setStarting(key);
    try {
      const result = await api.createSession(key, "office_hours");
      navigate(`/office-hours/${result.id}`, {
        state: { openingJobId: result.job_id ?? null },
      });
    } catch (e) {
      setError(e instanceof Error ? e.message : "We couldn't start the session.");
      setStarting(null);
    }
  };

  if (loading) return <PageLoading label="Loading Office Hours" />;

  const advisorName = (key: string) => advisors.find((a) => a.key === key)?.name ?? key;

  return (
    <div className="content page-fade-in">
      <div className="page-header">
        <div className="page-eyebrow">Advisory sessions</div>
        <h1 className="page-title">Office Hours</h1>
        <p className="page-subtitle">Choose a specialist advisor to open a session, or resume one below.</p>
      </div>

      {error && <ErrorAlert message={error} />}

      <div className="section-label">Your advisors</div>
      {advisors.length === 0 ? (
        <EmptyState icon={<IconOfficeHours />} title="No advisors available right now" />
      ) : (
        <div className="pick-grid">
          {advisors.map((a) => (
            <div className="pick-card" key={a.key}>
              <div className="pc-portrait">{initials(a.name)}</div>
              <h3>{a.name}</h3>
              <p className="pc-desc">{a.description}</p>
              <div className="pc-foot">
                <button className="btn primary sm" onClick={() => void start(a.key)} disabled={!!starting}>
                  {starting === a.key ? (
                    <><Spinner /> Opening…</>
                  ) : (
                    <>Start session <IconArrowRight style={{ width: 15, height: 15 }} /></>
                  )}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="section-label">Session history</div>
      {sessions.length === 0 ? (
        <EmptyState
          icon={<IconOfficeHours />}
          title="No sessions yet"
          hint="Pick an advisor above to open your first office hours."
        />
      ) : (
        <div className="list">
          {sessions.map((s) => (
            <Link to={`/office-hours/${s.id}`} className="list-row" key={s.id} style={{ color: "inherit" }}>
              <div className="lr-icon">{initials(advisorName(s.advisor_key))}</div>
              <div className="lr-main">
                <div className="lr-title">{s.title || advisorName(s.advisor_key)}</div>
                <div className="lr-sub">{advisorName(s.advisor_key)} · {relativeTime(s.created_at)}</div>
              </div>
              {s.status === "concluded" ? (
                <span className="badge green"><span className="dot-mark" /> Concluded</span>
              ) : (
                <span className="badge live"><span className="dot-mark" /> In session</span>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
