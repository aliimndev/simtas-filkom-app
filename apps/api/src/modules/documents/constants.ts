export const DOC_TYPES = [
  "proposal",
  "draft_chapter",
  "seminar_doc",
  "defense_doc",
  "final_thesis",
  "revision_sheet",
  "endorsement_letter",
] as const;

export const PENDING_REVIEW = "pending_review";

// Thesis statuses that allow document uploads (Go isDocumentEligible).
export const ELIGIBLE_STATUSES = new Set([
  "in_progress",
  "seminar_ready",
  "seminar_done",
  "defense_ready",
  "defense_done",
  "graduated",
]);
