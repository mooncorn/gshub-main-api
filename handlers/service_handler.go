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

	serviceID64, err := strconv.ParseUint(serviceIdStr, 10, 32)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid service id", err, "")
		return
	}

	service, err := appCtx.ServiceRepository.GetServicePreloaded(uint(serviceID64))
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Cannot get service", err, "null")
		return
	}

	c.JSON(http.StatusOK, service)
}
