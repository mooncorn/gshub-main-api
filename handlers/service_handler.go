package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/utils"
	ctx "github.com/mooncorn/gshub-main-api/context"
)

func GetService(c *gin.Context, appCtx *ctx.AppContext) {
	serviceIdStr := c.Param("id")

	// Check if serviceId is a valid integer
	serviceId, err := strconv.Atoi(serviceIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service id"})
		return
	}

	service, err := appCtx.ServiceRepository.GetServicePreloaded(serviceId)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Cannot get service", err, "null")
		return
	}

	c.JSON(http.StatusOK, service)
}
