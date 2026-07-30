package entity

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ThesisID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"thesis_id"`
	Thesis        Thesis     `gorm:"foreignKey:ThesisID" json:"-"`
	UploadedBy    uuid.UUID  `gorm:"type:uuid;not null" json:"uploaded_by"`
	Uploader      User       `gorm:"foreignKey:UploadedBy" json:"uploader"`
	DocumentType  string     `gorm:"type:varchar(50);not null;index" json:"document_type"`
	ChapterNumber *int       `json:"chapter_number,omitempty"`
	Version       int        `gorm:"not null;default:1" json:"version"`
	FileName      string     `gorm:"type:varchar(255);not null" json:"file_name"`
	FileURL       string     `gorm:"type:text;not null" json:"-"`
	FileSize      *int64     `json:"file_size,omitempty"`
	Status        string     `gorm:"type:varchar(30);not null;default:pending_review;index" json:"status"`
	ReviewerID    *uuid.UUID `gorm:"type:uuid" json:"reviewer_id,omitempty"`
	Reviewer      *User      `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	ReviewerNotes *string    `gorm:"type:text" json:"reviewer_notes,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

func (Document) TableName() string { return "documents" }
