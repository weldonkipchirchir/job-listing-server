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

type JobHandler struct {
	Collection   *mongo.Collection
	errorHandler *handler.ErrorHandler
}

func NewJobCollection(collection *mongo.Collection, errorHandler *handler.ErrorHandler) *JobHandler {
	return &JobHandler{
		Collection:   collection,
		errorHandler: errorHandler,
	}
}

func (jh *JobHandler) GetAllJobs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var jobs []models.Job
	cursor, err := jh.Collection.Find(ctx, bson.M{})
	if err != nil {
		jh.errorHandler.HandleInternalServerError(c)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &jobs); err != nil {
		jh.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (jh *JobHandler) CreateJob(c *gin.Context) {
	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	role, ok := c.MustGet("role").(string)

	if !ok {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}
	if role != "admin" {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}
	UserID, ok := c.MustGet("id").(string)
	if !ok {
		jh.errorHandler.HandleInternalServerError(c)
		return
	}

	job.ID = primitive.NewObjectID()

	ownerID, err := primitive.ObjectIDFromHex(UserID)
	if err != nil {
		jh.errorHandler.HandleInternalServerError(c)
		return
	}
	job.UserID = ownerID

	_, err = jh.Collection.InsertOne(context.TODO(), job)
	if err != nil {
		jh.errorHandler.HandleInternalServerError(c) // Use ErrorHandler to handle internal server error
		return
	}
	c.JSON(201, job)
}

func (jh *JobHandler) Updatejob(c *gin.Context) {
	id := c.Param("id")

	ObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	var job models.Job
	cxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = jh.Collection.FindOne(cxt, bson.M{"_id": ObjectId}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			jh.errorHandler.HandleNotFound(c)
			return
		}
		jh.errorHandler.HandleInternalServerError(c)
		return
	}

	var Updatejob models.Job

	if err := c.ShouldBindJSON(&Updatejob); err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	update := bson.M{"$set": Updatejob}

	_, err = jh.Collection.UpdateOne(context.TODO(), bson.M{"_id": ObjectId}, update)

	if err != nil {
		jh.errorHandler.HandleInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking updated"})
}
