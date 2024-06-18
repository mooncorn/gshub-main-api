package main

import (
	"log"
	"strings"

	coreDB "github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-main-api/config"
	ctx "github.com/mooncorn/gshub-main-api/context"
	"github.com/mooncorn/gshub-main-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	gormDB := coreDB.NewGormDB(config.Env.DSN)
	coreDB.SetDatabase(gormDB)

	// AutoMigrate the models
	err := coreDB.GetDatabase().GetDB().AutoMigrate(
		&models.User{},
		&models.Plan{},
		&models.Service{},
		&models.ServiceEnv{},
		&models.ServiceEnvValue{},
		&models.ServiceVolume{},
		&models.ServicePort{},
		&models.Server{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if strings.ToLower(config.Env.GinMode) == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	appCtx := ctx.NewAppContext(coreDB.GetDatabase().GetDB())

	r := gin.Default()

	// Middlewares
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		ExposeHeaders: []string{"Content-Length"},
		AllowHeaders:  []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}))

	r.Use(middlewares.CheckUser)

	// Public routes
	r.POST("/signin", appCtx.HandlerWrapper(handlers.SignIn))
	r.GET("/metadata", appCtx.HandlerWrapper(handlers.GetMetadata))

	r.Use(middlewares.RequireUser)

	// Protected routes
	r.GET("/user", appCtx.HandlerWrapper(handlers.GetUser))

	r.POST("/servers", appCtx.HandlerWrapper(handlers.CreateInstance))
	r.DELETE("/servers/:id", appCtx.HandlerWrapper(handlers.TerminateInstance))
	r.POST("/servers/:id/start", appCtx.HandlerWrapper(handlers.StartInstance))
	r.POST("/servers/:id/stop", appCtx.HandlerWrapper(handlers.StopInstance))

	r.GET("/services/:id", appCtx.HandlerWrapper(handlers.GetService))

	// TODO: Require ADMIN role
	r.POST("/servers/update-server-apis", appCtx.HandlerWrapper(handlers.UpdateServerAPIs))

	r.Run(":" + config.Env.Port)
}
