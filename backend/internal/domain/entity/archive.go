package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ThesisArchive struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ThesisID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"thesis_id"`
	Thesis         Thesis         `gorm:"foreignKey:ThesisID" json:"thesis"`
	FileURL        string         `gorm:"type:text;not null" json:"-"`
	FileName       string         `gorm:"type:varchar(255);not null" json:"file_name"`
	AbstractID     string         `gorm:"type:text;not null" json:"abstract_id"`
	AbstractEN     *string        `gorm:"type:text" json:"abstract_en,omitempty"`
	Keywords       pq.StringArray `gorm:"type:text[]" json:"keywords"`
	GraduationYear int            `gorm:"not null" json:"graduation_year"`
	ArchivedBy     uuid.UUID      `gorm:"type:uuid;not null" json:"archived_by"`
	Archiver       User           `gorm:"foreignKey:ArchivedBy" json:"archiver"`
	ArchivedAt     time.Time      `gorm:"not null;default:now()" json:"archived_at"`
	SearchVector   *string        `gorm:"type:tsvector;<-:false" json:"-"` // read-only, updated by trigger
	CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;default:now()" json:"updated_at"`
}

func (ThesisArchive) TableName() string { return "thesis_archives" }
