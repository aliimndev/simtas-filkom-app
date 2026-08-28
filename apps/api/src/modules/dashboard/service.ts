import { sql } from "drizzle-orm";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { schema } from "@sims/db";

// StatusLabel maps thesis status → display label (mirrors Go usecase.StatusLabel).
const StatusLabel: Record<string, string> = {
  submitted: "Menunggu Review",
  approved: "Judul Disetujui",
  in_progress: "Bimbingan",
  seminar_ready: "Siap Seminar",
  seminar_done: "Pasca Seminar",
  defense_ready: "Siap Sidang",
  defense_done: "Pasca Sidang",
  graduated: "Lulus",
  cancelled: "Dibatalkan",
};

type SummaryRow = Record<string, unknown>;

// postgres-js execute returns a RowList (array) that also exposes `.rows` for some
// driver versions; normalize to a plain array either way (ponytail: one helper).
async function run(db: any, query: any): Promise<any[]> {
  const res: any = await db.execute(query);
  return Array.isArray(res) ? res : res.rows ?? [];
}

export interface DashboardSummaryResponse {
  academic_summary: { total_active: number; total_graduated: number; avg_completion_months: number };
  by_status: { status: string; label: string; count: number }[];
  graduation_trend: { month: string; count: number }[];
}

// Summary mirrors Go dashboard_usecase.Summary / dashboard_repository_impl
// (GetAcademicSummary + GetThesisByStatus + GetGraduationTrend), no optional filters.
export async function getSummary(): Promise<DashboardSummaryResponse> {
  const db = getDb(loadConfig().databaseUrl);

  const academic: any[] = await run(
    db,
    sql`
      SELECT
        COUNT(*) FILTER (WHERE theses.status NOT IN ('graduated','cancelled')) AS total_active,
        COUNT(*) FILTER (WHERE theses.status = 'graduated') AS total_graduated,
        COALESCE(AVG(EXTRACT(EPOCH FROM (theses.graduated_at - theses.submitted_at)) / 2592000.0), 0) AS avg_completion_months
      FROM theses
      JOIN academic_years ay ON ay.id = theses.academic_year_id
      JOIN users su ON su.id = theses.student_id
      WHERE theses.deleted_at IS NULL
    `,
  );

  const byStatus: any[] = await run(
    db,
    sql`
      SELECT theses.status AS status, COUNT(*) AS count
      FROM theses
      JOIN academic_years ay ON ay.id = theses.academic_year_id
      JOIN users su ON su.id = theses.student_id
      WHERE theses.deleted_at IS NULL
      GROUP BY theses.status
      ORDER BY theses.status
    `,
  );

  const trend: any[] = await run(
    db,
    sql`
      SELECT TO_CHAR(theses.graduated_at, 'YYYY-MM') AS month, COUNT(*) AS count
      FROM theses
      JOIN academic_years ay ON ay.id = theses.academic_year_id
      JOIN users su ON su.id = theses.student_id
      WHERE theses.graduated_at IS NOT NULL AND theses.deleted_at IS NULL
      GROUP BY month
      ORDER BY month DESC
      LIMIT 12
    `,
  );

  const a = academic[0] ?? {};
  const graduationTrend = (trend as SummaryRow[])
    .map((r) => ({ month: String(r.month), count: Number(r.count) }))
    .reverse(); // ascending for charts (mirrors Go's in-place reversal)

  return {
    academic_summary: {
      total_active: Number(a.total_active ?? 0),
      total_graduated: Number(a.total_graduated ?? 0),
      avg_completion_months: Number(a.avg_completion_months ?? 0),
    },
    by_status: (byStatus as SummaryRow[]).map((r) => {
      const status = String(r.status);
      const label = StatusLabel[status] ?? status;
      return { status, label, count: Number(r.count) };
    }),
    graduation_trend: graduationTrend,
  };
}
