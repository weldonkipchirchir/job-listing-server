package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/controllers"
	"github.com/weldonkipchirchir/job-listing-server/db"
	"github.com/weldonkipchirchir/job-listing-server/handler"
	"github.com/weldonkipchirchir/job-listing-server/middleware"
)

func BookmarksRoutes(router *gin.Engine) {
	errorHandler := handler.NewErrorHandler()
	bookmarkHandler := controllers.NewBookmarkHandler(db.GetCollection("bookmarks"), errorHandler)
	bookmarkGroup := router.Group("/api/v1/bookmarks")
	bookmarkGroup.Use(middleware.Authentication())
	{
		bookmarkGroup.GET("/", bookmarkHandler.GetBookMarks)
		bookmarkGroup.POST("/", bookmarkHandler.CreateBookmark)
		bookmarkGroup.DELETE("/:id", bookmarkHandler.DeleteBookmark)
	}
}
