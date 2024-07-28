package controllers

import (
	"bytes"
	"context"
	"io"
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
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

type ApplicationHandler struct {
	Collection   *mongo.Collection
	errorHandler *handler.ErrorHandler
}

func NewApplicationHandler(collection *mongo.Collection, errorHandler *handler.ErrorHandler) *ApplicationHandler {
	return &ApplicationHandler{
		Collection:   collection,
		errorHandler: errorHandler,
	}
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

	objectId, err := primitive.ObjectIDFromHex(userIDStr)
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
	application.UserID = objectId
	application.Email = email
	application.Name = name

	statusCapitalize := utils.CapitalizeFirstLetter(application.Status)
	application.Status = statusCapitalize

	// Store the resume PDF in GridFS
	bucket, err := gridfs.NewBucket(ah.Collection.Database())
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	uploadStream, err := bucket.OpenUploadStream(application.Resume.Filename)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	defer uploadStream.Close()

	_, err = uploadStream.Write(application.Resume.Data)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	application.Resume = models.PDF{
		Filename:    application.Resume.Filename,
		ContentType: application.Resume.ContentType,
		Data:        []byte(uploadStream.FileID.(primitive.ObjectID).Hex()), // Storing the ObjectID as a string
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = ah.Collection.InsertOne(ctx, application)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	c.JSON(http.StatusCreated, application)
}

func (ah *ApplicationHandler) GetAdminApplications(c *gin.Context) {
	// Check role
	role, ok := c.MustGet("role").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}
	if role != "admin" {
		ah.errorHandler.HandleUnauthorized(c)
		return
	}

	// Retrieve user ID
	userId, ok := c.MustGet("id").(string)
	if !ok {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	// Convert user ID to ObjectID
	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		ah.errorHandler.HandleBadRequest(c)
		return
	}

	// Query jobs associated with the user
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

	// If no jobs found, return empty response
	if len(jobs) == 0 {
		c.JSON(http.StatusOK, jobs)
		return
	}

	// Extract job IDs
	var jobIDs []primitive.ObjectID
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	// Query applications for the user's jobs
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

	// If no applications found, return appropriate message
	if len(applications) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No applications available for the user's jobs"})
		return
	}

	// Prepare response with PDF data encoded in base64
	var applicationResponses []models.ApplicationAdminResponse
	bucket, err := gridfs.NewBucket(ah.Collection.Database())
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	for _, application := range applications {
		var resumeData []byte
		if application.Resume.Data != nil {
			resumeID, err := primitive.ObjectIDFromHex(string(application.Resume.Data))
			if err != nil {
				ah.errorHandler.HandleInternalServerError(c)
				return
			}

			var buf bytes.Buffer
			downloadStream, err := bucket.OpenDownloadStream(resumeID)
			if err != nil {
				ah.errorHandler.HandleInternalServerError(c)
				return
			}
			defer downloadStream.Close()

			_, err = io.Copy(&buf, downloadStream)
			if err != nil {
				ah.errorHandler.HandleInternalServerError(c)
				return
			}

			resumeData = buf.Bytes()
		}

		applicationResponse := models.ApplicationAdminResponse{
			ID:      application.ID,
			JobID:   application.JobID,
			Resume:  models.PDF{Filename: application.Resume.Filename, ContentType: application.Resume.ContentType, Data: resumeData}, // Directly assign PDF data
			Name:    application.Name,
			Email:   application.Email,
			Status:  application.Status,
			JobName: application.JobName,
		}
		applicationResponses = append(applicationResponses, applicationResponse)
	}

	// Return response
	c.JSON(http.StatusOK, applicationResponses)
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
		// If new resume data is provided, update it
		bucket, err := gridfs.NewBucket(ah.Collection.Database())
		if err != nil {
			ah.errorHandler.HandleInternalServerError(c)
			return
		}

		uploadStream, err := bucket.OpenUploadStream(updateApplication.Resume.Filename)
		if err != nil {
			ah.errorHandler.HandleInternalServerError(c)
			return
		}
		defer uploadStream.Close()

		_, err = uploadStream.Write(updateApplication.Resume.Data)
		if err != nil {
			ah.errorHandler.HandleInternalServerError(c)
			return
		}

		updateFields["resume.data"] = uploadStream.FileID.(primitive.ObjectID).Hex()
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
	c.JSON(http.StatusOK, gin.H{"message": "application deleted"})
}

func (ah *ApplicationHandler) SearchAdminApplications(c *gin.Context) {
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

	// Query jobs associated with the user
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

	// If no jobs found, return empty response
	if len(jobs) == 0 {
		c.JSON(http.StatusOK, jobs)
		return
	}

	// Extract job IDs
	var jobIDs []primitive.ObjectID
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	var filters bson.M = bson.M{"jobId": bson.M{"$in": jobIDs}}

	if email := c.Query("email"); email != "" {
		filters["email"] = bson.M{"$regex": email, "$options": "i"}
	}

	var application []models.Application
	cursor, err = ah.Collection.Find(ctx, filters)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &application); err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	if len(application) == 0 {
		application = []models.Application{}
	}
	c.JSON(http.StatusOK, application)
}

func (ah *ApplicationHandler) AdminInformation(c *gin.Context) {
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

	// Query jobs associated with the user
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

	var activeJobs int        // Initialize activeJobs to 0
	var totalApplications int // Initialize totalApplications to 0
	var newApplication int    // Initialize newApplication to 0

	// If no jobs found, return empty response
	if len(jobs) == 0 {
		res := gin.H{
			"activeJobs":        activeJobs,
			"totalApplications": totalApplications,
			"totalUsers":        totalApplications, // Assuming totalUsers should have the same value as totalApplications
			"newApplication":    newApplication,
		}
		c.JSON(http.StatusOK, res)
		return
	}

	// Extract job IDs
	var jobIDs []primitive.ObjectID
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
		activeJobs += 1
	}

	var filters bson.M = bson.M{"jobId": bson.M{"$in": jobIDs}}

	var application []models.Application
	cursor, err = ah.Collection.Find(ctx, filters)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &application); err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	// Update totalApplications based on the length of the application slice
	totalApplications = len(application)

	var filtersStatus bson.M = bson.M{"status": "Pending", "jobId": bson.M{"$in": jobIDs}}

	var applicationStatus []models.Application
	cursor, err = ah.Collection.Find(ctx, filtersStatus)
	if err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &applicationStatus); err != nil {
		ah.errorHandler.HandleInternalServerError(c)
		return
	}

	// Update newApplication based on the length of the applicationStatus slice
	newApplication = len(applicationStatus)

	res := gin.H{
		"activeJobs":        activeJobs,
		"totalApplications": totalApplications,
		"totalUsers":        totalApplications, // Assuming totalUsers should have the same value as totalApplications
		"newApplication":    newApplication,
	}

	c.JSON(http.StatusOK, res)
}
