package entity

import "time"

type Role struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (Role) TableName() string { return "roles" }
