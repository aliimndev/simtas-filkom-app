export type ArchiveActor = {
  userId: string;
  role: string;
  ipAddress?: string | null;
  userAgent?: string | null;
};

export type ArchiveInput = {
  fileUrl: string;
  fileName: string;
  abstractId: string;
  abstractEn?: string | null;
  keywords?: string[] | null;
  graduationYear: number;
};
