package service_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/utils"
	"github.com/mooncorn/gshub-main-api/app"
	service_presets "github.com/mooncorn/gshub-main-api/service/presets"
)

func GetService(c *gin.Context, appCtx *app.Context) {
	serviceIdStr := c.Param("id")

	serviceID64, err := strconv.ParseUint(serviceIdStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid service id", err, "")
		return
	}

	service, err := appCtx.ServiceRepository.GetService(uint(serviceID64))
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Cannot get service", err, "null")
		return
	}

	config, err := service_presets.GetServiceConfiguration(service.NameID)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Cannot get service config", err, "null")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": service,
		"config":  config,
	})
}
