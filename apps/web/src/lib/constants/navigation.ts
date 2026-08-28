import {
  LayoutDashboard,
  FileText,
  MessagesSquare,
  FolderOpen,
  GraduationCap,
  Users,
  CalendarRange,
  Archive,
  BookOpen,
  ClipboardCheck,
  ShieldCheck,
  FilePen,
} from "lucide-svelte";

type IconComponent = typeof FileText;

export interface NavItem {
  href: string;
  label: string;
  icon: IconComponent;
  roles?: string[];
}

export interface NavSection {
  title?: string;
  items: NavItem[];
}

export const NAV_SECTIONS: NavSection[] = [
  {
    items: [{ href: "/dashboard", label: "Dashboard", icon: LayoutDashboard }],
  },
  {
    title: "Tugas Akhir Skripsi",
    items: [
      { href: "/thesis", label: "Tugas Akhir Skripsi Saya", icon: FileText, roles: ["mahasiswa"] },
      { href: "/theses", label: "Daftar Tugas Akhir Skripsi", icon: BookOpen, roles: ["admin_fakultas", "kaprodi"] },
      { href: "/supervision", label: "Bimbingan", icon: MessagesSquare, roles: ["mahasiswa", "dosen_pembimbing"] },
      { href: "/documents", label: "Dokumen", icon: FolderOpen, roles: ["mahasiswa", "dosen_pembimbing"] },
      {
        href: "/title-change-reviews",
        label: "Review Perubahan Judul",
        icon: FilePen,
        roles: ["dosen_pembimbing"],
      },
    ],
  },
  {
    title: "Ujian",
    items: [
      {
        href: "/seminars",
        label: "Seminar",
        icon: GraduationCap,
        roles: ["admin_fakultas", "kaprodi", "dosen_penguji", "mahasiswa", "dosen_pembimbing"],
      },
      {
        href: "/defenses",
        label: "Sidang",
        icon: ClipboardCheck,
        roles: ["admin_fakultas", "kaprodi", "dosen_penguji", "mahasiswa", "dosen_pembimbing"],
      },
      { href: "/schedules", label: "Jadwal", icon: CalendarRange, roles: ["admin_fakultas", "kaprodi"] },
    ],
  },
  {
    title: "Administrasi",
    items: [
      { href: "/admin/users", label: "Manajemen Pengguna", icon: Users, roles: ["admin_fakultas"] },
      { href: "/admin/academic-years", label: "Tahun Akademik", icon: CalendarRange, roles: ["admin_fakultas"] },
      { href: "/admin/audit-logs", label: "Audit Log", icon: ShieldCheck, roles: ["admin_fakultas"] },
      { href: "/archives", label: "Arsip", icon: Archive, roles: ["admin_fakultas", "kaprodi"] },
    ],
  },
];

export function navItemsForRoles(roles: string[]): NavSection[] {
  if (!roles || roles.length === 0) return [];
  const norm = roles.map((r) => r.toLowerCase());
  const has = (itemRoles?: string[]) =>
    !itemRoles ||
    itemRoles.length === 0 ||
    itemRoles.some((r) => norm.includes(r.toLowerCase()));
  return NAV_SECTIONS
    .map((s) => ({ ...s, items: s.items.filter((i) => has(i.roles)) }))
    .filter((s) => s.items.length > 0);
}

export const APP_NAME = "SIMTAS FILKOM";

const ROLE_LABELS: Record<string, string> = {
  ADMIN_FAKULTAS: "Admin Fakultas",
  KAPRODI: "Kaprodi",
  MAHASISWA: "Mahasiswa",
  DOSEN_PEMBIMBING: "Dosen Pembimbing",
  DOSEN_PENGUJI: "Dosen Penguji",
  admin_fakultas: "Admin Fakultas",
  kaprodi: "Kaprodi",
  mahasiswa: "Mahasiswa",
  dosen_pembimbing: "Dosen Pembimbing",
  dosen_penguji: "Dosen Penguji",
};

export function roleLabel(role?: string | null): string {
  if (!role) return "—";
  return ROLE_LABELS[role] ?? ROLE_LABELS[role.toUpperCase()] ?? role;
}
