package instance_handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/utils"
	"github.com/mooncorn/gshub-main-api/app"
)

// Updates the APIs on all running instances by executing an update script.
func RolloutInstanceUpdate(c *gin.Context, appCtx *app.Context) {
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
