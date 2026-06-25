import { useState } from "react";
import { Link, useNavigate, useLocation } from "react-router-dom";
import { api, setToken } from "../api/client";
import { ErrorAlert, Spinner } from "../components/ui";
import { IconArrowRight } from "../components/Icons";

export default function Login() {
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from;
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setBusy(true);
    try {
      const { token } = await api.login(email, password);
      setToken(token);
      navigate(from && from !== "/login" ? from : "/dashboard");
    } catch (err) {
      setError(
        err instanceof Error ? err.message || "That email and password didn't match." : "Sign-in failed.",
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="auth-wrap">
      <aside className="auth-aside">
        <Link to="/" className="auth-aside-seal">
          <span className="seal-mark">OH</span>
          <span className="seal-word" style={{ color: "#f4f2ec" }}>
            OfficeHours<span className="tld">.ai</span>
          </span>
        </Link>
        <blockquote className="auth-quote">
          A partner who has <em>read your file</em> before they speak.
        </blockquote>
        <div className="auth-aside-foot">
          Specialist advisors score your readiness, set your next goals, and surface the programs
          you actually qualify for.
        </div>
      </aside>

      <div className="auth-panel">
        <form className="auth-card" onSubmit={submit}>
          <div className="auth-head">
            <h1>Welcome back</h1>
            <p>Sign in to return to the practice.</p>
          </div>

          {error && <ErrorAlert message={error} />}

          <div className="field">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              className="input"
              type="email"
              autoComplete="username"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="founder@company.com"
              required
            />
          </div>
          <div className="field">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              className="input"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
            />
          </div>

          <button className="btn primary" type="submit" disabled={busy} style={{ width: "100%", marginTop: 6 }}>
            {busy ? (
              <><Spinner /> Signing in…</>
            ) : (
              <>Sign in <IconArrowRight style={{ width: 16, height: 16 }} /></>
            )}
          </button>

          <p style={{ marginTop: 20, textAlign: "center", fontSize: 13, color: "var(--slate-2)" }}>
            No account yet?{" "}
            <Link to="/register" style={{ color: "var(--brass-deep)", fontWeight: 600 }}>
              How accounts work
            </Link>
          </p>
        </form>
      </div>
    </div>
  );
}
