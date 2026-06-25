import { Link } from "react-router-dom";
import { IconArrowRight } from "../components/Icons";

// A teaser of the signature instrument: an advisor's dossier with the 5 Signals.
const PREVIEW = [
  { name: "Market", val: 4.1, cls: "good", pct: 82 },
  { name: "Commercial", val: 3.2, cls: "mid", pct: 64 },
  { name: "Innovation", val: 4.4, cls: "good", pct: 88 },
  { name: "Scalability", val: 2.6, cls: "mid", pct: 52 },
  { name: "Green", val: 1.8, cls: "weak", pct: 36 },
];

const barColor: Record<string, string> = {
  good: "var(--band-good)",
  mid: "var(--band-mid)",
  weak: "var(--band-weak)",
};

export default function Landing() {
  return (
    <div className="landing">
      <nav className="landing-nav">
        <Link to="/" className="seal">
          <span className="seal-mark">OH</span>
          <span className="seal-word">
            OfficeHours<span className="tld">.ai</span>
          </span>
        </Link>
        <Link to="/login" className="btn secondary sm">Sign in</Link>
      </nav>

      <section className="landing-hero">
        <div>
          <span className="hero-kicker">Office hours for founders</span>
          <h1>
            A partner who has <em>read your file</em> before they speak.
          </h1>
          <p className="lead">
            Describe your company once. We build an evidence-based profile, then you hold office
            hours with specialist advisors who score your readiness, set your goals, and surface
            the programs you actually qualify for.
          </p>
          <div className="hero-cta">
            <Link to="/login" className="btn primary lg">
              Start your first meeting <IconArrowRight style={{ width: 16, height: 16 }} />
            </Link>
            <Link to="/register" className="btn secondary lg">How accounts work</Link>
          </div>
          <div className="hero-credentials">
            <div className="cred">
              <div className="cred-num">5</div>
              <div className="cred-label">Composite signals</div>
            </div>
            <div className="cred">
              <div className="cred-num">6</div>
              <div className="cred-label">Maturity stages</div>
            </div>
            <div className="cred">
              <div className="cred-num">1</div>
              <div className="cred-label">Onboarding, ever</div>
            </div>
          </div>
        </div>

        <div className="hero-artifact" aria-hidden>
          <div className="artifact-head">
            <div className="artifact-portrait">AM</div>
            <div>
              <div className="artifact-name">Amira, Market Advisor</div>
              <div className="artifact-role">Reading · agritech, pre-seed</div>
            </div>
          </div>
          {PREVIEW.map((s) => (
            <div className="artifact-row" key={s.name}>
              <span className="artifact-label">{s.name}</span>
              <span className="artifact-gauge">
                <span style={{ width: `${s.pct}%`, background: barColor[s.cls] }} />
              </span>
              <span className="artifact-val">{s.val.toFixed(1)}</span>
            </div>
          ))}
        </div>
      </section>

      <section className="landing-features">
        <div className="lf">
          <div className="lf-step">First · Profile</div>
          <h3>An evidence-based profile</h3>
          <p>
            One onboarding. The diagnoser reads your description and infers your maturity stage —
            with the evidence behind it, not a guess.
          </p>
        </div>
        <div className="lf">
          <div className="lf-step">Then · Signals</div>
          <h3>Five composite readings</h3>
          <p>
            Market, Commercial Offer, Innovation, Scalability, and Green — each scored with
            sub-criterion breakdowns and a plain-language rationale.
          </p>
        </div>
        <div className="lf">
          <div className="lf-step">Next · Action</div>
          <h3>Concrete next steps</h3>
          <p>
            Conclude a session and the scorer sets goals and action items grounded in real
            programs, then tracks your journey in the logbook.
          </p>
        </div>
      </section>
    </div>
  );
}
