package models

import (
	"time"

	"gorm.io/gorm"
)

// Server represents a server instance in the system
type Instance struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	RealID    string         `gorm:"not null" json:"realId"`
	Name      string         `json:"name"`

	Ready    bool   `json:"ready"`
	PublicIP string `json:"publicIp"`

	ServiceID uint `gorm:"not null" json:"serviceId"` // Reference to the service
	PlanID    uint `gorm:"not null" json:"planId"`    // Reference to the plan
	UserID    uint `gorm:"not null" json:"userId"`    // Reference to the user

	Cycles       []InstanceCycles       `json:"cycles"`
	BurnedCycles []InstanceBurnedCycles `json:"burnedCycles"`
}
