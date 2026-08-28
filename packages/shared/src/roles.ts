export const ADMIN_FAKULTAS = "ADMIN_FAKULTAS";
export const KAPRODI = "KAPRODI";
export const DOSEN_PEMBIMBING = "DOSEN_PEMBIMBING";
export const DOSEN_PENGUJI = "DOSEN_PENGUJI";
export const MAHASISWA = "MAHASISWA";
export type RoleName =
  | typeof ADMIN_FAKULTAS
  | typeof KAPRODI
  | typeof DOSEN_PEMBIMBING
  | typeof DOSEN_PENGUJI
  | typeof MAHASISWA;
export function isRoleName(v: unknown): v is RoleName {
  return [ADMIN_FAKULTAS, KAPRODI, DOSEN_PEMBIMBING, DOSEN_PENGUJI, MAHASISWA].includes(v as RoleName);
}
