import { roleLabel } from "./navigation";

export type StatusVariant =
  | "pending"
  | "approved"
  | "rejected"
  | "draft"
  | "in_progress"
  | "completed";

export interface StatusProps {
  variant: StatusVariant;
  label: string;
}

export function thesisStatusProps(status: string): StatusProps {
  switch (status) {
    case "submitted":
      return { variant: "pending", label: "Menunggu Review" };
    case "approved":
      return { variant: "approved", label: "Disetujui" };
    case "rejected":
      return { variant: "rejected", label: "Ditolak" };
    case "cancelled":
      return { variant: "rejected", label: "Dibatalkan" };
    case "in_progress":
      return { variant: "in_progress", label: "Dalam Bimbingan" };
    case "seminar_ready":
      return { variant: "in_progress", label: "Siap Seminar" };
    case "seminar_done":
      return { variant: "in_progress", label: "Pasca Seminar" };
    case "defense_ready":
      return { variant: "in_progress", label: "Siap Sidang" };
    case "defense_done":
      return { variant: "in_progress", label: "Pasca Sidang" };
    case "graduated":
      return { variant: "completed", label: "Lulus" };
    default:
      return { variant: "draft", label: status.replace(/_/g, " ") };
  }
}

export function seminarStatusProps(status: string): StatusProps {
  switch (status) {
    case "passed":
      return { variant: "completed", label: "Lulus" };
    case "failed":
      return { variant: "rejected", label: "Tidak Lulus" };
    case "scheduled":
      return { variant: "in_progress", label: "Terjadwal" };
    case "pending":
      return { variant: "pending", label: "Diajukan" };
    default:
      return { variant: "draft", label: status ? status.replace(/_/g, " ") : "—" };
  }
}

export function titleChangeStatusProps(
  status: "PENDING" | "APPROVED" | "REJECTED" | "CANCELLED" | string,
): StatusProps {
  switch (status) {
    case "PENDING":
      return { variant: "pending", label: "Menunggu Persetujuan" };
    case "APPROVED":
      return { variant: "approved", label: "Disetujui" };
    case "REJECTED":
      return { variant: "rejected", label: "Ditolak" };
    default:
      return { variant: "draft", label: "Dibatalkan" };
  }
}

export function roleStatusProps(role?: string | null): StatusProps {
  switch ((role ?? "").toUpperCase()) {
    case "ADMIN_FAKULTAS":
      return { variant: "completed", label: roleLabel(role) };
    case "KAPRODI":
      return { variant: "in_progress", label: roleLabel(role) };
    case "DOSEN_PEMBIMBING":
      return { variant: "approved", label: roleLabel(role) };
    case "DOSEN_PENGUJI":
      return { variant: "pending", label: roleLabel(role) };
    case "MAHASISWA":
      return { variant: "draft", label: roleLabel(role) };
    default:
      return { variant: "draft", label: roleLabel(role) };
  }
}
