package instance_models

import (
	"time"

	"gorm.io/gorm"
)

type InstanceBurnedCycle struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	InstanceID uint           `gorm:"not null" json:"instanceId"`
	Amount     uint           `json:"amount"`
}
