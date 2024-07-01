package user_handlers

import (
	"net/http"

	"github.com/mooncorn/gshub-main-api/app"
	"github.com/mooncorn/gshub-main-api/user/user_models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUser retrieves the user information from the database based on the email stored in the context.
func GetUser(c *gin.Context, appCtx *app.Context) {
	// Retrieve the user email from the context
	userEmail, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User email not found in context"})
		return
	}

	// Cast the user email to string
	email, ok := userEmail.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user email format"})
		return
	}

	// Fetch the user information from the database
	var user user_models.User
	if err := appCtx.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Return the user information as JSON
	c.JSON(http.StatusOK, gin.H{"user": user})
}
