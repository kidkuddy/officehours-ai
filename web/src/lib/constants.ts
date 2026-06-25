// Frozen domain constants from BUILD_SPEC §2. Order matters for the stepper.
export const STAGES = [
  "Ideation",
  "Market Validation",
  "Structuration",
  "Fundraising",
  "Launch Planning",
  "Growth",
] as const;

export const SIGNAL_NAMES = [
  "Market",
  "Commercial Offer",
  "Innovation",
  "Scalability",
  "Green",
] as const;

export const HORIZON_LABEL: Record<string, string> = {
  immediate: "Immediate",
  short: "Short term",
  medium: "Medium term",
};
