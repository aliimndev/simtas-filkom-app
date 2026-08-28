// ── types ──
export interface ThesisDetail {
  id: string;
  title: string;
  abstract: string | null;
  fieldOfStudy: string | null;
  thesisType: string;
  status: string;
  kaprodiNotes: string | null;
  student: { id: string; fullName: string; nim: string | null };
  supervisors: { id: string; fullName: string }[];
  academicYear: { name: string; semester: string };
  submittedAt: string;
  approvedAt: string | null;
}

export interface Actor {
  userId: string;
  role: string;
  ipAddress: string;
  userAgent: string;
}
