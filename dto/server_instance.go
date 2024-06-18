package dto

import (
	"time"

	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-main-api/instance"
)

type ServerInstance struct {
	ID        uint               `json:"id"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty"`
	ServiceID uint               `json:"serviceId"`
	PlanID    uint               `json:"planId"`
	UserID    uint               `json:"userId"`
	Instance  *instance.Instance `json:"instance"`
}

func NewServerInstance(server *models.Server, instance *instance.Instance) ServerInstance {
	return ServerInstance{
		ID:        server.ID,
		CreatedAt: server.CreatedAt,
		UpdatedAt: server.UpdatedAt,
		DeletedAt: &server.DeletedAt.Time,
		ServiceID: server.ServiceID,
		PlanID:    server.PlanID,
		UserID:    server.UserID,
		Instance:  instance,
	}
}
