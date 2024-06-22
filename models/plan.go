package models

import (
	"time"

	"gorm.io/gorm"
)

// Plan represents an instance type in the system
type Plan struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	InstanceType string         `gorm:"not null" json:"instanceType"` // Type of instance
	Name         string         `gorm:"not null" json:"name"`         // Name of the plan
	VCores       int            `gorm:"not null" json:"vCores"`       // Number of virtual cores
	Memory       int            `gorm:"not null" json:"memory"`       // Amount of memory in MB
	Price        float64        `gorm:"not null" json:"price"`        // Price of the plan per hour
	Disk         int            `gorm:"not null" json:"disk"`         // Disk space in GB
	Enabled      bool           `gorm:"not null" json:"enabled"`      // Indicates if the plan is enabled
}
