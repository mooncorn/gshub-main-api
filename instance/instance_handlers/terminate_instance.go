package instance_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-main-api/app"
	"github.com/mooncorn/gshub-main-api/utils"
)

// TerminateInstance terminates an EC2 instance and deletes its record from the instance repository.
//
// If the instance exists, it proceeds to terminate the instance using the InstanceClient.
// After successfully terminating the instance, it deletes the corresponding instance record
// from the instance repository. If any step fails, an appropriate error response is returned.
func TerminateInstance(c *gin.Context, appCtx *app.Context) {
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

	// Terminate the instance
	if err := appCtx.InstanceClient.TerminateInstances(c, []string{instance.RealID}); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Unable to terminate instance", err, userEmail)
		return
	}

	// Delete the instance record
	if err := appCtx.InstanceRepository.DeleteUserInstance(userEmail, instance.ID); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Unable to delete instance", err, userEmail)
		return
	}

	c.Status(http.StatusOK)
}
