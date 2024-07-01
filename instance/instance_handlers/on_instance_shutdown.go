package instance_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/app"
	"github.com/mooncorn/gshub-main-api/instance/instance_models"
	"github.com/mooncorn/gshub-main-api/utils"
)

type ShutdownPayload struct {
	BurnedCycleAmount uint `json:"burnedCycleAmount"`
}

func OnInstanceShutdown(c *gin.Context, appCtx *app.Context) {
	instanceIDStr := c.Param("id")
	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, instanceIDStr)
		return
	}

	var request ShutdownPayload
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
	instance.Ready = false
	instance.PublicIP = ""
	if err := appCtx.InstanceRepository.SaveInstance(instance); err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to save instance", err, instanceIDStr)
		return
	}

	// create burned cycle
	burnedCycle := instance_models.InstanceBurnedCycle{
		InstanceID: instance.ID,
		Amount:     request.BurnedCycleAmount,
	}
	if err = appCtx.InstanceBurnedCyclesRepository.CreateBurnedInstanceBurnedCycles(&burnedCycle); err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to create burned cycle", err, instanceIDStr)
		return
	}

	// return burned cycles
	c.JSON(http.StatusOK, gin.H{
		"burnedCycle": burnedCycle,
	})
}
