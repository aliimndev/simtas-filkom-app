export const STAFF_ROLES = new Set(["ADMIN_FAKULTAS", "KAPRODI"]);

export type DefenseActor = {
  userId: string;
  role: string;
  ipAddress?: string | null;
  userAgent?: string | null;
};

export type DefenseListFilter = {
  thesisId?: string;
  page: number;
  perPage: number;
};

export type DefenseScoreInput = {
  componentName: string;
  componentWeight: number;
  score: number;
  notes?: string | null;
};

export type DefenseFinalizeInput = {
  revisionNotes?: string | null;
};

export type SuratTugasInput = {
  letterNumber: string;
  issueDate: string;
  fileName: string;
  fileUrl: string;
};

export type ScheduleDefenseInput = {
  scheduledAt: string;
  room: string;
  examinerIds: string[];
};

export class DefenseError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
  ) {
    super(message);
  }
}
