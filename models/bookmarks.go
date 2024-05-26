package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bookmark struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	JobID  primitive.ObjectID `json:"jobId" bson:"jobId" validate:"required"`
	UserID primitive.ObjectID `json:"userId" bson:"userId" validate:"required"`
}
