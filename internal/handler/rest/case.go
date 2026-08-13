package rest

import (
	stdErrors "errors"
	"net/http"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) CreateCaseByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateCaseRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create case request"))
		return
	}

	thumbnail, err := c.FormFile("thumbnail")
	if err != nil {
		if !stdErrors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid thumbnail file"))
			return
		}
	} else {
		req.Thumbnail = thumbnail
	}

	result, err := r.service.CaseService.CreateCaseByAdmin(adminUser.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "case created", result)
}

func (r *Rest) UpdateCaseByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateCaseRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update case request"))
		return
	}

	thumbnail, err := c.FormFile("thumbnail")
	if err != nil {
		if !stdErrors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid thumbnail file"))
			return
		}
	} else {
		req.Thumbnail = thumbnail
	}

	result, err := r.service.CaseService.UpdateCaseByAdmin(adminUser.UserID, caseID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case updated", result)
}

func (r *Rest) PublishCaseByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.PublishCaseByAdmin(adminUser.UserID, caseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case published", result)
}

func (r *Rest) HardDeleteCaseByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.HardDeleteCaseByAdmin(adminUser.UserID, caseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case deleted", result)
}

func (r *Rest) ListCasesByAdmin(c *gin.Context) {
	var req model.AdminListCasesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list cases request"))
		return
	}

	result, err := r.service.CaseService.ListCasesByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "cases retrieved", result)
}

func (r *Rest) GetCaseDetailByAdmin(c *gin.Context) {
	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.GetCaseDetailByAdmin(caseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case detail retrieved", result)
}

func (r *Rest) ListCaseEvidencesByAdmin(c *gin.Context) {
	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.ListCaseEvidencesByAdmin(caseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case evidences retrieved", result)
}

func (r *Rest) ListEvidenceOptionsByAdmin(c *gin.Context) {
	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.ListEvidenceOptionsByAdmin(caseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "evidence options retrieved", result)
}

func (r *Rest) GetCaseEvidenceDetailByAdmin(c *gin.Context) {
	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseEvidenceID, err := helper.ParseUUIDParam(c, "caseEvidenceID", "invalid case evidence id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.GetCaseEvidenceDetailByAdmin(caseID, caseVersionID, caseEvidenceID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case evidence detail retrieved", result)
}

func (r *Rest) ListCaseQuestionsByAdmin(c *gin.Context) {
	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.ListCaseQuestionsByAdmin(caseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case questions retrieved", result)
}

func (r *Rest) GetCaseQuestionDetailByAdmin(c *gin.Context) {
	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseQuestionID, err := helper.ParseUUIDParam(c, "caseQuestionID", "invalid case question id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.GetCaseQuestionDetailByAdmin(caseID, caseVersionID, caseQuestionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case question detail retrieved", result)
}

func (r *Rest) GetCaseChatbotConfigByAdmin(c *gin.Context) {
	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.GetCaseChatbotConfigByAdmin(caseID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case chatbot config retrieved", result)
}

func (r *Rest) UpsertCaseChatbotConfigByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpsertCaseChatbotConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid upsert case chatbot config request"))
		return
	}

	result, err := r.service.CaseService.UpsertCaseChatbotConfigByAdmin(adminUser.UserID, caseID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case chatbot config saved", result)
}

func (r *Rest) GetCaseScoringOutcomeConfigByAdmin(c *gin.Context) {
	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.GetCaseScoringOutcomeConfigByAdmin(caseID, caseVersionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case scoring outcome config retrieved", result)
}

func (r *Rest) UpsertCaseScoringOutcomeConfigByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpsertCaseScoringOutcomeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid upsert case scoring outcome config request"))
		return
	}

	result, err := r.service.CaseService.UpsertCaseScoringOutcomeConfigByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case scoring outcome config saved", result)
}

func (r *Rest) GetCaseLookupsByAdmin(c *gin.Context) {
	result, err := r.service.CaseService.GetCaseLookups()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case lookups retrieved", result)
}

func (r *Rest) CreateSocialPostEvidenceByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateSocialPostEvidenceRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create social post evidence request"))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if !stdErrors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid evidence image file"))
			return
		}
	} else {
		req.Image = image
	}

	result, err := r.service.CaseService.CreateSocialPostEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "social post evidence created", result)
}

func (r *Rest) CreateArticleEvidenceByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateArticleEvidenceRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create article evidence request"))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if !stdErrors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid evidence image file"))
			return
		}
	} else {
		req.Image = image
	}

	result, err := r.service.CaseService.CreateArticleEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "article evidence created", result)
}

func (r *Rest) CreateBlogEvidenceByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateBlogEvidenceRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create blog evidence request"))
		return
	}

	result, err := r.service.CaseService.CreateBlogEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "blog evidence created", result)
}

func (r *Rest) CreateForumThreadEvidenceByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateForumThreadEvidenceRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create forum thread evidence request"))
		return
	}

	result, err := r.service.CaseService.CreateForumThreadEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "forum thread evidence created", result)
}

func (r *Rest) CreateChatTranscriptEvidenceByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateChatTranscriptEvidenceRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create chat transcript evidence request"))
		return
	}

	result, err := r.service.CaseService.CreateChatTranscriptEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "chat transcript evidence created", result)
}

func (r *Rest) CreatePublicAnnouncementEvidenceByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreatePublicAnnouncementEvidenceRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create public announcement evidence request"))
		return
	}

	result, err := r.service.CaseService.CreatePublicAnnouncementEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "public announcement evidence created", result)
}

func (r *Rest) CreateMCQQuestionByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateMCQQuestionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create mcq question request"))
		return
	}

	result, err := r.service.CaseService.CreateMCQQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "mcq question created", result)
}

func (r *Rest) CreateOpenEndedQuestionByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateOpenEndedQuestionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create open ended question request"))
		return
	}

	result, err := r.service.CaseService.CreateOpenEndedQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "open ended question created", result)
}

func (r *Rest) CreateConfidenceSliderQuestionByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateConfidenceSliderQuestionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create confidence slider question request"))
		return
	}

	result, err := r.service.CaseService.CreateConfidenceSliderQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "confidence slider question created", result)
}

func (r *Rest) UpdateSocialPostEvidenceByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseEvidenceID, err := parseAdminEvidenceRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateSocialPostEvidenceRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update social post evidence request"))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if !stdErrors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid evidence image file"))
			return
		}
	} else {
		req.Image = image
	}

	result, err := r.service.CaseService.UpdateSocialPostEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, caseEvidenceID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "social post evidence updated", result)
}

func (r *Rest) UpdateArticleEvidenceByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseEvidenceID, err := parseAdminEvidenceRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateArticleEvidenceRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update article evidence request"))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if !stdErrors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid evidence image file"))
			return
		}
	} else {
		req.Image = image
	}

	result, err := r.service.CaseService.UpdateArticleEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, caseEvidenceID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "article evidence updated", result)
}

func (r *Rest) UpdateBlogEvidenceByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseEvidenceID, err := parseAdminEvidenceRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateBlogEvidenceRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update blog evidence request"))
		return
	}

	result, err := r.service.CaseService.UpdateBlogEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, caseEvidenceID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "blog evidence updated", result)
}

func (r *Rest) UpdateForumThreadEvidenceByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseEvidenceID, err := parseAdminEvidenceRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateForumThreadEvidenceRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update forum thread evidence request"))
		return
	}

	result, err := r.service.CaseService.UpdateForumThreadEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, caseEvidenceID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "forum thread evidence updated", result)
}

func (r *Rest) UpdateChatTranscriptEvidenceByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseEvidenceID, err := parseAdminEvidenceRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateChatTranscriptEvidenceRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update chat transcript evidence request"))
		return
	}

	result, err := r.service.CaseService.UpdateChatTranscriptEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, caseEvidenceID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "chat transcript evidence updated", result)
}

func (r *Rest) UpdatePublicAnnouncementEvidenceByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseEvidenceID, err := parseAdminEvidenceRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdatePublicAnnouncementEvidenceRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update public announcement evidence request"))
		return
	}

	result, err := r.service.CaseService.UpdatePublicAnnouncementEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, caseEvidenceID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "public announcement evidence updated", result)
}

func (r *Rest) DeleteCaseEvidenceByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseEvidenceID, err := parseAdminEvidenceRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.DeleteCaseEvidenceByAdmin(adminUser.UserID, caseID, caseVersionID, caseEvidenceID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case evidence deleted", result)
}

func (r *Rest) UpdateMCQQuestionByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseQuestionID, err := parseAdminQuestionRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateMCQQuestionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update mcq question request"))
		return
	}

	result, err := r.service.CaseService.UpdateMCQQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, caseQuestionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "mcq question updated", result)
}

func (r *Rest) UpdateOpenEndedQuestionByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseQuestionID, err := parseAdminQuestionRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateOpenEndedQuestionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update open ended question request"))
		return
	}

	result, err := r.service.CaseService.UpdateOpenEndedQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, caseQuestionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "open ended question updated", result)
}

func (r *Rest) UpdateConfidenceSliderQuestionByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseQuestionID, err := parseAdminQuestionRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateConfidenceSliderQuestionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update confidence slider question request"))
		return
	}

	result, err := r.service.CaseService.UpdateConfidenceSliderQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, caseQuestionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "confidence slider question updated", result)
}

func (r *Rest) UpdateClaimClassificationQuestionByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseQuestionID, err := parseAdminQuestionRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateClaimClassificationQuestionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update claim classification question request"))
		return
	}

	result, err := r.service.CaseService.UpdateClaimClassificationQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, caseQuestionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "claim classification question updated", result)
}

func (r *Rest) DeleteCaseQuestionByAdmin(c *gin.Context) {
	adminUser, caseID, caseVersionID, caseQuestionID, err := parseAdminQuestionRoute(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.CaseService.DeleteCaseQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, caseQuestionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case question deleted", result)
}

func parseAdminEvidenceRoute(c *gin.Context) (*entity.User, uuid.UUID, uuid.UUID, uuid.UUID, error) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	caseEvidenceID, err := helper.ParseUUIDParam(c, "caseEvidenceID", "invalid case evidence id")
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	return adminUser, caseID, caseVersionID, caseEvidenceID, nil
}

func parseAdminQuestionRoute(c *gin.Context) (*entity.User, uuid.UUID, uuid.UUID, uuid.UUID, error) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	caseQuestionID, err := helper.ParseUUIDParam(c, "caseQuestionID", "invalid case question id")
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	return adminUser, caseID, caseVersionID, caseQuestionID, nil
}

func (r *Rest) CreateClaimClassificationQuestionByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseVersionID, err := helper.ParseUUIDParam(c, "caseVersionID", "invalid case version id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateClaimClassificationQuestionRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create claim classification question request"))
		return
	}

	result, err := r.service.CaseService.CreateClaimClassificationQuestionByAdmin(adminUser.UserID, caseID, caseVersionID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "claim classification question created", result)
}

func (r *Rest) ListCasesForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.ListUserCasesRequest
	err = c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list cases request"))
		return
	}

	result, err := r.service.CaseService.ListCasesForUser(user.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "cases retrieved", result)
}
