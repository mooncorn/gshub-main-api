package instance_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/app"
	"github.com/mooncorn/gshub-main-api/utils"
)

// Start user's instance based on the provided instance ID.
func StartInstance(c *gin.Context, appCtx *app.Context) {
	instanceIDStr := c.Param("id")
	userEmail := c.GetString("userEmail")

	instanceID64, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid instance id", err, userEmail)
		return
	}

	// Check if the instance exists
	instance, err := appCtx.InstanceRepository.GetUserInstance(userEmail, uint(instanceID64))
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
