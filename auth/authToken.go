package auth

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SignedDetails struct {
	Id    string
	Name  string
	Email string
	Role  string
	jwt.StandardClaims
}

var secretKey = os.Getenv("SECRET_KEY")

func TokenGenerator(id string, name string, email string, role string) (token string, refreshToken string, err error) {
	claims := &SignedDetails{
		Id:    id,
		Name:  name,
		Email: email,
		Role:  role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
			Issuer:    "jobly",
			IssuedAt:  time.Now().Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		Id:    id,
		Name:  name,
		Email: email,
		Role:  role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix(),
			Issuer:    "jobly",
			IssuedAt:  time.Now().Unix(),
		},
	}

	var tokenString string
	tokenString, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	var refreshTokenString string
	refreshTokenString, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	return tokenString, refreshTokenString, err
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		msg = err.Error()
		return nil, msg
	}

	claims, ok := token.Claims.(*SignedDetails)

	if !ok {
		msg = "The token is invalid"
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "The token is expired"
		return
	}
	return claims, msg
}

// update token when token is valid but is expired
func UpdateToken(refreshToken string) (token string, err error) {
	claims, msg := ValidateToken(refreshToken)
	if msg != "" {
		return "", errors.New(msg)
	}

	// Create new access token with updated expiration time
	newClaims := &SignedDetails{
		Id:    claims.Id,
		Name:  claims.Name,
		Email: claims.Email,
		Role:  claims.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			Issuer:    "jobly",
			IssuedAt:  time.Now().Unix(),
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims).SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return token, nil
}
