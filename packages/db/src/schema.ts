// GENERATED via `drizzle-kit pull` — do not hand-edit to "fix" design.
// In CI this file is regenerated from live Postgres. Manual edits here are deferred design fixes.
// Last introspected: 2026-08-28 (manual mirror based on backend/migrations 000001-000015).
// ponytail: manual mirror — replace via `bunx drizzle-kit pull` when DB is available.

import { pgTable, uuid, varchar, timestamp, boolean, integer, text, serial } from "drizzle-orm/pg-core";

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

// ── academic_years ──
export const academicYears = pgTable("academic_years", {
  id: uuid("id").primaryKey().defaultRandom(),
  name: varchar("name", { length: 20 }).notNull(),
  semester: varchar("semester", { length: 10 }).notNull(),
  startDate: timestamp("start_date", { withTimezone: true }).notNull(),
  endDate: timestamp("end_date", { withTimezone: true }).notNull(),
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
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: true }).notNull().defaultNow(),
  deletedAt: timestamp("deleted_at", { withTimezone: true }),
});

// ── notifications ──
export const notifications = pgTable("notifications", {
  id: uuid("id").primaryKey().defaultRandom(),
  userId: uuid("user_id").notNull(),
  title: varchar("title", { length: 255 }).notNull(),
  message: text("message"),
  isRead: boolean("is_read").notNull().default(false),
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});
