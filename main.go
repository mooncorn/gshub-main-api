package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-main-api/app"
	"gorm.io/gorm"

	"github.com/mooncorn/gshub-main-api/instance/instance_models"
	"github.com/mooncorn/gshub-main-api/plan/plan_models"
	"github.com/mooncorn/gshub-main-api/service/service_models"
	"github.com/mooncorn/gshub-main-api/user/user_models"

	"github.com/mooncorn/gshub-main-api/instance/instance_handlers"
	"github.com/mooncorn/gshub-main-api/metadata/metadata_handlers"
	"github.com/mooncorn/gshub-main-api/service/service_handlers"
	"github.com/mooncorn/gshub-main-api/user/user_handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	loadEnv()

	// Initialize database and migrate models
	gormDB := initializeDatabase()

	// Set Gin mode based on environment
	setGinMode()

	// Create application context
	appCtx := app.NewContext(gormDB)

	// Setup and start the main server
	mainRouter := setupMainRouter(appCtx)
	go startServer(mainRouter, ":8080")

	// Setup and start the instance server
	instanceRouter := setupInstanceRouter(appCtx)
	startServer(instanceRouter, ":8081")
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func initializeDatabase() *gorm.DB {
	gormDB := db.NewPostgresDB(os.Getenv("DSN"), &gorm.Config{})
	if err := gormDB.GetDB().AutoMigrate(
		&user_models.User{},
		&plan_models.Plan{},
		&service_models.Service{},
		&instance_models.Instance{},
		&instance_models.InstanceCycle{},
		&instance_models.InstanceBurnedCycle{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	return gormDB.DB
}

func setGinMode() {
	if strings.ToLower(os.Getenv("APP_ENV")) == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func setupMainRouter(appCtx *app.Context) *gin.Engine {
	r := gin.Default()

	// Middlewares
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		ExposeHeaders: []string{"Content-Length"},
		AllowHeaders:  []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}))
	r.Use(middlewares.CheckUser)

	// Public routes
	r.POST("/signin", appCtx.HandlerWrapper(user_handlers.SignIn))
	r.GET("/metadata", appCtx.HandlerWrapper(metadata_handlers.GetMetadata))

	r.Use(middlewares.RequireUser)

	// Protected routes
	r.GET("/user", appCtx.HandlerWrapper(user_handlers.GetUser))
	r.POST("/instance", appCtx.HandlerWrapper(instance_handlers.CreateInstance))
	r.DELETE("/instance/:id", appCtx.HandlerWrapper(instance_handlers.TerminateInstance))
	r.POST("/instance/:id/start", appCtx.HandlerWrapper(instance_handlers.StartInstance))
	r.POST("/instance/:id/stop", appCtx.HandlerWrapper(instance_handlers.StopInstance))
	r.GET("/services/:id", appCtx.HandlerWrapper(service_handlers.GetService))

	// Admin protected routes
	r.Use(middlewares.RequireRole("admin"))
	r.POST("/instance/rollout-update", appCtx.HandlerWrapper(instance_handlers.RolloutInstanceUpdate))

	return r
}

func setupInstanceRouter(appCtx *app.Context) *gin.Engine {
	r := gin.Default()
	r.GET("/startup/:id", appCtx.HandlerWrapper(instance_handlers.OnInstanceStartup))
	r.POST("/shutdown/:id", appCtx.HandlerWrapper(instance_handlers.OnInstanceShutdown))
	return r
}

func startServer(router *gin.Engine, address string) {
	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server on %s: %v", address, err)
	}
}

// func handleStartupEvent(c *gin.Context, appCtx *app.Context) {
// 	instanceIDStr := c.Param("id")
// 	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
// 	if err != nil {
// 		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, instanceIDStr)
// 		return
// 	}

// 	instance, err := appCtx.InstanceRepository.GetInstance(uint(instanceID64))
// 	if err != nil {
// 		utils.HandleError(c, http.StatusNotFound, "Instance not found", err, instanceIDStr)
// 		return
// 	}

// 	plan, err := appCtx.PlanRepository.GetPlan(instance.PlanID)
// 	if err != nil {
// 		utils.HandleError(c, http.StatusInternalServerError, "Failed to get plan", err, instanceIDStr)
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"instanceMemory": plan.Memory,
// 		"ownerId":        instance.UserID,
// 		"cycles":         65,
// 	})
// }

// type EventShutdownData struct {
// 	BurnedCycles float64 `json:"burnedCycles"`
// }

// func handleShutdownEvent(c *gin.Context, appCtx *app.Context) {
// 	instanceIDStr := c.Param("id")
// 	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
// 	if err != nil {
// 		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, instanceIDStr)
// 		return
// 	}

// 	var request EventShutdownData
// 	if err := c.BindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, utils.ErrorMessage{Error: "Invalid request"})
// 		return
// 	}

// 	if err := appCtx.InstanceRepository.UpdateInstance(uint(instanceID64), false, ""); err != nil {
// 		utils.HandleError(c, http.StatusBadRequest, "Failed to update instance", err, instanceIDStr)
// 		return
// 	}

// 	fmt.Printf("%f cycles burned\n", request.BurnedCycles)
// 	c.Status(http.StatusOK)
// }
