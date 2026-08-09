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

func (r *Rest) GetCaseLookupsByAdmin(c *gin.Context) {
	result, err := r.service.CaseService.GetCaseLookups()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case lookups retrieved", result)
}
