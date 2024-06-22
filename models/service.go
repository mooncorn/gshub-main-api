package models

import (
	"time"

	"gorm.io/gorm"
)

// Service represents a service that can be part of a plan
type Service struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	NameID    string         `gorm:"uniqueIndex" json:"nameId"` // Unique identifier
	MinMem    int            `gorm:"not null" json:"minMem"`    // Minimum memory required in MB
	RecMem    int            `gorm:"not null" json:"recMem"`    // Recommended memory in MB
	Name      string         `gorm:"not null" json:"name"`      // Short name of the service
	NameLong  string         `gorm:"not null" json:"nameLong"`  // Full name of the service
	Image     string         `gorm:"not null" json:"-"`         // Image of the docker container
}
