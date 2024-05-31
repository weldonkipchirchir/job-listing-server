package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/db"
	"github.com/weldonkipchirchir/job-listing-server/handler"
	"github.com/weldonkipchirchir/job-listing-server/models"
	"github.com/weldonkipchirchir/job-listing-server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ApplicationHandler struct {
	Collection   *mongo.Collection
	errorHandler *handler.ErrorHandler
}

func NewApplicationHandler(collection *mongo.Collection, erroHandler *handler.ErrorHandler) *ApplicationHandler {
	return &ApplicationHandler{
		Collection:   collection,
		errorHandler: erroHandler,
	}
}
func (ah *ApplicationHandler) GetAdminApplications(c *gin.Context) {
	role, ok := c.MustGet("role").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}
	if role != "admin" {
		ah.errorHandler.HandleUnauthorized(c)
		return
	}

	userId, ok := c.MustGet("id").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	var jobs []models.Job
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.M{"userId": objectId}
	cursor, err := db.DB.Collection("jobs").Find(ctx, filter)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &jobs)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	// Check if the user has no jobs
	if len(jobs) == 0 {
		c.JSON(http.StatusOK, jobs)
		return
	}

	var jobIDs []primitive.ObjectID
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	appFilter := bson.M{"jobId": bson.M{"$in": jobIDs}}
	appCursor, err := ah.Collection.Find(ctx, appFilter)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	defer appCursor.Close(ctx)

	var applications []models.Application
	err = appCursor.All(ctx, &applications)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	// Check if there are no applications for the user's job IDs
	if len(applications) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No applications available for the user's jobs"})
		return
	}

	var applicationResponses []models.ApplicationAdminResponse
	for _, application := range applications {
		applicationResponse := models.ApplicationAdminResponse{
			ID:      application.ID,
			JobID:   application.JobID,
			Resume:  application.Resume,
			Name:    application.Name,
			Email:   application.Email,
			Status:  application.Status,
			JobName: application.JobName,
		}
		applicationResponses = append(applicationResponses, applicationResponse)
	}

	c.JSON(http.StatusOK, applicationResponses)
}

func (ah *ApplicationHandler) CreateApplications(c *gin.Context) {

	role, ok := c.MustGet("role").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}
	name, ok := c.MustGet("name").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}
	email, ok := c.MustGet("email").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	if role != "user" {
		ah.errorHandler.HandleUnauthorized(c)
		return
	}

	userId, ok := c.Get("id")
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	userIDStr, ok := userId.(string)
	if !ok {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	ObjectId, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	var application models.Application

	if err := c.BindJSON(&application); err != nil {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	application.ID = primitive.NewObjectID()
	application.UserID = ObjectId
	application.Email = email
	application.Name = name

	statusCapitalize := utils.CapitalizeFirstLetter(application.Status)
	application.Status = statusCapitalize

	//check if the job exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = ah.Collection.InsertOne(ctx, application)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusCreated, application)
}
func (ah *ApplicationHandler) GetApplications(c *gin.Context) {
	role, ok := c.MustGet("role").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}
	if role != "user" {
		ah.errorHandler.HandleUnauthorized(c)
		return
	}

	userId, ok := c.MustGet("id").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	var applications []models.Application
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.M{"userId": objectId}
	cursor, err := ah.Collection.Find(ctx, filter)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &applications)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	var applicationResponses []models.ApplicationUserResponse

	for _, application := range applications {
		applicationResponse := models.ApplicationUserResponse{
			ID:      application.ID,
			JobID:   application.JobID,
			Status:  application.Status,
			JobName: application.JobName,
			Company: application.Company,
		}
		applicationResponses = append(applicationResponses, applicationResponse)
	}

	c.JSON(http.StatusOK, applicationResponses)
}

func (ah *ApplicationHandler) EditApplication(c *gin.Context) {
	id := c.Params.ByName("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	role, ok := c.Get("role")
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	if role != "admin" {
		ah.errorHandler.HandleUnauthorized(c)
		return
	}

	userId, ok := c.Get("id")
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	_, err = primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	var application models.Application
	filter := bson.M{"_id": objectID}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = ah.Collection.FindOne(ctx, filter).Decode(&application)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ah.errorHandler.HandleNotFound(c)
			return
		}
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	var updateApplication models.Application
	if err := c.BindJSON(&updateApplication); err != nil {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	updateFields := bson.M{}

	if updateApplication.Name != "" {
		updateFields["name"] = updateApplication.Name
	}
	if updateApplication.Status != "" {
		updateFields["status"] = utils.CapitalizeFirstLetter(updateApplication.Status)
	}
	if updateApplication.Resume.Filename != "" {
		updateFields["resume.filename"] = updateApplication.Resume.Filename
	}
	if updateApplication.Resume.ContentType != "" {
		updateFields["resume.contentType"] = updateApplication.Resume.ContentType
	}
	if len(updateApplication.Resume.Data) > 0 {
		updateFields["resume.data"] = updateApplication.Resume.Data
	}
	if updateApplication.Email != "" {
		updateFields["email"] = updateApplication.Email
	}
	if updateApplication.Company != "" {
		updateFields["company"] = updateApplication.Company
	}
	if updateApplication.JobName != "" {
		updateFields["jobName"] = updateApplication.JobName
	}

	update := bson.M{"$set": updateFields}
	_, err = ah.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application updated"})
}

// delete application
func (ah *ApplicationHandler) DeleteApplication(c *gin.Context) {
	id := c.Params.ByName("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.M{"_id": objectId}

	_, err = ah.Collection.DeleteOne(ctx, filter)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ah.errorHandler.HandleNotFound(c)
			return
		}
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "application delete"})
}
