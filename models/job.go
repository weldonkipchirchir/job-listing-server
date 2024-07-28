package models

import (
	"github.com/weldonkipchirchir/job-listing-server/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Job struct {
	ID                    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	JobName               string             `json:"jobName" bson:"jobName" validate:"required,min=3"`
	Type                  string             `json:"type" bson:"type" validate:"required,min=3"`
	Location              string             `json:"location" bson:"location" validate:"required,min=3"`
	SalaryHigh            string             `json:"salaryHigh" bson:"salaryHigh" validate:"required"`
	SalaryLow             string             `json:"salaryLow" bson:"salaryLow" validate:"required"`
	Company               string             `json:"company" bson:"company" validate:"required,min=3"`
	ImageLink             string             `json:"imageLink" bson:"imageLink" validate:"required"`
	Sponsored             bool               `json:"sponsored" bson:"sponsored" validate:"required"`
	UserID                primitive.ObjectID `json:"userId" bson:"userId" validate:"required"`
	Currency              utils.Currency     `json:"currency" bson:"currency" validate:"required,min3"`
	MandatoryRequirements []string           `json:"mandatoryRequirements" bson:"mandatoryRequirements" validate:"required"`
	OptionalRequirements  []string           `json:"optionalRequirements" bson:"optionalRequirements"`
	JobDescription        string             `json:"jobDescription" bson:"jobDescription" validate:"required"`
	Industry              string             `json:"industry" bson:"industry" validate:"required"`
	DaysAgo               int                `json:"daysAgo" bson:"-"`
}
