// Shared chat surface for Office Hours (advisor) and Learn (concept) sessions.
// Flow per BUILD_SPEC §4/§7: POST a user message -> get {job_id} ->
// poll GET /jobs/:id -> refetch messages. Supports file upload (-> /documents),
// a "Conclude session" action, and a thinking state while the agent works.
// Also handles the advisor's opening turn: if the session was just created (openingJobId)
// or if getSession returns a pending_job, we poll that job before the user sends.
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { useNavigate } from "react-router-dom";
import { api, type ActionItem, type Message, type Session } from "../api/client";
import { useJobPoll } from "../hooks/useJobPoll";
import Markdown from "./Markdown";
import { ErrorAlert, PageLoading, Thinking, initials } from "./ui";
import { IconSend, IconAttach, IconArrowLeft, IconCheck } from "./Icons";
import { HORIZON_LABEL } from "../lib/constants";

type Props = {
  sessionId: string;
  title: string;
  subtitle?: string;
  kind: "office_hours" | "learn";
  allowUpload?: boolean;
  backTo: string;
  openingJobId?: string | null;
};

export default function Chat({
  sessionId,
  title,
  subtitle,
  kind,
  allowUpload = true,
  backTo,
  openingJobId,
}: Props) {
  const navigate = useNavigate();
  const job = useJobPoll();

  const [session, setSession] = useState<Session | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [actionItems, setActionItems] = useState<ActionItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const [draft, setDraft] = useState("");
  const [attachment, setAttachment] = useState<File | null>(null);
  const [sending, setSending] = useState(false);
  const [concluding, setConcluding] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const openingPolled = useRef(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const fileRef = useRef<HTMLInputElement>(null);
  const textRef = useRef<HTMLTextAreaElement>(null);

  const refetch = useCallback(async () => {
    const data = await api.getSession(sessionId);
    setSession(data.session);
    setMessages(data.messages);
    setActionItems(data.action_items ?? []);
  }, [sessionId]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setLoadError(null);

    api
      .getSession(sessionId)
      .then(async (data) => {
        if (!active) return;
        setSession(data.session);
        setMessages(data.messages);
        setActionItems(data.action_items ?? []);
        setLoading(false);

        const pendingId =
          (!openingPolled.current && openingJobId ? openingJobId : null) ??
          (!openingPolled.current && data.pending_job ? data.pending_job.id : null);

        if (pendingId && !openingPolled.current) {
          openingPolled.current = true;
          try {
            await job.poll(pendingId);
            if (active) await refetch();
          } catch {
            // Opening turn error is non-fatal; let the user type anyway.
          }
        }
      })
      .catch((e) => {
        if (active) {
          setLoadError(e instanceof Error ? e.message : "We couldn't open this session. Try again.");
          setLoading(false);
        }
      });

    return () => {
      active = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sessionId]);

  const thinking = job.phase === "running";

  useLayoutEffect(() => {
    const el = scrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages, thinking]);

  const autoGrow = () => {
    const el = textRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 160) + "px";
  };

  const concluded = session?.status === "concluded";

  const send = useCallback(async () => {
    const content = draft.trim();
    if ((!content && !attachment) || sending || thinking || concluded) return;
    setActionError(null);
    setSending(true);
    try {
      if (attachment) {
        await api.uploadDocument(attachment);
        setAttachment(null);
        if (fileRef.current) fileRef.current.value = "";
      }
      const text = content || (attachment ? `I've uploaded a document: ${attachment.name}` : "");
      setMessages((prev) => [
        ...prev,
        {
          id: `tmp-${Date.now()}`,
          session_id: sessionId,
          role: "user",
          content: text,
          created_at: new Date().toISOString(),
        },
      ]);
      setDraft("");
      if (textRef.current) textRef.current.style.height = "auto";

      const { job_id } = await api.sendMessage(sessionId, text);
      await job.poll(job_id);
      await refetch();
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Your message didn't send. Try again.");
      void refetch();
    } finally {
      setSending(false);
    }
  }, [draft, attachment, sending, thinking, concluded, sessionId, job, refetch]);

  const conclude = useCallback(async () => {
    if (concluding || concluded) return;
    setActionError(null);
    setConcluding(true);
    try {
      const { job_id } = await api.concludeSession(sessionId);
      await job.poll(job_id);
      await refetch();
      navigate("/dashboard");
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "We couldn't conclude the session. Try again.");
    } finally {
      setConcluding(false);
    }
  }, [concluding, concluded, sessionId, job, refetch, navigate]);

  const onKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      void send();
    }
  };

  if (loading) return <PageLoading label="Opening the session" />;
  if (loadError)
    return (
      <div className="content">
        <ErrorAlert message={loadError} />
      </div>
    );

  const portrait = initials(title);
  const visibleMessages = messages.filter((m) => m.role !== "system");
  const advisorOpening = thinking && visibleMessages.length === 0;

  return (
    <div className="session-shell">
      <header className="chat-header">
        <button className="rail-logout" onClick={() => navigate(backTo)} aria-label="Back" title="Back"
          style={{ background: "var(--card)", border: "1px solid var(--line-strong)", color: "var(--slate)" }}>
          <IconArrowLeft />
        </button>
        <div className="ch-portrait">{portrait}</div>
        <div className="ch-meta">
          <div className="ch-title">{title}</div>
          {subtitle && <div className="ch-sub">{subtitle}</div>}
        </div>
        {concluded ? (
          <span className="badge green"><span className="dot-mark" /> Concluded</span>
        ) : (
          <span className="badge live"><span className="dot-mark" /> In session</span>
        )}
      </header>

      <div className="chat-scroll" ref={scrollRef}>
        <div className="chat-inner">
          {advisorOpening && (
            <div className="msg assistant">
              <div className="avatar">{portrait}</div>
              <div className="bubble">
                <div className="role">{title}</div>
                <div style={{ paddingTop: 4 }}>
                  <Thinking />
                </div>
              </div>
            </div>
          )}

          {!thinking && visibleMessages.length === 0 && (
            <div className="alert info">
              {kind === "learn"
                ? "Ask anything about this topic. Answers are grounded in the knowledge base."
                : "Share your situation. The advisor responds with guidance grounded in real programs."}
            </div>
          )}

          {visibleMessages.map((m) => (
            <div className={`msg ${m.role}`} key={m.id}>
              <div className="avatar">{m.role === "user" ? "You" : portrait}</div>
              <div className="bubble">
                <div className="role">{m.role === "user" ? "You" : title}</div>
                {m.role === "user" ? (
                  <div className="body">{m.content}</div>
                ) : (
                  <Markdown>{m.content}</Markdown>
                )}
              </div>
            </div>
          ))}

          {thinking && visibleMessages.length > 0 && (
            <div className="msg assistant">
              <div className="avatar">{portrait}</div>
              <div className="bubble">
                <div className="role">{title}</div>
                <div style={{ paddingTop: 4 }}>
                  <Thinking />
                </div>
              </div>
            </div>
          )}

          {kind === "office_hours" && actionItems.length > 0 && (
            <div className="action-rail">
              <div className="action-rail-header">Action items</div>
              {actionItems.map((ai) => (
                <div className="action-item" key={ai.id}>
                  <span className="ai-h">{HORIZON_LABEL[ai.horizon] ?? ai.horizon}</span>
                  <div className="ai-body">
                    <div className="ai-title">{ai.title}</div>
                    {ai.rationale && <div className="ai-meta">{ai.rationale}</div>}
                    {ai.program_ref && <div className="ai-meta">Program · {ai.program_ref}</div>}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="chat-input-bar">
        <div className="chat-input-inner">
          {actionError && <ErrorAlert message={actionError} />}
          {job.error && job.phase === "error" && <ErrorAlert message={job.error} />}

          {concluded ? (
            <div className="alert info">
              Session concluded. Your Signals, goals, and logbook are updated — see the Dashboard.
            </div>
          ) : (
            <>
              {attachment && (
                <span className="attach-pill">
                  <IconAttach style={{ width: 13, height: 13 }} /> {attachment.name}
                  <button
                    onClick={() => {
                      setAttachment(null);
                      if (fileRef.current) fileRef.current.value = "";
                    }}
                    aria-label="Remove attachment"
                  >
                    ×
                  </button>
                </span>
              )}
              <div className="composer">
                <textarea
                  ref={textRef}
                  value={draft}
                  placeholder={
                    advisorOpening
                      ? `${title} is reading your file…`
                      : thinking
                      ? `${title} is thinking…`
                      : "Write to the advisor…"
                  }
                  onChange={(e) => {
                    setDraft(e.target.value);
                    autoGrow();
                  }}
                  onKeyDown={onKeyDown}
                  rows={1}
                  disabled={sending || thinking}
                  aria-label="Message"
                />
                {allowUpload && (
                  <>
                    <input
                      ref={fileRef}
                      type="file"
                      accept=".md,.txt,.pdf"
                      style={{ display: "none" }}
                      onChange={(e) => setAttachment(e.target.files?.[0] ?? null)}
                    />
                    <button
                      type="button"
                      className="icon-btn"
                      title="Attach a document"
                      aria-label="Attach a document"
                      onClick={() => fileRef.current?.click()}
                      disabled={sending || thinking}
                    >
                      <IconAttach />
                    </button>
                  </>
                )}
                <button
                  type="button"
                  className="icon-btn send"
                  onClick={() => void send()}
                  disabled={sending || thinking || (!draft.trim() && !attachment)}
                  title="Send"
                  aria-label="Send message"
                >
                  <IconSend />
                </button>
              </div>

              {kind === "office_hours" && (
                <div className="chat-actions">
                  <button
                    className="conclude-btn"
                    onClick={() => void conclude()}
                    disabled={concluding || thinking || messages.length === 0}
                  >
                    {concluding ? (
                      <>
                        <span className="spinner" style={{ width: 12, height: 12 }} /> Scoring the session…
                      </>
                    ) : (
                      <>
                        <IconCheck style={{ width: 13, height: 13 }} /> Conclude session
                      </>
                    )}
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
