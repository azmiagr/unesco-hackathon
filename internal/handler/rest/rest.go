package rest

import (
	"fmt"
	"os"

	"github.com/azmiagr/unesco-hackathon/internal/service"
	"github.com/azmiagr/unesco-hackathon/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Rest struct {
	router     *gin.Engine
	service    *service.Service
	middleware middleware.Interface
}

func NewRest(service *service.Service, middleware middleware.Interface) *Rest {
	return &Rest{
		router:     gin.Default(),
		service:    service,
		middleware: middleware,
	}
}

func (r *Rest) MountEndpoint() {
	r.router.Use(r.middleware.Cors())
	baseUrl := r.router.Group("/api/v1")

	auth := baseUrl.Group("/auth")
	auth.GET("/session", r.GetRegisterSession)
	auth.POST("/register/start", r.StartRegister)
	auth.POST("/register/verify-otp", r.VerifyRegisterOtp)
	auth.POST("/register/avatar", r.SelectRegisterAvatar)
	auth.POST("/register/complete", r.CompleteRegisterProfile)
	auth.POST("/login", r.Login)
}

func (r *Rest) Run() {
	addr := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")

	r.router.Run(fmt.Sprintf("%s:%s", addr, port))
}
