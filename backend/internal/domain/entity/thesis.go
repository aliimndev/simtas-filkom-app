package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Thesis struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StudentID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"student_id"`
	Student        User           `gorm:"foreignKey:StudentID" json:"student"`
	AcademicYearID uuid.UUID      `gorm:"type:uuid;not null;index" json:"academic_year_id"`
	AcademicYear   AcademicYear   `gorm:"foreignKey:AcademicYearID" json:"academic_year"`
	Title          string         `gorm:"type:varchar(500);not null" json:"title"`
	Abstract       *string        `gorm:"type:text" json:"abstract,omitempty"`
	FieldOfStudy   *string        `gorm:"type:varchar(100)" json:"field_of_study,omitempty"`
	ThesisType     string         `gorm:"type:varchar(20);not null" json:"thesis_type"`
	Status         string         `gorm:"type:varchar(30);not null;default:submitted" json:"status"`
	KaprodiNotes   *string        `gorm:"type:text" json:"kaprodi_notes,omitempty"`
	SubmittedAt    time.Time      `gorm:"not null;default:now()" json:"submitted_at"`
	ApprovedAt     *time.Time     `json:"approved_at,omitempty"`
	GraduatedAt    *time.Time     `json:"graduated_at,omitempty"`
	CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	Supervisors []User `gorm:"many2many:thesis_supervisors;joinForeignKey:ThesisID;joinReferences:SupervisorID" json:"supervisors,omitempty"`
}

func (Thesis) TableName() string { return "theses" }

type ThesisSupervisor struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ThesisID     uuid.UUID `gorm:"type:uuid;not null;index" json:"thesis_id"`
	Thesis       Thesis    `gorm:"foreignKey:ThesisID" json:"-"`
	SupervisorID uuid.UUID `gorm:"type:uuid;not null" json:"supervisor_id"`
	Supervisor   User      `gorm:"foreignKey:SupervisorID" json:"supervisor"`
	AssignedAt   time.Time `gorm:"not null;default:now()" json:"assigned_at"`
	AssignedBy   uuid.UUID `gorm:"type:uuid;not null" json:"assigned_by"`
}

func (ThesisSupervisor) TableName() string { return "thesis_supervisors" }
