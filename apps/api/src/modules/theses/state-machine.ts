// ── thesis state machine ──
const VALID_TRANSITIONS: Record<string, string[]> = {
  submitted: ["approved", "rejected", "cancelled"],
  approved: ["in_progress", "cancelled"],
  in_progress: ["seminar_ready", "cancelled"],
  seminar_ready: ["defense_ready"],
  defense_ready: ["defense_done"],
  defense_done: ["graduated"],
  rejected: [],
  graduated: [],
  cancelled: [],
};

export function canTransition(from: string, to: string): boolean {
  if (to === "cancelled") return from !== "cancelled";
  return (VALID_TRANSITIONS[from] ?? []).includes(to);
}
