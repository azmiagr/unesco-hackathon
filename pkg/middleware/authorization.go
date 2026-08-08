package middleware

import (
	"net/http"

	constants "github.com/azmiagr/unesco-hackathon/pkg/constant"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (m *middleware) OnlyRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleNameValue, exists := c.Get(roleNameContextKey)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		roleName, ok := roleNameValue.(string)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if roleName == role {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "access denied", nil)
		c.Abort()
	}
}

func (m *middleware) OnlyAdmin() gin.HandlerFunc {
	return m.OnlyRoles(constants.RoleAdmin)
}
