export { seminarsRoutes, seminarSubmissionRoutes } from "./routes";
export {
  ACTIVE_SEMINAR_STATUSES,
  SeminarError,
  getSeminar,
  listSeminars,
  scheduleSeminar,
  submitSeminar,
} from "./service";
export type { SeminarActor, SeminarListFilter, ScheduleSeminarInput } from "./types";
