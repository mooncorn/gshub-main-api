package app

import (
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/instance/instance_aws"
	"github.com/mooncorn/gshub-main-api/instance/instance_repositories"
	"github.com/mooncorn/gshub-main-api/plan/plan_repositories"
	"github.com/mooncorn/gshub-main-api/service/service_repositories"
	"github.com/mooncorn/gshub-main-api/user/user_repositories"

	"gorm.io/gorm"
)

type Context struct {
	DB                             *gorm.DB
	InstanceClient                 *instance_aws.AWSClient
	UserRepository                 *user_repositories.UserRepository
	ServiceRepository              *service_repositories.ServiceRepository
	PlanRepository                 *plan_repositories.PlanRepository
	InstanceRepository             *instance_repositories.InstanceRepository
	InstanceCyclesRepository       *instance_repositories.InstanceCyclesRepository
	InstanceBurnedCyclesRepository *instance_repositories.InstanceBurnedCyclesRepository
}

func NewContext(dbInstance *gorm.DB) *Context {
	return &Context{
		DB:                             dbInstance,
		InstanceClient:                 instance_aws.NewAWSClient(),
		UserRepository:                 user_repositories.NewUserRepository(dbInstance),
		ServiceRepository:              service_repositories.NewServiceRepository(dbInstance),
		PlanRepository:                 plan_repositories.NewPlanRepository(dbInstance),
		InstanceRepository:             instance_repositories.NewInstanceRepository(dbInstance),
		InstanceCyclesRepository:       instance_repositories.NewInstanceCyclesRepository(dbInstance),
		InstanceBurnedCyclesRepository: instance_repositories.NewInstanceBurnedCyclesRepository(dbInstance),
	}
}

func (appCtx *Context) HandlerWrapper(handler func(*gin.Context, *Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c, appCtx)
	}
}
