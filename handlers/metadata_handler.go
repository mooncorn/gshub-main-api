package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-core/utils"
	ctx "github.com/mooncorn/gshub-main-api/context"
)

func GetMetadata(c *gin.Context, appCtx *ctx.AppContext) {
	userEmail := c.GetString("userEmail")

	var user *models.User
	var instances *[]models.Instance
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
