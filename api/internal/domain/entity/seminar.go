package entity

import (
	"time"

	"github.com/google/uuid"
)

type Seminar struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ThesisID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"thesis_id"`
	Thesis      Thesis     `gorm:"foreignKey:ThesisID" json:"thesis"`
	Status      string     `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`
	ScheduledAt *time.Time `gorm:"index" json:"scheduled_at,omitempty"`
	Room        *string    `gorm:"type:varchar(100)" json:"room,omitempty"`
	Notes       *string    `gorm:"type:text" json:"notes,omitempty"`
	FinalScore  *float64   `gorm:"type:decimal(5,2)" json:"final_score,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;default:now()" json:"updated_at"`

	// Associations
	Examiners []User         `gorm:"many2many:seminar_examiners;joinForeignKey:SeminarID;joinReferences:ExaminerID" json:"examiners,omitempty"`
	Scores    []SeminarScore `gorm:"foreignKey:SeminarID" json:"scores,omitempty"`
}

func (Seminar) TableName() string { return "seminars" }

type SeminarExaminer struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SeminarID  uuid.UUID `gorm:"type:uuid;not null;index" json:"seminar_id"`
	ExaminerID uuid.UUID `gorm:"type:uuid;not null" json:"examiner_id"`
	Examiner   User      `gorm:"foreignKey:ExaminerID" json:"examiner"`
	AssignedBy uuid.UUID `gorm:"type:uuid;not null" json:"assigned_by"`
}

func (SeminarExaminer) TableName() string { return "seminar_examiners" }

type SeminarScore struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SeminarID       uuid.UUID `gorm:"type:uuid;not null;index" json:"seminar_id"`
	ExaminerID      uuid.UUID `gorm:"type:uuid;not null" json:"examiner_id"`
	Examiner        User      `gorm:"foreignKey:ExaminerID" json:"examiner"`
	ComponentName   string    `gorm:"type:varchar(100);not null" json:"component_name"`
	ComponentWeight float64   `gorm:"type:decimal(5,2);not null" json:"component_weight"`
	Score           float64   `gorm:"type:decimal(5,2);not null" json:"score"`
	CreatedAt       time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (SeminarScore) TableName() string { return "seminar_scores" }
