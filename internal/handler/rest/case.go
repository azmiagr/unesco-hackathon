package rest

import (
	stdErrors "errors"
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
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
