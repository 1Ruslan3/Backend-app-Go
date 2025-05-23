package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"premium_cars_app/pkg/utils"
)

func AuthRequired(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token"})
			return
		}

		username, err := utils.ParseJWT(token)
		if err != nil || !utils.TokenExists(rdb, token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		_ = utils.RefreshTokenTTL(rdb, token)

		c.Set("username", username)
		c.Next()
	}
}
