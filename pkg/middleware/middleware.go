package middleware

import (
	"github.com/azmiagr/unesco-hackathon/internal/service"
	"github.com/azmiagr/unesco-hackathon/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type Interface interface {
	Cors() gin.HandlerFunc
	AuthenticateUser(c *gin.Context)
	OnlyRoles(allowedRoles ...string) gin.HandlerFunc
	OnlyAdmin() gin.HandlerFunc
}

const (
	userContextKey     = "user"
	roleNameContextKey = "role_name"
)

type middleware struct {
	service *service.Service
	jwtAuth jwt.Interface
}

func Init(service *service.Service, jwtAuth jwt.Interface) Interface {
	return &middleware{
		service: service,
		jwtAuth: jwtAuth,
	}
}
