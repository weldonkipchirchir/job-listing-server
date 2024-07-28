package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/handler"
	"github.com/weldonkipchirchir/job-listing-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SearchLogHandler struct {
	Collection   *mongo.Collection
	errorHandler *handler.ErrorHandler
}

func NewSearchLogHandler(collection *mongo.Collection, errorHandler *handler.ErrorHandler) *SearchLogHandler {
	return &SearchLogHandler{
		Collection:   collection,
		errorHandler: errorHandler,
	}
}

func (sh *SearchLogHandler) CreateSearchLog(c *gin.Context) {
	var searchLog models.SearchLog

	userId, ok := c.Get("id")
	if !ok {
		sh.errorHandler.HandleBadRequest(c)
		return
	}

	UserIdStr, ok := userId.(string)
	if !ok {
		sh.errorHandler.HandleBadRequest(c)
		return
	}

	UserID, err := primitive.ObjectIDFromHex(UserIdStr)
	if err != nil {
		sh.errorHandler.HandleBadRequest(c)
		return
	}

	if err := c.BindJSON(&searchLog); err != nil {
		sh.errorHandler.HandleInternalServerError(c)
		return
	}

	searchLog.ID = primitive.NewObjectID()
	searchLog.UserID = UserID

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = sh.Collection.InsertOne(ctx, searchLog)
	if err != nil {
		sh.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusCreated, "searchlog created")
}

func (sh *SearchLogHandler) GetSearchLog(c *gin.Context) {
	var searchLog []models.SearchLog

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := sh.Collection.Find(ctx, bson.M{})
	if err != nil {
		sh.errorHandler.HandleInternalServerError(c)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &searchLog); err != nil {
		sh.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, searchLog)
}

// update searchlog
func (sh *SearchLogHandler) UpdateSearchLog(c *gin.Context) {

}
