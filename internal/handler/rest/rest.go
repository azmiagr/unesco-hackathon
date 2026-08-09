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
	auth.POST("/login/admin/verify-otp", r.VerifyAdminLoginOtp)

	admin := baseUrl.Group("/admin")
	admin.Use(r.middleware.AuthenticateUser, r.middleware.OnlyAdmin())
	admin.GET("/users", r.ListAdminUsers)
	admin.GET("/users/:userID", r.GetAdminUserDetail)
	admin.GET("roles", r.GetAllRoles)
	admin.POST("/users", r.CreateUserByAdmin)
	admin.PATCH("/users/:userID", r.UpdateUserByAdmin)
	admin.PATCH("/users/:userID/access", r.UpdateAdminUserAccess)
	admin.DELETE("/users/:userID", r.HardDeleteUserByAdmin)

	cases := admin.Group("/cases")
	cases.GET("/lookups", r.GetCaseLookupsByAdmin)
	cases.GET("", r.ListCasesByAdmin)
	cases.GET("/:caseID/evidences", r.ListCaseEvidencesByAdmin)
	cases.GET("/:caseID", r.GetCaseDetailByAdmin)
	cases.POST("", r.CreateCaseByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/social-post", r.CreateSocialPostEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/article", r.CreateArticleEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/blog", r.CreateBlogEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/forum-thread", r.CreateForumThreadEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/chat-transcript", r.CreateChatTranscriptEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/public-announcement", r.CreatePublicAnnouncementEvidenceByAdmin)
}

func (r *Rest) Run() {
	addr := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")

	r.router.Run(fmt.Sprintf("%s:%s", addr, port))
}
