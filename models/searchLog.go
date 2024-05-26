package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SearchLog struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"userId" bson:"userId" validate:"required"`
	JobID  primitive.ObjectID `json:"jobId" bson:"jobId" validate:"required"`
}
