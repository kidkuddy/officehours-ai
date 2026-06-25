// Theme state for "The Practice": light "paper" vs dark "ink".
// Persisted in localStorage; defaults to the OS preference. The actual
// data-theme attribute is applied to <html> (also pre-paint in main.tsx
// to avoid a flash of the wrong theme).
import { useCallback, useEffect, useState } from "react";

export type Theme = "light" | "dark";

const THEME_KEY = "oh_theme";

export function resolveInitialTheme(): Theme {
  const stored = localStorage.getItem(THEME_KEY);
  if (stored === "light" || stored === "dark") return stored;
  const prefersDark =
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-color-scheme: dark)").matches;
  return prefersDark ? "dark" : "light";
}

export function applyTheme(theme: Theme): void {
  document.documentElement.setAttribute("data-theme", theme);
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(() => {
    const current = document.documentElement.getAttribute("data-theme");
    if (current === "light" || current === "dark") return current;
    return resolveInitialTheme();
  });

  // Keep <html> + storage in sync whenever theme changes.
  useEffect(() => {
    applyTheme(theme);
    localStorage.setItem(THEME_KEY, theme);
  }, [theme]);

  // Follow OS changes only while the user hasn't made an explicit choice.
  useEffect(() => {
    if (typeof window.matchMedia !== "function") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = (e: MediaQueryListEvent) => {
      if (!localStorage.getItem(THEME_KEY)) setThemeState(e.matches ? "dark" : "light");
    };
    mq.addEventListener?.("change", onChange);
    return () => mq.removeEventListener?.("change", onChange);
  }, []);

  const toggle = useCallback(() => {
    setThemeState((t) => (t === "dark" ? "light" : "dark"));
  }, []);

  const setTheme = useCallback((t: Theme) => setThemeState(t), []);

  return { theme, toggle, setTheme };
}
