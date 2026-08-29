export const ACTIVE_SEMINAR_STATUSES = ["pending", "scheduled", "passed"] as const;
export const STAFF_ROLES = new Set(["ADMIN_FAKULTAS", "KAPRODI"]);

export type SeminarActor = {
  userId: string;
  role: string;
  ipAddress?: string | null;
  userAgent?: string | null;
};

export type ScheduleSeminarInput = {
  scheduledAt: string;
  room: string;
  examinerIds: string[];
};

export type SeminarListFilter = {
  status?: string;
  page: number;
  perPage: number;
};

export class SeminarError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
  ) {
    super(message);
  }
}
