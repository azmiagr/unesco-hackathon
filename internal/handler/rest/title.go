package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) ListTitlesForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	result, err := r.service.TitleService.ListTitlesForUser(user.UserID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "titles retrieved", result)
}

func (r *Rest) EquipTitleForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	titleID, err := helper.ParseUUIDParam(c, "titleID", "invalid title id")
	if err != nil {
		response.HandleError(c, err)
		return
	}
	result, err := r.service.TitleService.EquipTitleForUser(user.UserID, titleID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "title equipped", result)
}
