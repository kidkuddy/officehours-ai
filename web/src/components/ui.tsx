import type { ReactNode } from "react";

export function Spinner() {
  return <span className="spinner" aria-label="loading" />;
}

export function PageLoading({ label = "Loading" }: { label?: string }) {
  return (
    <div className="page-loading">
      <Spinner /> {label}…
    </div>
  );
}

export function Thinking() {
  return (
    <span className="thinking" aria-label="thinking" role="status">
      <span /> <span /> <span />
    </span>
  );
}

export function ErrorAlert({ message }: { message: string }) {
  return (
    <div className="alert error" role="alert">
      {message}
    </div>
  );
}

export function EmptyState({
  icon,
  title,
  hint,
  action,
}: {
  icon?: ReactNode;
  title: string;
  hint?: string;
  action?: ReactNode;
}) {
  return (
    <div className="empty">
      {icon && <div className="empty-icon">{icon}</div>}
      <h3>{title}</h3>
      {hint && <p>{hint}</p>}
      {action && <div>{action}</div>}
    </div>
  );
}

export function initials(s: string): string {
  const parts = s.replace(/[-_.@]/g, " ").trim().split(/\s+/);
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function relativeTime(iso: string): string {
  const d = new Date(iso);
  const diff = Date.now() - d.getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  const days = Math.floor(h / 24);
  if (days < 7) return `${days}d ago`;
  return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}
