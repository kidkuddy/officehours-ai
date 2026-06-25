// Typed fetch wrapper hitting /api with a bearer token from localStorage.
// Routes mirror BUILD_SPEC §4.

const TOKEN_KEY = "oh_token";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
  status: number;
  body: unknown;
  constructor(status: number, message: string, body: unknown) {
    super(message);
    this.status = status;
    this.body = body;
  }
}

type RequestOptions = {
  method?: string;
  body?: unknown;
  // When set, body is sent as multipart/form-data (no JSON Content-Type).
  formData?: FormData;
};

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = {};
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;

  let body: BodyInit | undefined;
  if (opts.formData) {
    body = opts.formData;
  } else if (opts.body !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(opts.body);
  }

  const res = await fetch(`/api${path}`, {
    method: opts.method ?? (body ? "POST" : "GET"),
    headers,
    body,
  });

  const text = await res.text();
  let parsed: unknown = undefined;
  if (text) {
    try {
      parsed = JSON.parse(text);
    } catch {
      parsed = text;
    }
  }

  if (!res.ok) {
    const msg =
      parsed && typeof parsed === "object" && "error" in (parsed as object)
        ? String((parsed as { error: unknown }).error)
        : res.statusText;
    throw new ApiError(res.status, msg, parsed);
  }

  return parsed as T;
}

// --- Types (mirror schema in BUILD_SPEC §2) ---

export type User = {
  id: string;
  email: string;
  name: string;
  created_at: string;
};

export type Profile = {
  user_id: string;
  company_text: string;
  stage: string;
  stage_evidence: unknown[];
  created_at: string;
  updated_at: string;
  // Authoritative onboarding flag from the backend (onboarded_at != null).
  onboarded: boolean;
};

export type Subscore = {
  criterion: string;
  weight: number;
  score: number;
  contribution: number;
};

export type Signal = {
  id: string;
  user_id: string;
  name: string;
  score: number;
  subscores: Subscore[];
  rationale: string;
  floor_triggered: boolean;
  updated_at: string;
};

export type Session = {
  id: string;
  user_id: string;
  kind: string;
  advisor_key: string;
  title: string;
  status: string;
  outcomes: string;
  created_at: string;
  concluded_at?: string | null;
};

export type Message = {
  id: string;
  session_id: string;
  role: string;
  content: string;
  created_at: string;
};

export type Goal = {
  id: string;
  user_id: string;
  title: string;
  description: string;
  status: string;
  source_session_id?: string | null;
  created_at: string;
  done_at?: string | null;
};

export type ActionItem = {
  id: string;
  user_id: string;
  session_id?: string | null;
  title: string;
  horizon: string;
  rationale: string;
  program_ref: string;
  status: string;
  created_at: string;
};

export type Document = {
  id: string;
  user_id?: string | null;
  collection: string;
  filename: string;
  mime: string;
  path: string;
  created_at: string;
};

export type EventItem = {
  id: string;
  user_id: string;
  kind: string;
  payload: Record<string, unknown>;
  created_at: string;
};

export type PendingJob = {
  id: string;
  status: string;
} | null;

export type AdvisorRef = { key: string; name: string; description: string };

export type JobResult = {
  status: string;
  output: Record<string, unknown>;
  error: string;
};

export type Dashboard = {
  stage: string;
  signals: Signal[];
  stats: Record<string, unknown>;
  parcours: EventItem[];
};

// --- API surface ---

export const api = {
  login: (email: string, password: string) =>
    request<{ token: string; user: User }>("/auth/login", {
      body: { email, password },
    }),

  me: () => request<{ user: User; profile: Profile | null }>("/me"),

  onboarding: (text: string) =>
    request<{ job_id: string }>("/onboarding", { body: { text } }),

  profile: () => request<Profile & { signals?: Signal[] }>("/profile"),

  signals: () => request<Signal[]>("/signals"),

  dashboard: () => request<Dashboard>("/dashboard"),

  goals: () => request<Goal[]>("/goals"),

  advisors: () => request<AdvisorRef[]>("/advisors"),

  learn: () => request<AdvisorRef[]>("/learn"),

  // Returns Session + optional job_id for the advisor's opening turn.
  createSession: (advisor_key: string, kind: string) =>
    request<Session & { job_id?: string }>("/sessions", { body: { advisor_key, kind } }),

  listSessions: (kind?: string) =>
    request<Session[]>(`/sessions${kind ? `?kind=${encodeURIComponent(kind)}` : ""}`),

  // pending_job is any queued/running job for this session (advisor's turn).
  getSession: (id: string) =>
    request<{ session: Session; messages: Message[]; action_items: ActionItem[]; pending_job?: PendingJob }>(
      `/sessions/${id}`,
    ),

  sendMessage: (id: string, content: string) =>
    request<{ job_id: string }>(`/sessions/${id}/messages`, { body: { content } }),

  concludeSession: (id: string) =>
    request<{ job_id: string }>(`/sessions/${id}/conclude`, { method: "POST" }),

  uploadDocument: (file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    return request<{ document: Document }>("/documents", { formData: fd });
  },

  listDocuments: () => request<Document[]>("/documents"),

  getJob: (id: string) => request<JobResult>(`/jobs/${id}`),

  // Logbook: reverse-chronological event feed.
  events: () => request<EventItem[]>("/events"),
};
