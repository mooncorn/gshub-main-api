package handlers

import (
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-main-api/aws"
)

func CreateInstance(c *gin.Context) {
	// Request structure for binding JSON input
	var request struct {
		PlanID    int `json:"PlanID"`
		ServiceID int `json:"ServiceID"`
	}

	// Bind JSON input to the request structure
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	email := c.GetString("userEmail")

	// Get User
	dbInstance := db.GetDatabase()
	var user models.User
	if err := dbInstance.GetDB().Where(&models.User{Email: email}).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user"})
		return
	}

	// Get Plan
	var plan models.Plan
	if err := dbInstance.GetDB().First(&plan, request.PlanID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid plan"})
		return
	}

	// Get Service
	var service models.Service
	if err := dbInstance.GetDB().First(&service, request.ServiceID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid service"})
		return
	}

	// Create new instance AWS EC2 instance
	ec2InstanceClient, err := aws.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create instance client"})
		return
	}

	ec2Instance, err := ec2InstanceClient.CreateInstance(c, types.InstanceType(plan.InstanceType), &aws.ContainerOptions{
		Image:  service.Image,
		Env:    strings.Split(service.Env, ","),
		Volume: service.Volume,
		Ports:  strings.Split(service.Ports, ","),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create instance", "details": err.Error()})
		return
	}

	server := models.Server{
		ServiceID:  service.ID,
		PlanID:     plan.ID,
		UserID:     user.ID,
		InstanceID: ec2Instance.Id,
	}

	if err := dbInstance.GetDB().Create(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		ec2InstanceClient.TerminateInstance(c, ec2Instance.Id)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"server": server})
}

func GetInstances(c *gin.Context) {
	email := c.GetString("userEmail")

	// Get User
	dbInstance := db.GetDatabase()
	var user models.User
	if err := dbInstance.GetDB().Where(&models.User{Email: email}).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user"})
		return
	}

	// Get servers that belong to the user
	var servers []models.Server
	if err := dbInstance.GetDB().Where(&models.Server{UserID: user.ID}).Find(&servers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
		return
	}

	instanceClient, err := aws.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance client"})
		return
	}

	var instanceIds []string
	for _, s := range servers {
		instanceIds = append(instanceIds, s.InstanceID)
	}

	instances, err := instanceClient.GetInstances(c, instanceIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get instances"})
		return
	}

	c.JSON(http.StatusOK, instances)
}
