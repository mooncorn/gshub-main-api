package instance_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/app"
	"github.com/mooncorn/gshub-main-api/instance/instance_models"
	service_presets "github.com/mooncorn/gshub-main-api/service/presets"
	"github.com/mooncorn/gshub-main-api/utils"
)

type StartupPayload struct {
	FailedBurnedCycleAmount uint   `json:"failedBurnedCycleAmount"`
	PublicIP                string `json:"publicIp"`
}

func OnInstanceStartup(c *gin.Context, appCtx *app.Context) {
	instanceIDStr := c.Param("id")
	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, instanceIDStr)
		return
	}

	var request StartupPayload
	if err := c.BindJSON(&request); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request", err, instanceIDStr)
		return
	}

	// check if instance exists
	instance, err := appCtx.InstanceRepository.GetInstance(uint(instanceID64))
	if err != nil {
		utils.HandleError(c, http.StatusNotFound, "Instance not found", err, instanceIDStr)
		return
	}

	// update instance
	instance.Ready = true
	instance.PublicIP = request.PublicIP
	if err := appCtx.InstanceRepository.SaveInstance(instance); err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to save instance", err, instanceIDStr)
		return
	}

	// get plan
	plan, err := appCtx.PlanRepository.GetPlan(instance.PlanID)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Plan not found", err, instanceIDStr)
		return
	}

	// get available cycles for this instance
	cyclesAmount, err := appCtx.InstanceCyclesRepository.GetInstanceCyclesSum(instance.ID)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to get cycles amount", err, instanceIDStr)
		return
	}

	// get services
	services, err := appCtx.ServiceRepository.GetServices()
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Cannot get services", err, "null")
		return
	}

	// get service configs
	configs, err := service_presets.GetServiceConfigurations()
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Cannot get service config", err, "null")
		return
	}

	if request.FailedBurnedCycleAmount > 0 {
		// create burned cycle
		burnedCycle := &instance_models.InstanceBurnedCycle{
			InstanceID: instance.ID,
			Amount:     request.FailedBurnedCycleAmount,
		}
		if err = appCtx.InstanceBurnedCyclesRepository.CreateBurnedInstanceBurnedCycles(burnedCycle); err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Failed to create burned cycle", err, instanceIDStr)
			return
		}
	}

	// return instance data
	c.JSON(http.StatusOK, gin.H{
		"instanceMemory": plan.Memory,
		"ownerId":        instance.UserID,
		"cycles":         cyclesAmount,
		"services":       services,
		"serviceConfigs": configs,
	})
}
