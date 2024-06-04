package main

import (
	"log"
	"os"

	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-main-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	dsn := os.Getenv("DSN")

	gormDB := db.NewGormDB(dsn)
	db.SetDatabase(gormDB)

	// AutoMigrate the models
	err = db.GetDatabase().GetDB().AutoMigrate(&models.Plan{}, &models.Service{}, &models.Server{}, &models.User{}, &models.ServiceEnv{}, &models.ServiceEnvValue{}, &models.ServicePort{}, &models.ServiceVolume{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
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

	r.Run(":" + port)
}
