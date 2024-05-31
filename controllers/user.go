package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/auth"
	"github.com/weldonkipchirchir/job-listing-server/db"
	"github.com/weldonkipchirchir/job-listing-server/models"
	"github.com/weldonkipchirchir/job-listing-server/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Register(c *gin.Context) {
	var user models.User

	user.ID = primitive.NewObjectID()

	if err := c.BindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := services.RegisterUser(&user)
	if err == services.ErrEmailTaken {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "otherError"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})

}

func Login(c *gin.Context) {
	var loginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"error":   "Invalid json format",
			"message": err.Error(),
		})
		return
	}

	user, err := services.LoginUser(loginRequest.Email, loginRequest.Password)

	if err == services.ErrUserNotFound || err == services.ErrInvalidCredentials {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	userIDString := user.ID.Hex()

	token, refreshToken, err := auth.TokenGenerator(userIDString, user.Name, user.Email, user.Role)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	userResponse := models.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}

	res := gin.H{
		"access_token":  token,
		"refresh_token": refreshToken,
		"user":          userResponse,
	}

	expirationTimeToken := time.Now().Add(24 * time.Hour * 7)
	expirationTimeRefreshToken := time.Now().Add(24 * time.Hour * 30)
	c.SetCookie("token", token, int(time.Until(expirationTimeToken).Seconds()), "/", "", false, true)
	c.SetCookie("refreshToken", refreshToken, int(time.Until(expirationTimeRefreshToken).Seconds()), "/", "", false, true)

	res["message"] = "Login successful"

	c.JSON(http.StatusOK, res)
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.SetCookie("refreshToken", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// update user settings
func Settings(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	userEmail := email.(string)

	userId, ok := c.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bad request"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	err = services.UserSettings(userEmail)
	if err == services.ErrUserNotFound || err == services.ErrInvalidCredentials {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bad request"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	updateFields := bson.M{}

	// Check and add each field to the update document if it's not empty
	if user.Name != "" {
		updateFields["name"] = user.Name
	}
	if user.Email != "" {
		updateFields["email"] = user.Email
	}
	if user.Phone != "" {
		updateFields["phone"] = user.Phone
	}
	if user.Address != "" {
		updateFields["address"] = user.Address
	}
	if user.Role != "" {
		updateFields["role"] = user.Role
	}

	// Check if the password field is non-empty
	if user.Password != "" {
		// Hash the password
		hashedPassword, err := services.HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		// Set the hashed password
		updateFields["password"] = hashedPassword
	}

	// If there are no fields to update, return bad request
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": updateFields}
	_, err = db.DB.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}
