export { defensesRoutes, defenseSubmissionRoutes, suratTugasRoutes } from "./routes";
export { cancelSuratTugas, createSuratTugas, finalizeDefense, getDefense, issueSuratTugas, listDefenses, saveDefenseScores, scheduleDefense, submitDefense } from "./service";
export type { DefenseActor, DefenseFinalizeInput, DefenseListFilter, DefenseScoreInput, ScheduleDefenseInput, SuratTugasInput } from "./types";
