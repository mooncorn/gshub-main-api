package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/middlewares"
	"github.com/mooncorn/gshub-core/utils"
	"github.com/mooncorn/gshub-main-api/config"
	ctx "github.com/mooncorn/gshub-main-api/context"
	"github.com/mooncorn/gshub-main-api/handlers"
	"github.com/mooncorn/gshub-main-api/models"
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
	r.Use(middlewares.RequireRole("admin"))
	r.POST("/servers/update-server-apis", appCtx.HandlerWrapper(handlers.UpdateServerAPIs))

	go r.Run(":" + config.Env.Port)

	// API for instances
	r2 := gin.Default()
	r2.GET("/startup/:id", appCtx.HandlerWrapper(handleStartupEvent))
	r2.POST("/shutdown/:id", appCtx.HandlerWrapper(handleShutdownEvent))
	r2.Run(":8081")
}

func handleStartupEvent(c *gin.Context, appCtx *ctx.AppContext) {
	instanceIDStr := c.Param("id")
	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, instanceIDStr)
		return
	}

	// get instance
	instance, err := appCtx.InstanceRepository.GetInstance(uint(instanceID64))
	if err != nil {
		utils.HandleError(c, http.StatusNotFound, "Instance not found", err, instanceIDStr)
		return
	}

	// get plan
	plan, err := appCtx.PlanRepository.GetPlan(instance.PlanID)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to get plan", err, instanceIDStr)
		return
	}

	// get service
	service, err := appCtx.ServiceRepository.GetService(instance.ServiceID)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to get service", err, instanceIDStr)
		return
	}
	if appCtx.InstanceRepository.UpdateInstance(uint(instanceID64), true, "") != nil {
		utils.HandleError(c, http.StatusBadRequest, "Failed to update instance", err, instanceIDStr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instanceMemory":               plan.Memory,
		"serviceNameId":                service.NameID,
		"serviceImage":                 service.Image,
		"serviceMinimumMemoryRequired": service.MinMem,
		"ownerId":                      instance.UserID,
		"cycles":                       instance.Cycles,
	})
}

type EventShutdownData struct {
	BurnedCycles uint `json:"burnedCycles"`
}

func handleShutdownEvent(c *gin.Context, appCtx *ctx.AppContext) {
	instanceIDStr := c.Param("id")

	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, instanceIDStr)
		return
	}

	var request EventShutdownData

	// Bind JSON input to the request structure
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorMessage{Error: "Invalid request"})
		return
	}

	if appCtx.InstanceRepository.UpdateInstance(uint(instanceID64), false, "") != nil {
		utils.HandleError(c, http.StatusBadRequest, "Failed to update instance", err, instanceIDStr)
		return
	}

	fmt.Printf("Create burned cycles: %d\n", request.BurnedCycles)

	c.Status(http.StatusOK)
}
