package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/db"
	"github.com/weldonkipchirchir/job-listing-server/handler"
	"github.com/weldonkipchirchir/job-listing-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookmarkHandler struct {
	Collection   *mongo.Collection
	errorHandler *handler.ErrorHandler
}

func NewBookmarkHandler(collection *mongo.Collection, errorHandler *handler.ErrorHandler) *BookmarkHandler {
	return &BookmarkHandler{
		Collection:   collection,
		errorHandler: errorHandler,
	}
}

// create bookmark
func (bh *BookmarkHandler) CreateBookmark(c *gin.Context) {
	userId, ok := c.Get("id")
	if !ok {
		bh.errorHandler.HandleBadRequest(c)
		return
	}

	UserIdStr, ok := userId.(string)
	if !ok {
		bh.errorHandler.HandleBadRequest(c)
		return
	}
	ObjectUserId, err := primitive.ObjectIDFromHex(UserIdStr)
	if err != nil {
		bh.errorHandler.HandleInternalServerError(c)
		return
	}

	var bookmark models.Bookmark

	if err := c.BindJSON(&bookmark); err != nil {
		bh.errorHandler.HandleInternalServerError(c)
		return
	}

	bookmark.ID = primitive.NewObjectID()
	bookmark.UserID = ObjectUserId

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = bh.Collection.InsertOne(ctx, bookmark)
	if err != nil {
		bh.errorHandler.HandleInternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, bookmark)
}

// GetBookMarks retrieves bookmarks for a user
func (bh *BookmarkHandler) GetBookmarks(c *gin.Context) {
	userId, ok := c.Get("id")
	if !ok {
		bh.errorHandler.HandleBadRequest(c)
		return
	}

	UserIdStr, ok := userId.(string)
	if !ok {
		bh.errorHandler.HandleBadRequest(c)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(UserIdStr)
	if err != nil {
		bh.errorHandler.HandleBadRequest(c)
		return
	}

	var bookmarks []models.Bookmark
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"userId": objectId}
	cursor, err := bh.Collection.Find(ctx, filter)
	if err != nil {
		bh.errorHandler.HandleInternalServerError(c)
		return
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &bookmarks)
	if err != nil {
		bh.errorHandler.HandleInternalServerError(c)
		return
	}

	var jobIDs []primitive.ObjectID
	for _, bookmark := range bookmarks {
		jobIDs = append(jobIDs, bookmark.JobID)
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, []models.Job{})
		return
	}

	var jobs []models.Job
	jobFilter := bson.M{"_id": bson.M{"$in": jobIDs}}
	jobCursor, err := db.DB.Collection("jobs").Find(ctx, jobFilter)
	if err != nil {
		bh.errorHandler.HandleInternalServerError(c)
		return
	}
	defer jobCursor.Close(ctx)

	err = jobCursor.All(ctx, &jobs)
	if err != nil {
		bh.errorHandler.HandleInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, jobs)
}

// update
func (bh *BookmarkHandler) DeleteBookmark(c *gin.Context) {
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		bh.errorHandler.HandleBadRequest(c)
		return
	}

	role, ok := c.MustGet("role").(string)
	if !ok {
		bh.errorHandler.HandleBadRequest(c)
		return
	}
	if role != "user" {
		bh.errorHandler.HandleUnauthorized(c)
		return
	}

	userId, ok := c.MustGet("id").(string)
	if !ok {
		bh.errorHandler.HandleBadRequest(c)
		return
	}

	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		bh.errorHandler.HandleBadRequest(c)
		return
	}

	filter := bson.M{"userId": userObjectId, "jobId": objectID}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = bh.Collection.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			bh.errorHandler.HandleNotFound(c)
			return
		}
		bh.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "job deleted"})
}
