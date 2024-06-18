package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/models"
	"github.com/mooncorn/gshub-core/utils"
	ctx "github.com/mooncorn/gshub-main-api/context"
	"github.com/mooncorn/gshub-main-api/dto"
	"github.com/mooncorn/gshub-main-api/instance"
)

func GetMetadata(c *gin.Context, appCtx *ctx.AppContext) {
	userEmail := c.GetString("userEmail")

	var user *models.User
	var servers *[]models.Server
	var err error

	if userEmail != "" {
		user, err = appCtx.UserRepository.GetUserByEmail(userEmail)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Could not get user", err, userEmail)
			return
		}

		servers, err = appCtx.ServerRepository.GetUserServers(user.ID)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Could not get servers", err, userEmail)
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

	var serverInstances []dto.ServerInstance
	if servers != nil && len(*servers) > 0 {
		var instanceIds []string
		for _, s := range *servers {
			instanceIds = append(instanceIds, s.InstanceID)
		}

		instances, err := appCtx.InstanceClient.GetInstances(c, &instanceIds)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Failed to get instances", err, userEmail)
			return
		}

		instancesMap := make(map[string]instance.Instance)
		for _, i := range *instances {
			instancesMap[i.Id] = i
		}

		for _, s := range *servers {
			i, ok := instancesMap[s.InstanceID]
			if !ok {
				utils.HandleError(c, http.StatusInternalServerError, "Failed to get instance", fmt.Errorf("instance should exist: %s", s.InstanceID), userEmail)
				return
			}

			serverInstances = append(serverInstances, dto.NewServerInstance(&s, &i))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":     user,
		"servers":  serverInstances,
		"services": services,
		"plans":    plans,
	})
}
