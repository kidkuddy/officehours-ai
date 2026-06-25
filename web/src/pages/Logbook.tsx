import { useEffect, useState } from "react";
import { api, type EventItem } from "../api/client";
import { ErrorAlert, PageLoading, EmptyState } from "../components/ui";
import { IconLogbook } from "../components/Icons";

const EVENT_LABEL: Record<string, string> = {
  stage_change: "Stage change",
  signal_update: "Signals updated",
  session: "Session",
  goal: "Goal",
  onboarding: "Onboarding",
  document: "Document",
};

function titleCase(s: string): string {
  return s.replace(/[_-]/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function eventTitle(ev: EventItem): string {
  const p = ev.payload || {};
  const pick = (...keys: string[]) => {
    for (const k of keys) {
      const v = p[k];
      if (typeof v === "string" && v) return v;
    }
    return null;
  };
  return pick("title", "summary", "message", "name", "stage", "outcome") ?? EVENT_LABEL[ev.kind] ?? titleCase(ev.kind);
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function Logbook() {
  const [events, setEvents] = useState<EventItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    api
      .events()
      .then((e) => active && setEvents(e))
      .catch((e) => active && setError(e instanceof Error ? e.message : "We couldn't load your logbook."))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, []);

  if (loading) return <PageLoading label="Loading your logbook" />;

  return (
    <div className="content page-fade-in">
      <div className="page-header">
        <div className="page-eyebrow">Your journey</div>
        <h1 className="page-title">Logbook</h1>
        <p className="page-subtitle">Every significant moment in your founder journey, newest first.</p>
      </div>

      {error && <ErrorAlert message={error} />}

      {events.length === 0 ? (
        <EmptyState
          icon={<IconLogbook />}
          title="Your logbook is empty"
          hint="Milestones appear here as you onboard, hold sessions, and hit your goals."
        />
      ) : (
        <div className="card pad-lg">
          <div className="timeline">
            {events.map((ev) => (
              <div className="tl-item" key={ev.id}>
                <div className="tl-kind">{EVENT_LABEL[ev.kind] ?? titleCase(ev.kind)}</div>
                <div className="tl-title">{eventTitle(ev)}</div>
                <div className="tl-time">{formatTimestamp(ev.created_at)}</div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
