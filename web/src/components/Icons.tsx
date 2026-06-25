// Line-icon set — a single consistent stroke weight, drawn for "The Practice".
// Used in the rail, buttons, and empty states.
import type { SVGProps } from "react";

type P = SVGProps<SVGSVGElement>;
const base = {
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.6,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
};

export const IconDashboard = (p: P) => (
  <svg {...base} {...p}><path d="M4 13h6V4H4zM14 20h6v-9h-6zM14 7h6V4h-6zM4 20h6v-3H4z" /></svg>
);
export const IconProfile = (p: P) => (
  <svg {...base} {...p}><circle cx="12" cy="8" r="3.4" /><path d="M5 20a7 7 0 0 1 14 0" /></svg>
);
export const IconOfficeHours = (p: P) => (
  <svg {...base} {...p}><path d="M4 5h16v11H8l-4 3z" /><path d="M8 9h8M8 12.5h5" /></svg>
);
export const IconLearn = (p: P) => (
  <svg {...base} {...p}><path d="M4 6.5 12 3l8 3.5L12 10z" /><path d="M7 8.5V14c0 1.1 2.2 2.5 5 2.5s5-1.4 5-2.5V8.5" /><path d="M20 7v5" /></svg>
);
export const IconDataRoom = (p: P) => (
  <svg {...base} {...p}><path d="M5 4h9l5 5v11H5z" /><path d="M14 4v5h5" /><path d="M8 13h8M8 16.5h6" /></svg>
);
export const IconGoals = (p: P) => (
  <svg {...base} {...p}><circle cx="12" cy="12" r="8" /><circle cx="12" cy="12" r="3.2" /><path d="M12 4v3M12 17v3M4 12h3M17 12h3" /></svg>
);
export const IconLogbook = (p: P) => (
  <svg {...base} {...p}><path d="M6 3h11a2 2 0 0 1 2 2v16l-3-2-3 2-3-2-3 2V5a2 2 0 0 1 2-2z" /><path d="M9 8h7M9 11.5h5" /></svg>
);
export const IconLogout = (p: P) => (
  <svg {...base} {...p}><path d="M14 4h4a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1h-4" /><path d="M10 8l-4 4 4 4M6 12h11" /></svg>
);
export const IconSend = (p: P) => (
  <svg {...base} {...p}><path d="M5 12h13M12 6l6 6-6 6" /></svg>
);
export const IconAttach = (p: P) => (
  <svg {...base} {...p}><path d="M18 8.5 9.7 16.8a3 3 0 0 1-4.2-4.2l8.4-8.4a4.5 4.5 0 0 1 6.4 6.4l-8.3 8.3" /></svg>
);
export const IconArrowRight = (p: P) => (
  <svg {...base} {...p}><path d="M5 12h14M13 6l6 6-6 6" /></svg>
);
export const IconArrowLeft = (p: P) => (
  <svg {...base} {...p}><path d="M19 12H5M11 6l-6 6 6 6" /></svg>
);
export const IconUpload = (p: P) => (
  <svg {...base} {...p}><path d="M12 16V5M7 10l5-5 5 5" /><path d="M5 19h14" /></svg>
);
export const IconCheck = (p: P) => (
  <svg {...base} {...p}><path d="M5 12.5 10 17l9-10" /></svg>
);
export const IconSpark = (p: P) => (
  <svg {...base} {...p}><path d="M12 3v4M12 17v4M3 12h4M17 12h4M6 6l2.5 2.5M15.5 15.5 18 18M18 6l-2.5 2.5M8.5 15.5 6 18" /></svg>
);
export const IconDoc = (p: P) => (
  <svg {...base} {...p}><path d="M6 3h8l4 4v14H6z" /><path d="M14 3v4h4" /></svg>
);
export const IconSun = (p: P) => (
  <svg {...base} {...p}><circle cx="12" cy="12" r="4" /><path d="M12 2v2.5M12 19.5V22M2 12h2.5M19.5 12H22M4.9 4.9l1.8 1.8M17.3 17.3l1.8 1.8M19.1 4.9l-1.8 1.8M6.7 17.3l-1.8 1.8" /></svg>
);
export const IconMoon = (p: P) => (
  <svg {...base} {...p}><path d="M20 14.5A8 8 0 0 1 9.5 4a7 7 0 1 0 10.5 10.5z" /></svg>
);
export const IconChevron = (p: P) => (
  <svg {...base} {...p}><path d="M9 6l6 6-6 6" /></svg>
);
