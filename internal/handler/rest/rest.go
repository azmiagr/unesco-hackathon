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
	baseUrl.GET("/avatars", r.ListAvailableAvatars)

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
	admin.GET("/item-categories", r.ListItemCategoriesByAdmin)
	admin.POST("/users", r.CreateUserByAdmin)
	admin.PATCH("/users/:userID", r.UpdateUserByAdmin)
	admin.PATCH("/users/:userID/access", r.UpdateAdminUserAccess)
	admin.DELETE("/users/:userID", r.HardDeleteUserByAdmin)

	items := admin.Group("/items")
	items.GET("", r.ListItemsByAdmin)
	items.GET("/:itemID", r.GetItemDetailByAdmin)
	items.POST("", r.CreateItemByAdmin)
	items.PATCH("/:itemID", r.UpdateItemByAdmin)
	items.DELETE("/:itemID", r.DeleteItemByAdmin)

	cases := admin.Group("/cases")
	cases.GET("/lookups", r.GetCaseLookupsByAdmin)
	cases.GET("", r.ListCasesByAdmin)
	cases.GET("/:caseID/evidences", r.ListCaseEvidencesByAdmin)
	cases.GET("/:caseID/evidence-options", r.ListEvidenceOptionsByAdmin)
	cases.GET("/:caseID/questions", r.ListCaseQuestionsByAdmin)
	cases.GET("/:caseID/versions/:caseVersionID/evidences/:caseEvidenceID", r.GetCaseEvidenceDetailByAdmin)
	cases.GET("/:caseID/versions/:caseVersionID/questions/:caseQuestionID", r.GetCaseQuestionDetailByAdmin)
	cases.GET("/:caseID/versions/:caseVersionID/scoring-outcome-config", r.GetCaseScoringOutcomeConfigByAdmin)
	cases.PUT("/:caseID/versions/:caseVersionID/scoring-outcome-config", r.UpsertCaseScoringOutcomeConfigByAdmin)
	cases.GET("/:caseID/chatbot-config", r.GetCaseChatbotConfigByAdmin)
	cases.PUT("/:caseID/chatbot-config", r.UpsertCaseChatbotConfigByAdmin)
	cases.GET("/:caseID", r.GetCaseDetailByAdmin)
	cases.POST("", r.CreateCaseByAdmin)
	cases.PATCH("/:caseID", r.UpdateCaseByAdmin)
	cases.PATCH("/:caseID/publish", r.PublishCaseByAdmin)
	cases.DELETE("/:caseID", r.HardDeleteCaseByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/social-post", r.CreateSocialPostEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/article", r.CreateArticleEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/blog", r.CreateBlogEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/forum-thread", r.CreateForumThreadEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/chat-transcript", r.CreateChatTranscriptEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/evidences/public-announcement", r.CreatePublicAnnouncementEvidenceByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/questions/mcq", r.CreateMCQQuestionByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/questions/open-ended", r.CreateOpenEndedQuestionByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/questions/confidence-slider", r.CreateConfidenceSliderQuestionByAdmin)
	cases.POST("/:caseID/versions/:caseVersionID/questions/claim-classification", r.CreateClaimClassificationQuestionByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/questions/:caseQuestionID/mcq", r.UpdateMCQQuestionByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/questions/:caseQuestionID/open-ended", r.UpdateOpenEndedQuestionByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/questions/:caseQuestionID/confidence-slider", r.UpdateConfidenceSliderQuestionByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/questions/:caseQuestionID/claim-classification", r.UpdateClaimClassificationQuestionByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/evidences/:caseEvidenceID/social-post", r.UpdateSocialPostEvidenceByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/evidences/:caseEvidenceID/article", r.UpdateArticleEvidenceByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/evidences/:caseEvidenceID/blog", r.UpdateBlogEvidenceByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/evidences/:caseEvidenceID/forum-thread", r.UpdateForumThreadEvidenceByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/evidences/:caseEvidenceID/chat-transcript", r.UpdateChatTranscriptEvidenceByAdmin)
	cases.PATCH("/:caseID/versions/:caseVersionID/evidences/:caseEvidenceID/public-announcement", r.UpdatePublicAnnouncementEvidenceByAdmin)
	cases.DELETE("/:caseID/versions/:caseVersionID/evidences/:caseEvidenceID", r.DeleteCaseEvidenceByAdmin)
	cases.DELETE("/:caseID/versions/:caseVersionID/questions/:caseQuestionID", r.DeleteCaseQuestionByAdmin)
}

func (r *Rest) Run() {
	addr := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")

	r.router.Run(fmt.Sprintf("%s:%s", addr, port))
}
