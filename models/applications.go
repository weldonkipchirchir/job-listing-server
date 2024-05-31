package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Application struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	JobID   primitive.ObjectID `json:"jobId" bson:"jobId" validate:"required"`
	JobName string             `json:"jobName" bson:"jobName" validate:"required"`
	UserID  primitive.ObjectID `json:"userId" bson:"userId" validate:"required"`
	Name    string             `json:"name" bson:"name" validate:"required"`
	Status  string             `json:"status" bson:"status" validate:"required,min=3"`
	Resume  PDF                `json:"resume" bson:"resume" validate:"required"`
	Email   string             `json:"email" bson:"email" validate:"email"`
	Company string             `json:"company" bson:"company" validate:"company"`
}

type PDF struct {
	Filename    string `json:"filename" bson:"filename" validate:"required"`
	ContentType string `json:"contentType" bson:"contentType" validate:"required"`
	Data        []byte `json:"data" bson:"data" validate:"required"`
}

type ApplicationUserResponse struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	JobID   primitive.ObjectID `json:"jobId" bson:"jobId" validate:"required"`
	Status  string             `json:"status" bson:"status" validate:"required,min=3"`
	JobName string             `json:"jobName" bson:"jobName" validate:"required"`
	Company string             `json:"company" bson:"company" validate:"company"`
}
type ApplicationAdminResponse struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	JobID   primitive.ObjectID `json:"jobId" bson:"jobId" validate:"required"`
	Status  string             `json:"status" bson:"status" validate:"required,min=3"`
	JobName string             `json:"jobName" bson:"jobName" validate:"required"`
	Name    string             `json:"name" bson:"name" validate:"required"`
	Email   string             `json:"email" bson:"email" validate:"email"`
	Resume  PDF                `json:"resume" bson:"resume" validate:"required"`
}
