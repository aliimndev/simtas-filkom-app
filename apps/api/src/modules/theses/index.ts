export { thesesRoutes } from "./routes";
export {
  ThesesError,
  createThesis,
  listTheses,
  getThesis,
  reviewThesis,
  assignSupervisors,
  softDeleteThesis,
} from "./service";
export type { Actor as ThesisActor, ThesisDetail } from "./types";
