package service_models

import (
	"time"

	"gorm.io/gorm"
)

// Service represents a supported service that can be hosted on an instance
type Service struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	// Service NameID is used to identify service configuration preset
	NameID string `gorm:"uniqueIndex" json:"nameId"`
}
