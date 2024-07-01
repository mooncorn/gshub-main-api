package utils

import (
	"log"

	"github.com/gin-gonic/gin"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

type SuccessMessage struct {
	Message string `json:"message"`
}

func HandleError(c *gin.Context, status int, message string, err error, userEmail string) {
	log.Printf("%s: Error: %s - Details: %v", userEmail, message, err)
	c.JSON(status, ErrorMessage{Error: message})
}

func HandleSuccess(c *gin.Context, status int, message string, userEmail string) {
	log.Printf("%s: Success: %s", userEmail, message)
	c.JSON(status, SuccessMessage{Message: message})
}
