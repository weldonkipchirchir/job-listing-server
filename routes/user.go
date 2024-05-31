package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/controllers"
	"github.com/weldonkipchirchir/job-listing-server/middleware"
)

func SetUpUsers(router *gin.Engine) {
	users := router.Group("/api/v1/users/")
	{
		// login users
		users.POST("/login", controllers.Login)
		// Create a new user
		users.POST("/register", controllers.Register)
		// logout users
		users.POST("/logout", controllers.Logout)

		// Apply middleware to all subsequent routes within the users group
		users.Use(middleware.Authentication())
		{
			// update user settings
			users.PUT("/settings", controllers.Settings)
		}
	}
}
