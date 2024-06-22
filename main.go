package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-main-api/config"
	ctx "github.com/mooncorn/gshub-main-api/context"
	"github.com/mooncorn/gshub-main-api/handlers"
	"gorm.io/gorm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	// Initialize database
	gormDB := db.NewPostgresDB(config.Env.DSN, &gorm.Config{})

	// Auto migrate the models
	err := gormDB.GetDB().AutoMigrate(
		&models.User{},
		&models.Plan{},
		&models.Service{},
		&models.Instance{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if strings.ToLower(config.Env.AppEnv) == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	appCtx := ctx.NewAppContext(gormDB.GetDB())

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

	// Admin protected routes
	r.Use(middlewares.RequireRole(models.UserRoleAdmin))
	r.POST("/servers/update-server-apis", appCtx.HandlerWrapper(handlers.UpdateServerAPIs))

	go r.Run(":" + config.Env.Port)

	// API for instances
	r2 := gin.Default()
	r2.GET("/startup", appCtx.HandlerWrapper(handleStartupEvent))
	r2.POST("/shutdown", appCtx.HandlerWrapper(handleShutdownEvent))
	r2.Run(":8081")
}

func handleStartupEvent(c *gin.Context, appCtx *ctx.AppContext) {
	fmt.Println("startup")
}

func handleShutdownEvent(c *gin.Context, appCtx *ctx.AppContext) {
	fmt.Println("shutdown")
}
