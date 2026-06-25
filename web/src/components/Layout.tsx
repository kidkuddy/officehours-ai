import { NavLink, Outlet, useNavigate, useLocation, Navigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { useTheme } from "../hooks/useTheme";
import { PageLoading, initials } from "./ui";
import RequireAuth from "./RequireAuth";
import {
  IconDashboard,
  IconProfile,
  IconOfficeHours,
  IconLearn,
  IconDataRoom,
  IconGoals,
  IconLogbook,
  IconLogout,
  IconSun,
  IconMoon,
} from "./Icons";

type Item = { to: string; label: string; Icon: (p: { className?: string }) => JSX.Element };

const NAV_GROUPS: { label: string; items: Item[] }[] = [
  {
    label: "Practice",
    items: [
      { to: "/dashboard", label: "Dashboard", Icon: IconDashboard },
      { to: "/profile", label: "Profile", Icon: IconProfile },
    ],
  },
  {
    label: "Sessions",
    items: [
      { to: "/office-hours", label: "Office Hours", Icon: IconOfficeHours },
      { to: "/learn", label: "Learn", Icon: IconLearn },
    ],
  },
  {
    label: "Records",
    items: [
      { to: "/data-room", label: "Data Room", Icon: IconDataRoom },
      { to: "/goals", label: "Goals", Icon: IconGoals },
      { to: "/logbook", label: "Logbook", Icon: IconLogbook },
    ],
  },
];

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { loading, user, profile, logout } = useAuth();
  const { theme, toggle } = useTheme();

  if (loading) return <PageLoading label="Opening the practice" />;

  // Authoritative onboarding flag from the backend profile.
  const isOnboarded = !!(profile && profile.onboarded);
  const isOnboardingRoute = location.pathname === "/onboarding";

  const doLogout = () => {
    logout();
    navigate("/login");
  };

  // Onboarding is a full-screen wizard with no rail. Let it render alone.
  if (isOnboardingRoute) {
    return (
      <RequireAuth user={user}>
        <Outlet />
      </RequireAuth>
    );
  }

  // Strict gate: an authenticated-but-not-onboarded user can reach ONLY the
  // wizard. We redirect here at the layout boundary, before any page route is
  // rendered or its data fetched — so no other screen ever flashes.
  if (user && !isOnboarded) {
    return (
      <RequireAuth user={user}>
        <Navigate to="/onboarding" replace />
      </RequireAuth>
    );
  }

  return (
    <RequireAuth user={user}>
      <div className="app-shell">
        <aside className="rail">
          <div className="rail-head">
            <NavLink to="/dashboard" className="seal" aria-label="OfficeHours.ai home">
              <span className="seal-mark">OH</span>
              <span>
                <span className="seal-word">
                  OfficeHours<span className="tld">.ai</span>
                </span>
                <span className="seal-kicker">Founder Advisory</span>
              </span>
            </NavLink>
          </div>

          <nav className="rail-nav" aria-label="Primary">
            {NAV_GROUPS.map((group) => (
              <div className="nav-group" key={group.label}>
                <div className="nav-group-label">{group.label}</div>
                {group.items.map(({ to, label, Icon }) => (
                  <NavLink
                    key={to}
                    to={to}
                    className={({ isActive }) => `nav-link${isActive ? " active" : ""}`}
                  >
                    <span className="nav-icon"><Icon /></span>
                    <span className="nav-label">{label}</span>
                  </NavLink>
                ))}
              </div>
            ))}
          </nav>

          {user && (
            <div className="rail-user">
              <div className="rail-avatar">{initials(user.name || user.email)}</div>
              <div className="rail-user-info">
                <div className="rail-user-name">{user.name || user.email}</div>
                <div className="rail-user-email">{user.email}</div>
              </div>
              <button
                className="rail-iconbtn"
                onClick={toggle}
                title={theme === "dark" ? "Switch to light" : "Switch to dark"}
                aria-label={theme === "dark" ? "Switch to light theme" : "Switch to dark theme"}
              >
                {theme === "dark" ? <IconSun /> : <IconMoon />}
              </button>
              <button className="rail-logout" onClick={doLogout} title="Sign out" aria-label="Sign out">
                <IconLogout />
              </button>
            </div>
          )}
        </aside>

        <main className="main">
          <Outlet />
        </main>
      </div>
    </RequireAuth>
  );
}
