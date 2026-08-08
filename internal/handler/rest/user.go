package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) ListAdminUsers(c *gin.Context) {
	var req model.AdminListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleError(c, errors.BadRequest("invalid list users request"))
		return
	}

	result, err := r.service.UserService.ListUsers(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "users retrieved", result)
}

func (r *Rest) GetAdminUserDetail(c *gin.Context) {
	userID, err := helper.ParseUserIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.UserService.GetUserDetail(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "user detail retrieved", result)
}

func (r *Rest) UpdateAdminUserAccess(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	targetUserID, err := helper.ParseUserIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateUserAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, errors.BadRequest("invalid update user access request"))
		return
	}

	result, err := r.service.UserService.UpdateUserAccess(adminUser.UserID, targetUserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "user access updated", result)
}

func (r *Rest) HardDeleteAdminUser(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	targetUserID, err := helper.ParseUserIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.UserService.HardDeleteUser(adminUser.UserID, targetUserID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "user deleted", result)
}
