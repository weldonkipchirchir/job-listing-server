package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/auth"
	"github.com/weldonkipchirchir/job-listing-server/models"
	"github.com/weldonkipchirchir/job-listing-server/services"
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

	token, refreshToken, err := auth.TokenGenerator(userIDString, user.FirstName, user.Email, user.Role)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	userResponse := models.UserResponse{
		ID:    user.ID,
		Name:  user.FirstName,
		Email: user.Email,
		Role:  user.Role,
	}

	res := gin.H{
		"access_token":  token,
		"refresh_token": refreshToken,
		"user":          userResponse,
	}

	// Set the new access token as a cookie
	c.SetCookie("token", token, int(time.Hour)*30, "/", "", false, true)
	c.SetCookie("refreshToken", refreshToken, int(time.Hour)*90, "/", "", false, true)

	res["message"] = "Login successful"

	c.JSON(http.StatusOK, res)
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
