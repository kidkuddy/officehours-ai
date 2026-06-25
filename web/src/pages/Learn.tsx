import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type AdvisorRef } from "../api/client";
import { ErrorAlert, PageLoading, EmptyState, initials } from "../components/ui";
import { IconArrowRight, IconLearn } from "../components/Icons";

export default function Learn() {
  const [concepts, setConcepts] = useState<AdvisorRef[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    api
      .learn()
      .then((c) => active && setConcepts(c))
      .catch((e) => active && setError(e instanceof Error ? e.message : "We couldn't load the topics."))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, []);

  if (loading) return <PageLoading label="Loading Learn" />;

  return (
    <div className="content page-fade-in">
      <div className="page-header">
        <div className="page-eyebrow">Concept tutors</div>
        <h1 className="page-title">Learn</h1>
        <p className="page-subtitle">
          Grounded conversations on the topics that matter at your stage. Pick one to start.
        </p>
      </div>

      {error && <ErrorAlert message={error} />}

      <div className="section-label">Available topics</div>
      {concepts.length === 0 ? (
        <EmptyState icon={<IconLearn />} title="No topics available yet" />
      ) : (
        <div className="pick-grid">
          {concepts.map((c) => (
            <Link to={`/learn/${c.key}`} className="pick-card" key={c.key} style={{ color: "inherit" }}>
              <div className="pc-portrait">{initials(c.name)}</div>
              <h3>{c.name}</h3>
              <p className="pc-desc">{c.description}</p>
              <div className="pc-foot">
                <span className="linklike">Start learning</span>
                <IconArrowRight style={{ width: 15, height: 15, color: "var(--brass)" }} />
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
