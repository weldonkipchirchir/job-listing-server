package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/controllers"
	"github.com/weldonkipchirchir/job-listing-server/db"
	"github.com/weldonkipchirchir/job-listing-server/handler"
	"github.com/weldonkipchirchir/job-listing-server/middleware"
)

func ApplicationRoutes(router *gin.Engine) {
	errorHandler := handler.NewErrorHandler()
	applicationHandler := controllers.NewApplicationHandler(db.GetCollection("applications"), errorHandler)
	applicationGroup := router.Group("/api/v1/applications")
	applicationGroup.Use(middleware.Authentication())
	{
		applicationGroup.GET("/admin", applicationHandler.GetAdminApplications)
		applicationGroup.GET("/admin/info", applicationHandler.AdminInformation)
		applicationGroup.GET("/admin/search", applicationHandler.SearchAdminApplications)
		applicationGroup.GET("/", applicationHandler.GetApplications)
		applicationGroup.POST("/", applicationHandler.CreateApplications)
		applicationGroup.PUT("/admin/:id", applicationHandler.EditApplication)
		applicationGroup.DELETE("/:id", applicationHandler.DeleteApplication)
	}
}
