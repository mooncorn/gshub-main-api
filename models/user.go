package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleDefault UserRole = "user"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"` // Ensures email is unique and not null
	Role      UserRole       `gorm:"not null" json:"role"`              // Ensures role is not null
}
