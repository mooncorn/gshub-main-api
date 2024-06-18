package context

import (
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/db"
	"github.com/mooncorn/gshub-main-api/instance"
	"gorm.io/gorm"
)

type AppContext struct {
	InstanceClient    *instance.Client
	UserRepository    *db.UserRepository
	ServiceRepository *db.ServiceRepository
	PlanRepository    *db.PlanRepository
	ServerRepository  *db.ServerRepository
}

func NewAppContext(dbInstance *gorm.DB) *AppContext {
	return &AppContext{
		InstanceClient:    instance.NewClient(),
		UserRepository:    db.NewUserRepository(dbInstance),
		ServiceRepository: db.NewServiceRepository(dbInstance),
		PlanRepository:    db.NewPlanRepository(dbInstance),
		ServerRepository:  db.NewServerRepository(dbInstance),
	}
}

func (appCtx *AppContext) HandlerWrapper(handler func(*gin.Context, *AppContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c, appCtx)
	}
}
