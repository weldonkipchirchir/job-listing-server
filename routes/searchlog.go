package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/controllers"
	"github.com/weldonkipchirchir/job-listing-server/db"
	"github.com/weldonkipchirchir/job-listing-server/handler"
)

func SearchLog(router *gin.Engine) {
	errorHandler := handler.NewErrorHandler()
	searchlogHandler := controllers.NewSearchLogHandler(db.GetCollection("searchlog"), errorHandler)
	searchLogGroup := router.Group("/api/v1/search")
	{
		searchLogGroup.GET("/", searchlogHandler.GetSearchLog)
		searchLogGroup.POST("/", searchlogHandler.CreateSearchLog)
	}
}
