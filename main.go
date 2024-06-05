package main

import (
	"log"
	"strings"

	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-main-api/config"
	"github.com/mooncorn/gshub-main-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	gormDB := db.NewGormDB(config.Env.DSN)
	db.SetDatabase(gormDB)

	// AutoMigrate the models
	err := db.GetDatabase().GetDB().AutoMigrate(
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

	r := gin.Default()

	// Middlewares
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		ExposeHeaders: []string{"Content-Length"},
		AllowHeaders:  []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}))

	r.Use(middlewares.CheckUser)

	// Public routes
	r.POST("/signin", handlers.SignIn)

	r.Use(middlewares.RequireUser)

	// Protected routes
	r.GET("/user", handlers.GetUser)
	r.GET("/metadata", handlers.GetMetadata)
	r.POST("/servers", handlers.CreateInstance)
	r.GET("/servers", handlers.GetInstances)

	r.GET("/services/:id", handlers.GetService)

	r.Run(":" + config.Env.Port)
}
