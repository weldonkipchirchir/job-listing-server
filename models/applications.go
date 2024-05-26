package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Application struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	JobID  primitive.ObjectID `json:"jobId" bson:"jobId" validate:"required"`
	UserID primitive.ObjectID `json:"userId" bson:"userId" validate:"required"`
	Status string             `json:"status" bson:"status" validate:"required,min=3"`
	Resume PDF                `json:"resume" bson:"resume" validate:"required"`
}

type PDF struct {
	Filename    string `json:"filename" bson:"filename" validate:"required"`
	ContentType string `json:"contentType" bson:"contentType" validate:"required"`
	Data        []byte `json:"data" bson:"data" validate:"required"`
}
