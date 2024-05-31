package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name" validate:"required,min=3"`
	Email    string             `json:"email" bson:"email" validate:"email"`
	Phone    string             `json:"phone" bson:"phone" validate:"phone"`
	Address  string             `json:"address" bson:"address" validate:"address"`
	Password string             `json:"password" bson:"password" validate:"required"`
	Role     string             `json:"role" bson:"role" validate:"role"`
}

type UserResponse struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name" bson:"name" validate:"required,min=3"`
	Email string             `json:"email" bson:"email" validate:"required,email"`
	Role  string             `json:"role" bson:"role" validate:"required"`
}
