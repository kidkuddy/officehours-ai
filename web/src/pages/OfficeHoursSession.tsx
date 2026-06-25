import { useEffect, useState } from "react";
import { useParams, useLocation } from "react-router-dom";
import { api } from "../api/client";
import Chat from "../components/Chat";
import { PageLoading } from "../components/ui";

export default function OfficeHoursSession() {
  const { id } = useParams();
  const location = useLocation();
  // job_id for the advisor's opening turn, passed via navigate state from OfficeHours.
  const openingJobId =
    (location.state as { openingJobId?: string | null } | null)?.openingJobId ??
    null;

  const [title, setTitle] = useState<string>("Advisor");
  const [ready, setReady] = useState(false);

  // Resolve the advisor's display name for the header (best-effort).
  useEffect(() => {
    let active = true;
    Promise.all([api.getSession(id!), api.advisors().catch(() => [])])
      .then(([data, advisors]) => {
        if (!active) return;
        const a = advisors.find((x) => x.key === data.session.advisor_key);
        setTitle(a?.name ?? data.session.advisor_key ?? "Advisor");
      })
      .catch(() => active && setTitle("Advisor"))
      .finally(() => active && setReady(true));
    return () => {
      active = false;
    };
  }, [id]);

  if (!id) return <PageLoading label="Loading" />;
  if (!ready) return <PageLoading label="Opening session" />;

  return (
    <Chat
      sessionId={id}
      title={title}
      subtitle="Office Hours · grounded in real programs"
      kind="office_hours"
      allowUpload
      backTo="/office-hours"
      openingJobId={openingJobId}
    />
  );
}
