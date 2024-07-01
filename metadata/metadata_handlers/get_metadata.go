package metadata_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/utils"
	"github.com/mooncorn/gshub-main-api/app"
	"github.com/mooncorn/gshub-main-api/instance/instance_models"
	"github.com/mooncorn/gshub-main-api/user/user_models"
)

func GetMetadata(c *gin.Context, appCtx *app.Context) {
	userEmail := c.GetString("userEmail")

	var user *user_models.User
	var instances *[]instance_models.Instance
	var err error

	if userEmail != "" {
		user, err = appCtx.UserRepository.GetUserByEmail(userEmail)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Could not get user", err, userEmail)
			return
		}

		instances, err = appCtx.InstanceRepository.GetUserInstances(user.ID)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Could not get instances", err, userEmail)
			return
		}
	}

	services, err := appCtx.ServiceRepository.GetServices()
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Could not get services", err, userEmail)
		return
	}

	plans, err := appCtx.PlanRepository.GetPlans()
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Could not get plans", err, userEmail)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":      user,
		"instances": instances,
		"services":  services,
		"plans":     plans,
	})
}
