import { Link } from "react-router-dom";
import { IconArrowRight } from "../components/Icons";

export default function Register() {
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
          Every founder gets a <em>seat</em> at the table.
        </blockquote>
        <div className="auth-aside-foot">
          Accounts are provisioned by an operator. Once yours exists, sign in and your first
          meeting begins.
        </div>
      </aside>

      <div className="auth-panel">
        <div className="auth-card" style={{ maxWidth: 440 }}>
          <div className="auth-head">
            <h1>Accounts are invite-only</h1>
            <p>There's no public sign-up yet. An operator provisions your login with the
              {" "}<code style={{ fontFamily: "var(--font-mono)", fontSize: 12.5, background: "var(--paper-2)", padding: "1px 6px", borderRadius: 5, border: "1px solid var(--line)" }}>ohctl</code>{" "}
              tool:</p>
          </div>

          <pre className="codeblock">{`ohctl seed user \\
  --email you@company.com \\
  --password "your-password" \\
  --name "Your Name"`}</pre>

          <p style={{ color: "var(--slate)", fontSize: 13.5, margin: "16px 0 22px" }}>
            Once your account exists, sign in on the next page.
          </p>

          <Link to="/login" className="btn primary">
            Go to sign in <IconArrowRight style={{ width: 16, height: 16 }} />
          </Link>
        </div>
      </div>
    </div>
  );
}
