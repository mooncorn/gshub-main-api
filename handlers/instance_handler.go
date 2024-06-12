package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-main-api/internal"
)

func RebootInstance(c *gin.Context) {
	instanceId := c.Param("id")

	instanceClient, err := internal.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance client"})
		return
	}

	err = instanceClient.RebootInstances(c, []string{instanceId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to reboot instance", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Reboot initiated", "instanceId": instanceId})
}

func StartInstance(c *gin.Context) {
	instanceId := c.Param("id")

	instanceClient, err := internal.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance client"})
		return
	}

	err = instanceClient.StartInstances(c, []string{instanceId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to start instance", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Start initiated", "instanceId": instanceId})
}

func StopInstance(c *gin.Context) {
	instanceId := c.Param("id")

	instanceClient, err := internal.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance client"})
		return
	}

	err = instanceClient.StopInstances(c, []string{instanceId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to stop instance", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Stop initiated", "instanceId": instanceId})
}

func CreateInstance(c *gin.Context) {
	// Request structure for binding JSON input
	var request struct {
		PlanID    int `json:"planId"`
		ServiceID int `json:"serviceId"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user"})
		return
	}

	// Get Plan
	var plan models.Plan
	if err := dbInstance.GetDB().First(&plan, request.PlanID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan"})
		return
	}

	// Get Service
	var service models.Service
	if err := dbInstance.GetDB().First(&service, request.ServiceID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service"})
		return
	}

	// Create new instance AWS EC2 instance
	ec2InstanceClient, err := internal.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create instance client"})
		return
	}

	ec2Instance, err := ec2InstanceClient.CreateInstance(c, types.InstanceType(plan.InstanceType))
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

func TerminateInstance(c *gin.Context) {
	instanceId := c.Param("id")

	instanceClient, err := internal.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance client"})
		return
	}

	err = instanceClient.TerminateInstances(c, []string{instanceId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to terminate instance", "details": err.Error()})
		return
	}

	dbInstance := db.GetDatabase()

	if err = dbInstance.GetDB().Where(&models.Server{InstanceID: instanceId}).Delete(&models.Server{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete instance", "details": err.Error()})
		return
	}

	c.Status(http.StatusOK)
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

	instanceClient, err := internal.NewInstanceClient(c)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get instances", "details": err.Error()})
		return
	}

	instancesMap := make(map[string]internal.Instance)
	for _, i := range instances {
		instancesMap[i.Id] = i
	}

	c.JSON(http.StatusOK, gin.H{"count": len(servers), "instances": instancesMap, "servers": servers})
}

func UpdateServerAPIs(c *gin.Context) {
	instanceClient, err := internal.NewInstanceClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance client"})
		return
	}

	instanceIds, err := instanceClient.GetRunningInstances(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get running instances", "details": err.Error()})
		return
	}

	if len(instanceIds) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "No running instances found"})
		return
	}

	cfg, err := config.LoadDefaultConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load SDK config", "details": err.Error()})
		return
	}

	ssmClient := ssm.NewFromConfig(cfg)

	data, err := os.ReadFile("./scripts/instance-update.sh")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read instance-update script", "details": err.Error()})
		return
	}

	commandInput := &ssm.SendCommandInput{
		InstanceIds:  instanceIds,
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]string{
			"commands": {string(data)},
		},
	}

	commandOutput, err := ssmClient.SendCommand(c, commandInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send command to instances", "details": err.Error()})
		return
	}

	commandId := *commandOutput.Command.CommandId

	// Wait for the command to execute using ssm.NewCommandExecutedWaiter
	waiter := ssm.NewCommandExecutedWaiter(ssmClient)
	waiterInput := &ssm.GetCommandInvocationInput{
		CommandId: aws.String(commandId),
	}

	// Wait for all instances
	for _, instanceId := range instanceIds {
		waiterInput.InstanceId = aws.String(instanceId)
		err = waiter.Wait(c, waiterInput, 10*time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to wait for command execution on instance %s", instanceId), "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "API update initiated on all running instances"})
}
