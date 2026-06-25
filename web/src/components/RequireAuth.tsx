import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import type { User } from "../api/client";

// Redirects to /login when there is no authenticated user, preserving the
// attempted location so we can bounce back after sign-in.
export default function RequireAuth({
  user,
  children,
}: {
  user: User | null;
  children: ReactNode;
}) {
  const location = useLocation();
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  return <>{children}</>;
}
