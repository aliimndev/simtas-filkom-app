// MANUAL mirror of the live Postgres schema (camelCase props -> snake_case columns).
// ponytail: we hand-mirror because `drizzle-kit pull` emits snake_case property names that
// break the camelCase query code; regenerate only if we switch the whole codebase to snake_case.
// Introspected 2026-08-28 from postgres://postgres@localhost:5433/simtas (all migrations applied).

import {
  pgTable, uuid, varchar, timestamp, boolean, integer, text, serial,
  date, numeric, bigint, jsonb, inet,
} from "drizzle-orm/pg-core";

// ── roles ──
export const roles = pgTable("roles", {
  id: serial("id").primaryKey(),
  name: varchar("name", { length: 50 }).notNull(),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── users ──
export const users = pgTable("users", {
  id: uuid("id").primaryKey().defaultRandom(),
  email: varchar("email", { length: 255 }).notNull(),
  passwordHash: varchar("password_hash", { length: 255 }).notNull(),
  fullName: varchar("full_name", { length: 255 }).notNull(),
  nimNidn: varchar("nim_nidn", { length: 50 }),
  roleId: integer("role_id").notNull(),
  studyProgram: varchar("study_program", { length: 100 }),
  profilePhotoUrl: text("profile_photo_url"),
  isActive: boolean("is_active").notNull().default(true),
  mustChangePassword: boolean("must_change_password").notNull().default(true),
  loginAttemptCount: integer("login_attempt_count").notNull().default(0),
  lockedUntil: timestamp("locked_until", { withTimezone: true }),
  lastLoginAt: timestamp("last_login_at", { withTimezone: true }),
  tokenVersion: integer("token_version").notNull().default(0),
  placeOfBirth: varchar("place_of_birth", { length: 100 }),
  address: text("address"),
  phone: varchar("phone", { length: 30 }),
  birthDate: varchar("birth_date", { length: 20 }),
  faculty: varchar("faculty", { length: 100 }),
  semester: integer("semester"),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
  deletedAt: timestamp("deleted_at", { withTimezone: true }),
});

// ── refresh_token_families ──
export const refreshTokenFamilies = pgTable("refresh_token_families", {
  id: uuid("id").primaryKey().defaultRandom(),
  userId: uuid("user_id").notNull(),
  familyId: uuid("family_id").notNull(),
  tokenJti: varchar("token_jti", { length: 255 }).notNull(),
  expiresAt: timestamp("expires_at", { withTimezone: true }).notNull(),
  rotatedAt: timestamp("rotated_at", { withTimezone: true }),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── token_blacklist (access-token revocation) ──
export const tokenBlacklist = pgTable("token_blacklist", {
  id: uuid("id").primaryKey().defaultRandom(),
  tokenJti: varchar("token_jti", { length: 255 }).notNull(),
  expiresAt: timestamp("expires_at", { withTimezone: true }).notNull(),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── password_reset_tokens ──
export const passwordResetTokens = pgTable("password_reset_tokens", {
  id: uuid("id").primaryKey().defaultRandom(),
  userId: uuid("user_id").notNull(),
  token: varchar("token", { length: 255 }).notNull(),
  expiresAt: timestamp("expires_at", { withTimezone: true }).notNull(),
  usedAt: timestamp("used_at", { withTimezone: true }),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── academic_years ──
export const academicYears = pgTable("academic_years", {
  id: uuid("id").primaryKey().defaultRandom(),
  name: varchar("name", { length: 20 }).notNull(),
  semester: varchar("semester", { length: 10 }).notNull(),
  startDate: date("start_date").notNull(),
  endDate: date("end_date").notNull(),
  isActive: boolean("is_active").notNull().default(false),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── theses ──
export const theses = pgTable("theses", {
  id: uuid("id").primaryKey().defaultRandom(),
  studentId: uuid("student_id").notNull(),
  academicYearId: uuid("academic_year_id").notNull(),
  title: varchar("title", { length: 500 }).notNull(),
  abstract: text("abstract"),
  fieldOfStudy: varchar("field_of_study", { length: 100 }),
  thesisType: varchar("thesis_type", { length: 20 }).notNull(),
  status: varchar("status", { length: 30 }).notNull().default("submitted"),
  kaprodiNotes: text("kaprodi_notes"),
  submittedAt: timestamp("submitted_at", { withTimezone: true }).notNull().defaultNow(),
  approvedAt: timestamp("approved_at", { withTimezone: true }),
  graduatedAt: timestamp("graduated_at", { withTimezone: true }),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
  deletedAt: timestamp("deleted_at", { withTimezone: true }),
});

// ── thesis_supervisors (join) ──
export const thesisSupervisors = pgTable("thesis_supervisors", {
  id: uuid("id").primaryKey().defaultRandom(),
  thesisId: uuid("thesis_id").notNull(),
  supervisorId: uuid("supervisor_id").notNull(),
  assignedAt: timestamp("assigned_at", { withTimezone: true }).notNull().defaultNow(),
  assignedBy: uuid("assigned_by").notNull(),
});

// ── title_change_requests ──
export const titleChangeRequests = pgTable("title_change_requests", {
  id: uuid("id").primaryKey().defaultRandom(),
  thesisId: uuid("thesis_id").notNull(),
  requestedById: uuid("requested_by_id").notNull(),
  previousTitle: text("previous_title").notNull(),
  requestedTitle: text("requested_title").notNull(),
  reason: text("reason"),
  status: varchar("status", { length: 20 }).notNull().default("PENDING"),
  reviewedById: uuid("reviewed_by_id"),
  reviewedAt: timestamp("reviewed_at", { withTimezone: true }),
  reviewNotes: text("review_notes"),
  cancelledById: uuid("cancelled_by_id"),
  cancelledAt: timestamp("cancelled_at", { withTimezone: true }),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
  deletedAt: timestamp("deleted_at", { withTimezone: true }),
});

// ── consultation_logs ──
export const consultationLogs = pgTable("consultation_logs", {
  id: uuid("id").primaryKey().defaultRandom(),
  thesisId: uuid("thesis_id").notNull(),
  createdBy: uuid("created_by").notNull(),
  consultationDate: date("consultation_date").notNull(),
  topicsDiscussed: text("topics_discussed").notNull(),
  notes: text("notes"),
  followUp: text("follow_up"),
  attachmentUrl: text("attachment_url"),
  status: varchar("status", { length: 20 }).notNull().default("pending"),
  approvedBy: uuid("approved_by"),
  approvedAt: timestamp("approved_at", { withTimezone: true }),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── seminars ──
export const seminars = pgTable("seminars", {
  id: uuid("id").primaryKey().defaultRandom(),
  thesisId: uuid("thesis_id").notNull(),
  status: varchar("status", { length: 20 }).notNull().default("pending"),
  scheduledAt: timestamp("scheduled_at", { withTimezone: true }),
  room: varchar("room", { length: 100 }),
  notes: text("notes"),
  finalScore: numeric("final_score"),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── seminar_examiners (join) ──
export const seminarExaminers = pgTable("seminar_examiners", {
  id: uuid("id").primaryKey().defaultRandom(),
  seminarId: uuid("seminar_id").notNull(),
  examinerId: uuid("examiner_id").notNull(),
  assignedBy: uuid("assigned_by").notNull(),
});

// ── seminar_scores ──
export const seminarScores = pgTable("seminar_scores", {
  id: uuid("id").primaryKey().defaultRandom(),
  seminarId: uuid("seminar_id").notNull(),
  examinerId: uuid("examiner_id").notNull(),
  componentName: varchar("component_name", { length: 100 }).notNull(),
  componentWeight: numeric("component_weight").notNull(),
  score: numeric("score").notNull(),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── thesis_defenses ──
export const thesisDefenses = pgTable("thesis_defenses", {
  id: uuid("id").primaryKey().defaultRandom(),
  thesisId: uuid("thesis_id").notNull(),
  status: varchar("status", { length: 20 }).notNull().default("pending"),
  scheduledAt: timestamp("scheduled_at", { withTimezone: true }),
  room: varchar("room", { length: 100 }),
  revisionNotes: text("revision_notes"),
  finalScore: numeric("final_score"),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── defense_examiners (join) ──
export const defenseExaminers = pgTable("defense_examiners", {
  id: uuid("id").primaryKey().defaultRandom(),
  defenseId: uuid("defense_id").notNull(),
  examinerId: uuid("examiner_id").notNull(),
  assignedBy: uuid("assigned_by").notNull(),
});

// ── defense_scores ──
export const defenseScores = pgTable("defense_scores", {
  id: uuid("id").primaryKey().defaultRandom(),
  defenseId: uuid("defense_id").notNull(),
  examinerId: uuid("examiner_id").notNull(),
  componentName: varchar("component_name", { length: 100 }).notNull(),
  componentWeight: numeric("component_weight").notNull(),
  score: numeric("score").notNull(),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── documents ──
export const documents = pgTable("documents", {
  id: uuid("id").primaryKey().defaultRandom(),
  thesisId: uuid("thesis_id").notNull(),
  uploadedBy: uuid("uploaded_by").notNull(),
  documentType: varchar("document_type", { length: 50 }).notNull(),
  chapterNumber: integer("chapter_number"),
  version: integer("version").notNull().default(1),
  fileName: varchar("file_name", { length: 255 }).notNull(),
  fileUrl: text("file_url").notNull(),
  fileSize: bigint("file_size", { mode: "number" }),
  status: varchar("status", { length: 30 }).notNull().default("pending_review"),
  reviewerId: uuid("reviewer_id"),
  reviewerNotes: text("reviewer_notes"),
  reviewedAt: timestamp("reviewed_at", { withTimezone: true }),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── thesis_archives ──
export const thesisArchives = pgTable("thesis_archives", {
  id: uuid("id").primaryKey().defaultRandom(),
  thesisId: uuid("thesis_id").notNull(),
  fileUrl: text("file_url").notNull(),
  fileName: varchar("file_name", { length: 255 }).notNull(),
  abstractId: text("abstract_id").notNull(),
  abstractEn: text("abstract_en"),
  keywords: text("keywords").array(),
  graduationYear: integer("graduation_year").notNull(),
  archivedBy: uuid("archived_by").notNull(),
  archivedAt: timestamp("archived_at", { withTimezone: true }).notNull().defaultNow(),
  searchVector: text("search_vector"),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── notifications ──
export const notifications = pgTable("notifications", {
  id: uuid("id").primaryKey().defaultRandom(),
  userId: uuid("user_id").notNull(),
  title: varchar("title", { length: 255 }).notNull(),
  message: text("message").notNull(),
  type: varchar("type", { length: 50 }).notNull(),
  link: varchar("link", { length: 500 }),
  isRead: boolean("is_read").notNull().default(false),
  readAt: timestamp("read_at", { withTimezone: true }),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── email_logs ──
export const emailLogs = pgTable("email_logs", {
  id: uuid("id").primaryKey().defaultRandom(),
  recipientEmail: varchar("recipient_email", { length: 255 }).notNull(),
  eventType: varchar("event_type", { length: 50 }).notNull(),
  subject: varchar("subject", { length: 255 }),
  status: varchar("status", { length: 20 }).notNull().default("sent"),
  provider: varchar("provider", { length: 50 }).notNull().default("resend"),
  errorMessage: text("error_message"),
  body: text("body"),
  attempts: integer("attempts").notNull().default(0),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});

// ── audit_logs ──
export const auditLogs = pgTable("audit_logs", {
  id: uuid("id").primaryKey().defaultRandom(),
  userId: uuid("user_id"),
  action: varchar("action", { length: 100 }).notNull(),
  entityType: varchar("entity_type", { length: 100 }),
  entityId: uuid("entity_id"),
  oldValue: jsonb("old_value"),
  newValue: jsonb("new_value"),
  ipAddress: inet("ip_address"),
  userAgent: text("user_agent"),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});
