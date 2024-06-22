package context

import (
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/instance"
	"github.com/mooncorn/gshub-main-api/repositories"
	"gorm.io/gorm"
)

type AppContext struct {
	DB                 *gorm.DB
	InstanceClient     *instance.Client
	UserRepository     *repositories.UserRepository
	ServiceRepository  *repositories.ServiceRepository
	PlanRepository     *repositories.PlanRepository
	InstanceRepository *repositories.InstanceRepository
}

func NewAppContext(dbInstance *gorm.DB) *AppContext {
	return &AppContext{
		DB:                 dbInstance,
		InstanceClient:     instance.NewClient(),
		UserRepository:     repositories.NewUserRepository(dbInstance),
		ServiceRepository:  repositories.NewServiceRepository(dbInstance),
		PlanRepository:     repositories.NewPlanRepository(dbInstance),
		InstanceRepository: repositories.NewInstanceRepository(dbInstance),
	}
}

func (appCtx *AppContext) HandlerWrapper(handler func(*gin.Context, *AppContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c, appCtx)
	}
}
