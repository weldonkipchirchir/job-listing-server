package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/controllers"
	"github.com/weldonkipchirchir/job-listing-server/db"
	"github.com/weldonkipchirchir/job-listing-server/handler"
	"github.com/weldonkipchirchir/job-listing-server/middleware"
)

func JobRoutes(router *gin.Engine) {
	errorHandler := handler.NewErrorHandler()
	jobHandler := controllers.NewJobCollection(db.GetCollection("jobs"), errorHandler)
	jobGroup := router.Group("/api/v1/jobs")
	jobGroup.GET("/sponsored", jobHandler.GetSponsoredJobs)
	jobGroup.Use(middleware.Authentication())
	{
		jobGroup.GET("/", jobHandler.GetAllJobs)
		jobGroup.GET("/:id", jobHandler.GetJobById)
		jobGroup.GET("/admin", jobHandler.GetAdminsJobs)
		jobGroup.POST("/create", jobHandler.CreateJob)
		jobGroup.PUT("/admin/:id", jobHandler.Updatejob)
		jobGroup.DELETE("/admin/:id", jobHandler.DeleteJob)
	}
}
