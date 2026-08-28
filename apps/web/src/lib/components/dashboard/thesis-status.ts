export type StatusVariant = "pending" | "approved" | "rejected" | "draft" | "in_progress" | "completed";

export function thesisStatusProps(status: string): { variant: StatusVariant; label: string } {
  switch (status) {
    case "submitted": return { variant: "pending", label: "Menunggu Review" };
    case "approved": return { variant: "approved", label: "Disetujui" };
    case "rejected": return { variant: "rejected", label: "Ditolak" };
    case "cancelled": return { variant: "rejected", label: "Dibatalkan" };
    case "in_progress": return { variant: "in_progress", label: "Dalam Bimbingan" };
    case "seminar_ready": return { variant: "in_progress", label: "Siap Seminar" };
    case "seminar_done": return { variant: "in_progress", label: "Pasca Seminar" };
    case "defense_ready": return { variant: "in_progress", label: "Siap Sidang" };
    case "defense_done": return { variant: "in_progress", label: "Pasca Sidang" };
    case "graduated": return { variant: "completed", label: "Lulus" };
    default: return { variant: "draft", label: status.replace(/_/g, " ") };
  }
}
