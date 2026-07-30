package entity

import (
	"time"

	"github.com/google/uuid"
)

type ThesisDefense struct {
	ID            uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ThesisID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"thesis_id"`
	Thesis        Thesis       `gorm:"foreignKey:ThesisID" json:"thesis"`
	Status        string       `gorm:"type:varchar(30);not null;default:pending;index" json:"status"`
	ScheduledAt   *time.Time   `gorm:"index" json:"scheduled_at,omitempty"`
	Room          *string      `gorm:"type:varchar(100)" json:"room,omitempty"`
	RevisionNotes *string      `gorm:"type:text" json:"revision_notes,omitempty"`
	FinalScore    *float64     `gorm:"type:decimal(5,2)" json:"final_score,omitempty"`
	CreatedAt     time.Time    `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time    `gorm:"not null;default:now()" json:"updated_at"`

	// Associations
	Examiners []User         `gorm:"many2many:defense_examiners;joinForeignKey:DefenseID;joinReferences:ExaminerID" json:"examiners,omitempty"`
	Scores    []DefenseScore `gorm:"foreignKey:DefenseID" json:"scores,omitempty"`
}

func (ThesisDefense) TableName() string { return "thesis_defenses" }

type DefenseExaminer struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DefenseID  uuid.UUID `gorm:"type:uuid;not null;index" json:"defense_id"`
	ExaminerID uuid.UUID `gorm:"type:uuid;not null" json:"examiner_id"`
	Examiner   User      `gorm:"foreignKey:ExaminerID" json:"examiner"`
	AssignedBy uuid.UUID `gorm:"type:uuid;not null" json:"assigned_by"`
}

func (DefenseExaminer) TableName() string { return "defense_examiners" }

type DefenseScore struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DefenseID       uuid.UUID `gorm:"type:uuid;not null;index" json:"defense_id"`
	ExaminerID      uuid.UUID `gorm:"type:uuid;not null" json:"examiner_id"`
	Examiner        User      `gorm:"foreignKey:ExaminerID" json:"examiner"`
	ComponentName   string    `gorm:"type:varchar(100);not null" json:"component_name"`
	ComponentWeight float64   `gorm:"type:decimal(5,2);not null" json:"component_weight"`
	Score           float64   `gorm:"type:decimal(5,2);not null" json:"score"`
	CreatedAt       time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (DefenseScore) TableName() string { return "defense_scores" }
