export const ADMIN_FAKULTAS = "ADMIN_FAKULTAS";
export const ROLE_KAPRODI = "KAPRODI";
export const ROLE_DOSEN_PEMBIMBING = "DOSEN_PEMBIMBING";
export const ROLE_MAHASISWA = "MAHASISWA";
export type RoleName = typeof ADMIN_FAKULTAS | typeof ROLE_KAPRODI | typeof ROLE_DOSEN_PEMBIMBING | typeof ROLE_MAHASISWA;
export function isRoleName(v: unknown): v is RoleName {
  return [ADMIN_FAKULTAS, ROLE_KAPRODI, ROLE_DOSEN_PEMBIMBING, ROLE_MAHASISWA].includes(v as RoleName);
}
