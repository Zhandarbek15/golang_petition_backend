package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func NewRoleAdminMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.Value("Role")
		if role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role is empty"})
		}

		if role != "Admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No access to this resource"})
		}
		c.Next()
	}
}
