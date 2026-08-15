package rest

import (
	"errors"
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) ListTitlesByAdmin(c *gin.Context) {
	var req model.AdminListTitlesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list titles request"))
		return
	}
	result, err := r.service.TitleService.ListTitlesByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "titles retrieved", result)
}

func (r *Rest) GetTitleByAdmin(c *gin.Context) {
	titleID, err := helper.ParseUUIDParam(c, "titleID", "invalid title id")
	if err != nil {
		response.HandleError(c, err)
		return
	}
	result, err := r.service.TitleService.GetTitleByAdmin(titleID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "title retrieved", result)
}

func (r *Rest) CreateTitleByAdmin(c *gin.Context) {
	var req model.AdminCreateTitleRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create title request"))
		return
	}
	image, err := c.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid title image file"))
			return
		}
	} else {
		req.Image = image
	}
	result, err := r.service.TitleService.CreateTitleByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "title created", result)
}

func (r *Rest) UpdateTitleByAdmin(c *gin.Context) {
	titleID, err := helper.ParseUUIDParam(c, "titleID", "invalid title id")
	if err != nil {
		response.HandleError(c, err)
		return
	}
	var req model.AdminUpdateTitleRequest
	if err := c.ShouldBind(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update title request"))
		return
	}
	image, err := c.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid title image file"))
			return
		}
	} else {
		req.Image = image
	}
	result, err := r.service.TitleService.UpdateTitleByAdmin(titleID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "title updated", result)
}

func (r *Rest) DeleteTitleByAdmin(c *gin.Context) {
	titleID, err := helper.ParseUUIDParam(c, "titleID", "invalid title id")
	if err != nil {
		response.HandleError(c, err)
		return
	}
	result, err := r.service.TitleService.DeleteTitleByAdmin(titleID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "title deleted", result)
}
