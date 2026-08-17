package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) GetUserProfile(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.ProfileService.GetUserProfile(user.UserID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "user profile retrieved", result)
}

func (r *Rest) UpdateOwnNickname(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.UpdateOwnNicknameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update nickname request"))
		return
	}

	result, err := r.service.UserService.UpdateOwnNickname(user.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "nickname updated", result)
}
