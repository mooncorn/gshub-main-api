package instance_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/app"
	"github.com/mooncorn/gshub-main-api/utils"
)

// Stop user's instance based on the provided instance ID.
func StopInstance(c *gin.Context, appCtx *app.Context) {
	instanceIDStr := c.Param("id")
	userEmail := c.GetString("userEmail")

	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, userEmail)
		return
	}

	// Check if the instance exists
	server, err := appCtx.InstanceRepository.GetUserInstance(userEmail, uint(instanceID64))
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
