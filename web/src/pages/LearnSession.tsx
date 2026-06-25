// The Learn route is keyed by concept key. We resolve (or create) a learn
// session for that concept, then hand off to the shared Chat component.
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { api } from "../api/client";
import Chat from "../components/Chat";
import { ErrorAlert, PageLoading } from "../components/ui";

export default function LearnSession() {
  const { key } = useParams();
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [openingJobId, setOpeningJobId] = useState<string | null>(null);
  const [title, setTitle] = useState("Tutor");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!key) return;
    let active = true;

    (async () => {
      try {
        const [concepts, existing] = await Promise.all([
          api.learn().catch(() => []),
          api.listSessions("learn").catch(() => []),
        ]);
        if (!active) return;

        const concept = concepts.find((c) => c.key === key);
        setTitle(concept?.name ?? key);

        // Resume the most recent active learn session for this concept,
        // otherwise create a fresh one.
        const match = existing
          .filter((s) => s.advisor_key === key && s.status === "active")
          .sort(
            (a, b) =>
              new Date(b.created_at).getTime() -
              new Date(a.created_at).getTime(),
          )[0];

        if (match) {
          setSessionId(match.id);
        } else {
          const created = await api.createSession(key, "learn");
          if (active) {
            setSessionId(created.id);
            // Capture the tutor's opening turn job if provided.
            setOpeningJobId(created.job_id ?? null);
          }
        }
      } catch (e) {
        if (active)
          setError(
            e instanceof Error ? e.message : "Failed to open the tutor.",
          );
      }
    })();

    return () => {
      active = false;
    };
  }, [key]);

  if (error)
    return (
      <div className="content">
        <ErrorAlert message={error} />
      </div>
    );
  if (!sessionId) return <PageLoading label="Opening tutor" />;

  return (
    <Chat
      sessionId={sessionId}
      title={title}
      subtitle="Learn · grounded in topic knowledge base"
      kind="learn"
      allowUpload={false}
      backTo="/learn"
      openingJobId={openingJobId}
    />
  );
}
