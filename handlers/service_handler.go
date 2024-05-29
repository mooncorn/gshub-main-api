package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mooncorn/gshub-core/db"
	"github.com/mooncorn/gshub-core/models"
)

func GetService(c *gin.Context) {
	serviceIdStr := c.Param("id")
	dbInstance := db.GetDatabase()

	// Check if serviceId is a valid integer
	serviceId, err := strconv.Atoi(serviceIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service id"})
		return
	}

	var service models.Service

	if err := dbInstance.GetDB().Model(&models.Service{}).Preload("Env.Values").Preload("Ports").Preload("Volumes").Where(serviceId).First(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service does not exist", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"service": service})
}
