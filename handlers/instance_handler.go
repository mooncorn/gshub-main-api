package handlers

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-core/utils"
	ctx "github.com/mooncorn/gshub-main-api/context"
	"github.com/mooncorn/gshub-main-api/instance"
)

// The payload for creating an instance
type CreateInstanceRequestBody struct {
	PlanID    uint `json:"planId" binding:"required"`
	ServiceID uint `json:"serviceId" binding:"required"`
}

// Start user's instance based on the provided instance ID.
func StartInstance(c *gin.Context, appCtx *ctx.AppContext) {
	instanceIDStr := c.Param("id")
	userEmail := c.GetString("userEmail")

	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, userEmail)
		return
	}

	// Check if the instance exists
	instance, err := appCtx.InstanceRepository.GetInstance(uint(instanceID64), userEmail)
	if err != nil {
		utils.HandleError(c, http.StatusNotFound, "Instance not found", err, userEmail)
		return
	}

	// Start the instance
	err = appCtx.InstanceClient.StartInstances(c, []string{instance.RealID})
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Unable to start instance", err, userEmail)
		return
	}

	c.Status(http.StatusOK)
}

// Stop user's instance based on the provided instance ID.
func StopInstance(c *gin.Context, appCtx *ctx.AppContext) {
	instanceIDStr := c.Param("id")
	userEmail := c.GetString("userEmail")

	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, userEmail)
		return
	}

	// Check if the instance exists
	server, err := appCtx.InstanceRepository.GetInstance(uint(instanceID64), userEmail)
	if err != nil {
		utils.HandleError(c, http.StatusNotFound, "Instance not found", err, userEmail)
		return
	}

	// Stop the instance
	err = appCtx.InstanceClient.StopInstances(c, []string{server.RealID})
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Unable to stop instance", err, userEmail)
		return
	}

	c.Status(http.StatusOK)
}

// CreateInstance creates a new instance and associates it with the user, plan, and service.
func CreateInstance(c *gin.Context, appCtx *ctx.AppContext) {
	var request CreateInstanceRequestBody

	// Bind JSON input to the request structure
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorMessage{Error: "Invalid request"})
		return
	}

	userEmail := c.GetString("userEmail")

	// Get User
	user, err := appCtx.UserRepository.GetUserByEmail(userEmail)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid user", err, userEmail)
		return
	}

	// Get Plan
	plan, err := appCtx.PlanRepository.GetPlan(request.PlanID)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid plan", err, userEmail)
		return
	}

	// Get Service
	service, err := appCtx.ServiceRepository.GetService(request.ServiceID)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid service", err, userEmail)
		return
	}

	instanceType, err := instance.ParseInstanceType(plan.InstanceType)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance type", err, userEmail)
		return
	}

	ec2Instance, err := appCtx.InstanceClient.CreateInstance(c, &instanceType)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Failed to create instance", err, userEmail)
		return
	}

	instance := models.Instance{
		ServiceID: service.ID,
		PlanID:    plan.ID,
		UserID:    user.ID,
		RealID:    ec2Instance.Id,
		Ready:     false,
	}

	if err := appCtx.InstanceRepository.CreateInstance(&instance); err != nil {
		appCtx.InstanceClient.TerminateInstances(c, []string{ec2Instance.Id})
		utils.HandleError(c, http.StatusInternalServerError, "Failed to create instance", err, userEmail)
		return
	}

	c.JSON(http.StatusCreated, instance)
}

// TerminateInstance terminates an EC2 instance and deletes its record from the instance repository.
//
// If the instance exists, it proceeds to terminate the instance using the InstanceClient.
// After successfully terminating the instance, it deletes the corresponding instance record
// from the instance repository. If any step fails, an appropriate error response is returned.
func TerminateInstance(c *gin.Context, appCtx *ctx.AppContext) {
	instanceIDStr := c.Param("id")
	userEmail := c.GetString("userEmail")

	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, userEmail)
		return
	}

	// Check if the instance exists
	instance, err := appCtx.InstanceRepository.GetInstance(uint(instanceID64), userEmail)
	if err != nil {
		utils.HandleError(c, http.StatusNotFound, "Instance not found", err, userEmail)
		return
	}

	// Terminate the instance
	if err := appCtx.InstanceClient.TerminateInstances(c, []string{instance.RealID}); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Unable to terminate instance", err, userEmail)
		return
	}

	// Delete the instance record
	if err := appCtx.InstanceRepository.DeleteInstance(instance.ID, userEmail); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Unable to delete instance", err, userEmail)
		return
	}

	c.Status(http.StatusOK)
}

// Updates the APIs on all running instances by executing an update script.
func UpdateServerAPIs(c *gin.Context, appCtx *ctx.AppContext) {
	userEmail := c.GetString("userEmail")

	instanceIds, err := appCtx.InstanceClient.GetRunningInstances(c)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to get running instances", err, userEmail)
		return
	}

	if len(*instanceIds) == 0 {
		c.JSON(http.StatusOK, utils.SuccessMessage{Message: "No running instances found"})
		return
	}

	data, err := os.ReadFile("./scripts/instance-update.sh")
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Unable to read instance-update script", err, userEmail)
		return
	}

	command := string(data)

	err = appCtx.InstanceClient.SendCommand(c, &command, instanceIds)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to send command to instances", err, userEmail)
		return
	}

	c.JSON(http.StatusOK, utils.SuccessMessage{Message: "API update initiated on all running instances"})
}
