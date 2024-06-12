package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-main-api/config"
	"github.com/mooncorn/gshub-main-api/internal"
)

// Returns user (if token is valid), a list of services, and plans.
func GetMetadata(c *gin.Context) {
	dbInstance := db.GetDatabase()
	email := c.GetString("userEmail")

	// Get user
	var user models.User
	if err := dbInstance.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		user = models.User{}
	}

	// Get services
	var services []models.Service
	if err := dbInstance.GetDB().Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch services"})
		return
	}

	// Get plans
	var plans []models.Plan
	if err := dbInstance.GetDB().Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch plans"})
		return
	}

	// Get servers
	var servers []models.Server
	if err := dbInstance.GetDB().Where("user_id = ?", user.ID).Find(&servers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch servers"})
		return
	}

	// Get instances
	ec2InstanceClient, err := internal.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create instance client"})
		return
	}

	var instanceIds []string
	for _, s := range servers {
		instanceIds = append(instanceIds, s.InstanceID)
	}

	instances, err := ec2InstanceClient.GetInstances(c, instanceIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get instances", "details": err.Error()})
		return
	}

	instancesMap := make(map[string]internal.Instance)
	for _, i := range instances {
		instancesMap[i.Id] = i
	}

	c.JSON(http.StatusOK, gin.H{
		"user":                   user,
		"services":               services,
		"plans":                  plans,
		"servers":                servers,
		"instances":              instancesMap,
		"latestServerAPIVersion": config.Env.LatestServerAPIVersion,
	})
}
