package instance_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/app"
	"github.com/mooncorn/gshub-main-api/instance/instance_aws"
	"github.com/mooncorn/gshub-main-api/instance/instance_models"
	"github.com/mooncorn/gshub-main-api/utils"
)

// The payload for creating an instance
type CreateInstanceRequestBody struct {
	PlanID    uint `json:"planId" binding:"required"`
	ServiceID uint `json:"serviceId" binding:"required"`
}

// CreateInstance creates a new instance and associates it with the user, plan, and service.
func CreateInstance(c *gin.Context, appCtx *app.Context) {
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

	instanceType, err := instance_aws.ParseInstanceType(plan.InstanceType)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance type", err, userEmail)
		return
	}

	ec2Instance, err := appCtx.InstanceClient.CreateInstance(c, &instanceType)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Failed to create instance", err, userEmail)
		return
	}

	instance := instance_models.Instance{
		PlanID:   plan.ID,
		UserID:   user.ID,
		RealID:   ec2Instance.Id,
		Ready:    false,
		Name:     "",
		PublicIP: "",
	}

	if err := appCtx.InstanceRepository.CreateInstance(&instance); err != nil {
		appCtx.InstanceClient.TerminateInstances(c, []string{ec2Instance.Id})
		utils.HandleError(c, http.StatusInternalServerError, "Failed to create instance", err, userEmail)
		return
	}

	c.JSON(http.StatusCreated, instance)
}
