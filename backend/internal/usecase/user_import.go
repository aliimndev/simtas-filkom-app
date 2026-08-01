package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/mail"
	"path/filepath"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
)

const (
	MaxImportFileSize = 5 * 1024 * 1024 // 5 MB

	importTemplateSheet   = "Data"
	importInstructionsShe = "Petunjuk"
)

var (
	ErrInvalidFileFormat = errors.New("format file tidak didukung, gunakan .csv atau .xlsx")
	ErrFileTooLarge      = errors.New("ukuran file melebihi 5 MB")
)

// ImportRow is one parsed row from the uploaded file.
type ImportRow struct {
	Row          int // 1-based file row (header = row 1)
	Email        string
	FullName     string
	NimNidn      string
	Role         string
	StudyProgram string
}

// ImportError describes a skipped row.
type ImportError struct {
	Row    int    `json:"row"`
	Email  string `json:"email"`
	Reason string `json:"reason"`
}

// ImportResult is returned by the import endpoint.
type ImportResult struct {
	TotalRows    int           `json:"total_rows"`
	SuccessCount int           `json:"success_count"`
	ErrorCount   int           `json:"error_count"`
	Errors       []ImportError `json:"errors"`
}

// ImportUsers parses and validates the uploaded file, bulk-creates valid rows
// in a single transaction, sends welcome emails, and records an audit log.
func (uc *UserUseCase) ImportUsers(ctx context.Context, filename string, data []byte, actor Actor) (*ImportResult, error) {
	if len(data) > MaxImportFileSize {
		return nil, ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(filename))
	var rows []ImportRow
	var err error
	switch ext {
	case ".csv":
		rows, err = parseCSV(data)
	case ".xlsx":
		rows, err = parseXLSX(data)
	default:
		return nil, ErrInvalidFileFormat
	}
	if err != nil {
		return nil, err
	}

	validUsers := make([]*entity.User, 0, len(rows))
	importErrors := make([]ImportError, 0)
	seenEmails := make(map[string]bool)
	emailToPassword := make(map[string]string)

	for _, row := range rows {
		reason := validateImportRow(row, seenEmails)

		// Validate role BEFORE the expensive bcrypt hashing and the DB email check.
		// Cache the role lookup result so it can be reused when building the user.
		var role *entity.Role
		if reason == "" {
			role, err = uc.userRepo.FindRoleByName(ctx, row.Role)
			if err != nil {
				reason = ErrRoleInvalid.Error()
			}
		}
		if reason == "" {
			// Email belum terdaftar? (spec: per-row validation, skip on error)
			existing, err := uc.userRepo.FindByEmail(ctx, row.Email)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			if existing != nil {
				reason = "Email sudah terdaftar"
			}
		}
		if reason != "" {
			importErrors = append(importErrors, ImportError{
				Row:    row.Row,
				Email:  row.Email,
				Reason: reason,
			})
			continue
		}
		seenEmails[row.Email] = true

		tempPassword := utils.GenerateRandomPassword(12)
		hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), 12)
		if err != nil {
			return nil, err
		}

		var nimNidn, studyProgram *string
		if row.NimNidn != "" {
			nimNidn = &row.NimNidn
		}
		if row.StudyProgram != "" {
			studyProgram = &row.StudyProgram
		}

		user := &entity.User{
			Email:              row.Email,
			PasswordHash:       string(hash),
			FullName:           row.FullName,
			NimNidn:            nimNidn,
			RoleID:             role.ID,
			Role:               *role,
			StudyProgram:       studyProgram,
			IsActive:           true,
			MustChangePassword: true,
		}
		validUsers = append(validUsers, user)
		emailToPassword[user.Email] = tempPassword
	}

	result := &ImportResult{
		TotalRows:    len(rows),
		SuccessCount: len(validUsers),
		ErrorCount:   len(importErrors),
		Errors:       importErrors,
	}

	if len(validUsers) > 0 {
		if err := uc.userRepo.BulkCreate(ctx, validUsers); err != nil {
			return nil, err
		}
		// Welcome emails (async, non-fatal)
		for _, u := range validUsers {
			u := u
			go func() {
				_ = uc.emailSvc.SendWelcomeEmail(context.Background(), u.Email, u.FullName, emailToPassword[u.Email])
			}()
		}
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionUserBulkImported,
		EntityType: "user",
		NewValue: map[string]interface{}{
			"total_rows":    result.TotalRows,
			"success_count": result.SuccessCount,
			"error_count":   result.ErrorCount,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})

	return result, nil
}

// BuildImportTemplate returns an .xlsx template with a "Data" sheet and a
// "Petunjuk" sheet describing valid role values.
func (uc *UserUseCase) BuildImportTemplate() ([]byte, error) {
	f := excelize.NewFile()

	dataSheet := importTemplateSheet
	// Rename default sheet to "Data"
	f.SetSheetName("Sheet1", dataSheet)

	headers := []string{"email", "full_name", "nim_nidn", "role", "study_program"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(dataSheet, cell, h)
	}

	// Instructions sheet
	_, _ = f.NewSheet(importInstructionsShe)
	instructions := []string{
		"Petunjuk Pengisian Template Import User",
		"",
		"1. Kolom wajib: email, full_name, role.",
		"2. Kolom opsional: nim_nidn, study_program.",
		"3. Nilai valid untuk kolom role:",
		"   - admin_fakultas",
		"   - kaprodi",
		"   - mahasiswa",
		"   - dosen_pembimbing",
		"   - dosen_penguji",
		"4. Email harus unik dan belum terdaftar di sistem.",
		"5. Baris pertama (header) wajib dipertahankan; isi data mulai baris kedua.",
	}
	for i, line := range instructions {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		_ = f.SetCellValue(importInstructionsShe, cell, line)
	}
	_ = f.SetColWidth(importInstructionsShe, "A", "A", 80)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// csvUserRow mirrors the import template's CSV header (row 1) — mapped by header name.
type csvUserRow struct {
	Email        string `csv:"email"`
	FullName     string `csv:"full_name"`
	NimNidn      string `csv:"nim_nidn"`
	Role         string `csv:"role"`
	StudyProgram string `csv:"study_program"`
}

func parseCSV(data []byte) ([]ImportRow, error) {
	var records []csvUserRow
	if err := gocsv.UnmarshalBytes(data, &records); err != nil {
		return nil, fmt.Errorf("gagal membaca file CSV: %w", err)
	}

	rows := make([]ImportRow, 0, len(records))
	for i, rec := range records {
		rows = append(rows, ImportRow{
			Row:          i + 2, // row 1 is the header
			Email:        strings.TrimSpace(rec.Email),
			FullName:     strings.TrimSpace(rec.FullName),
			NimNidn:      strings.TrimSpace(rec.NimNidn),
			Role:         strings.TrimSpace(rec.Role),
			StudyProgram: strings.TrimSpace(rec.StudyProgram),
		})
	}
	return rows, nil
}

func parseXLSX(data []byte) ([]ImportRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file Excel: %w", err)
	}
	defer f.Close()

	sheet := importTemplateSheet
	rows, err := f.GetRows(sheet)
	if err != nil {
		// Fallback: first sheet
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, ErrInvalidFileFormat
		}
		sheet = sheets[0]
		rows, err = f.GetRows(sheet)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca sheet Excel: %w", err)
		}
	}
	if len(rows) < 2 {
		return nil, nil
	}

	result := make([]ImportRow, 0, len(rows)-1)
	for i, rec := range rows[1:] {
		result = append(result, ImportRow{
			Row:          i + 2,
			Email:        cell(rec, 0),
			FullName:     cell(rec, 1),
			NimNidn:      cell(rec, 2),
			Role:         cell(rec, 3),
			StudyProgram: cell(rec, 4),
		})
	}
	return result, nil
}

func cell(record []string, idx int) string {
	if idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

// validateImportRow returns an error reason string, or "" if the row is valid.
func validateImportRow(row ImportRow, seenEmails map[string]bool) string {
	if row.Email == "" {
		return "Email tidak boleh kosong"
	}
	if _, err := mail.ParseAddress(row.Email); err != nil {
		return "Format email tidak valid"
	}
	if seenEmails[row.Email] {
		return "Email duplikat dalam file"
	}
	if row.FullName == "" {
		return "Nama lengkap tidak boleh kosong"
	}
	if row.Role == "" {
		return "Role tidak boleh kosong"
	}
	return ""
}
