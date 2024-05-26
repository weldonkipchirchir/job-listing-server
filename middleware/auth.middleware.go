package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/auth"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		access_token := c.GetHeader("Authorization")
		if access_token == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if len(access_token) > 7 && access_token[:7] == "Bearer " {
			access_token = access_token[7:]
		} else {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		claims, msg := auth.ValidateToken(access_token)
		if msg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			c.Abort()
			return
		}

		if claims.ExpiresAt < time.Now().Unix() {
			refresh_token := c.GetHeader("refreshToken")
			if refresh_token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not provided"})
				c.Abort()
				return
			}

			newAccessToken, err := auth.UpdateToken(refresh_token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
			// Set the new access token as a cookie
			c.SetCookie("token", newAccessToken, int(time.Hour)*30, "/", "", false, true)
			c.SetCookie("refreshToken", refresh_token, int(time.Hour)*90, "/", "", false, true)

			c.Set("id", claims.Id)
			c.Set("role", claims.Role)
			c.Set("email", claims.Email)
			c.Next()
		}

		c.Set("id", claims.Id)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Next()
	}
}
