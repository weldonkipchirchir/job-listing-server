package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/handler"
	"github.com/weldonkipchirchir/job-listing-server/models"
	"github.com/weldonkipchirchir/job-listing-server/utils"
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
		log.Println("error in bind", err)
		return
	}

	currencyCapitalize := utils.UpperCaseString(string(job.Currency))

	job.Currency = utils.Currency(currencyCapitalize)

	// Validate currency
	if !job.Currency.IsValid() {
		jh.errorHandler.HandleBadRequest(c)
		log.Println("error in cuu")
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

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	role, ok := c.Get("role")
	if !ok {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}
	if role != "admin" {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}

	userId, ok := c.Get("id")
	if !ok {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}

	userIdStr := userId.(string)

	userObjId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	var existingJob models.Job
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = jh.Collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&existingJob)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			jh.errorHandler.HandleNotFound(c)
			return
		}
		jh.errorHandler.HandleInternalServerError(c)
		return
	}

	if existingJob.UserID != userObjId {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}

	var updateJob models.Job
	if err := c.ShouldBindJSON(&updateJob); err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	updateFields := bson.M{}
	if updateJob.JobName != "" {
		updateFields["jobName"] = updateJob.JobName
	}
	if updateJob.Type != "" {
		updateFields["type"] = updateJob.Type
	}
	if updateJob.Location != "" {
		updateFields["location"] = updateJob.Location
	}
	if updateJob.SalaryHigh != "" {
		updateFields["salaryHigh"] = updateJob.SalaryHigh
	}
	if updateJob.SalaryLow != "" {
		updateFields["salaryLow"] = updateJob.SalaryLow
	}
	if updateJob.Company != "" {
		updateFields["company"] = updateJob.Company
	}
	if updateJob.ImageLink != "" {
		updateFields["imageLink"] = updateJob.ImageLink
	}
	if updateJob.Sponsored {
		updateFields["sponsored"] = updateJob.Sponsored
	}
	if updateJob.Currency != "" {
		updateFields["currency"] = updateJob.Currency
	}
	if len(updateJob.MandatoryRequirements) > 0 {
		updateFields["mandatoryRequirements"] = updateJob.MandatoryRequirements
	}
	if len(updateJob.OptionalRequirements) > 0 {
		updateFields["optionalRequirements"] = updateJob.OptionalRequirements
	}
	if updateJob.JobDescription != "" {
		updateFields["jobDescription"] = updateJob.JobDescription
	}
	if updateJob.Industry != "" {
		updateFields["industry"] = updateJob.Industry
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	update := bson.M{"$set": updateFields}
	_, err = jh.Collection.UpdateOne(context.TODO(), bson.M{"_id": objectId}, update)
	if err != nil {
		jh.errorHandler.HandleInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job updated"})
}

func (jh *JobHandler) DeleteJob(c *gin.Context) {
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	role, ok := c.Get("role")
	if !ok {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}
	if role != "admin" {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}

	userId, ok := c.Get("id")
	if !ok {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}

	userIdStr := userId.(string)

	userObjId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	var existingJob models.Job
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = jh.Collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingJob)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			jh.errorHandler.HandleNotFound(c)
			return
		}
		jh.errorHandler.HandleInternalServerError(c)
		return
	}

	if existingJob.UserID != userObjId {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}

	_, err = jh.Collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			jh.errorHandler.HandleNotFound(c)
			return
		}
		jh.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "job deleted"})
}

func (jh *JobHandler) GetJobById(c *gin.Context) {
	id := c.Param("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		jh.errorHandler.HandleBadRequest(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectId}

	var job models.Job

	err = jh.Collection.FindOne(ctx, filter).Decode(&job)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			jh.errorHandler.HandleNotFound(c)
			return
		}
		jh.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, job)
}

func (jh *JobHandler) GetAdminsJobs(c *gin.Context) {

	role, ok := c.MustGet("role").(string)
	if !ok {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	if role != "admin" {
		jh.errorHandler.HandleUnauthorized(c)
		return
	}

	userId, ok := c.MustGet("id").(string)
	if !ok {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		jh.errorHandler.HandleBadRequest(c)
		return
	}

	filter := bson.M{"userId": objectId}

	var jobs []models.Job
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := jh.Collection.Find(ctx, filter)
	if err != nil {
		jh.errorHandler.HandleInternalServerError(c)
		return
	}
	if err = cursor.All(ctx, &jobs); err != nil {
		jh.errorHandler.HandleInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, jobs)
}

func (jh *JobHandler) GetSponsoredJobs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var jobs []models.Job
	filter := bson.M{"sponsored": true}
	cursor, err := jh.Collection.Find(ctx, filter)
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
