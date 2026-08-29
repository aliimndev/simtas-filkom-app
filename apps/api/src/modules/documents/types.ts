export interface Actor {
  userId: string;
  role: string;
  ipAddress?: string | null;
  userAgent?: string | null;
}

export interface DocumentDetail {
  id: string;
  thesisId: string;
  documentType: string;
  chapterNumber: number | null;
  version: number;
  fileName: string;
  fileUrl: string;
  fileSize: number | null;
  status: string;
  uploadedBy: string;
  reviewerId: string | null;
  reviewerNotes: string | null;
  reviewedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface UploadInput {
  thesisId: string;
  documentType: string;
  chapterNumber?: number | null;
  fileName: string;
  fileUrl: string;
  fileSize?: number | null;
  actor: Actor;
}

export interface ListFilter {
  thesisId: string;
  documentType?: string;
  status?: string;
  userId: string;
  role: string;
}

export interface ReviewInput {
  id: string;
  decision: string;
  notes?: string | null;
  actor: Actor;
}
