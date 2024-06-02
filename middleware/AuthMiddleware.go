package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"petition_api/utils/auth"
)

// NewAuthMiddleware возвращает middleware функцию для аутентификации
func NewAuthMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCookie, err := c.Request.Cookie("access_token") // Чтение куки access_token из запроса
		if err != nil || authCookie == nil {
			logger.Warn("Access token cookie is missing or invalid")
			c.JSON(http.StatusForbidden, gin.H{"error": "Access token cookie is missing or invalid"})
			c.Abort()
			return
		}

		tokenString := authCookie.Value
		// Валидация токена
		claims, code, err := auth.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(code, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}

		logger.Debug(fmt.Sprintf("Authorized user: ID: %v : Role %v", claims.ID, claims.Role))

		c.Set("ID", claims.ID)
		c.Set("Role", claims.Role)
		c.Next()
	}
}
