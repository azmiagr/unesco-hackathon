package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) GetGameConfigByAdmin(c *gin.Context) {
	result, err := r.service.GameConfigService.GetGameConfigByAdmin()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game config retrieved", result)
}

func (r *Rest) UpsertGameConfigByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpsertGameConfigRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid upsert game config request: "+err.Error()))
		return
	}

	result, err := r.service.GameConfigService.UpsertGameConfigByAdmin(adminUser.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game config saved", result)
}

func (r *Rest) GetGameGeneralConfigByAdmin(c *gin.Context) {
	result, err := r.service.GameConfigService.GetGameGeneralConfigByAdmin()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game general config retrieved", result)
}

func (r *Rest) UpsertGameGeneralConfigByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpsertGameGeneralConfigRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid upsert game general config request: "+err.Error()))
		return
	}

	result, err := r.service.GameConfigService.UpsertGameGeneralConfigByAdmin(adminUser.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game general config saved", result)
}

func (r *Rest) GetGameAIConfigByAdmin(c *gin.Context) {
	result, err := r.service.GameConfigService.GetGameAIConfigByAdmin()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game ai config retrieved", result)
}

func (r *Rest) UpsertGameAIConfigByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpsertGameAIConfigRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid upsert game ai config request: "+err.Error()))
		return
	}

	result, err := r.service.GameConfigService.UpsertGameAIConfigByAdmin(adminUser.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game ai config saved", result)
}

func (r *Rest) ListGameLevelsByAdmin(c *gin.Context) {
	var req model.AdminListGameLevelsRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list game levels request"))
		return
	}

	result, err := r.service.GameConfigService.ListGameLevelsByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game levels retrieved", result)
}

func (r *Rest) GetGameLevelDetailByAdmin(c *gin.Context) {
	gameLevelID, err := helper.ParseUUIDParam(c, "gameLevelID", "invalid game level id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.GameConfigService.GetGameLevelDetailByAdmin(gameLevelID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game level detail retrieved", result)
}

func (r *Rest) CreateGameLevelByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateGameLevelRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create game level request: "+err.Error()))
		return
	}

	result, err := r.service.GameConfigService.CreateGameLevelByAdmin(adminUser.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "game level created", result)
}

func (r *Rest) UpdateGameLevelByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	gameLevelID, err := helper.ParseUUIDParam(c, "gameLevelID", "invalid game level id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateGameLevelRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update game level request: "+err.Error()))
		return
	}

	result, err := r.service.GameConfigService.UpdateGameLevelByAdmin(adminUser.UserID, gameLevelID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game level updated", result)
}

func (r *Rest) DeleteGameLevelByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	gameLevelID, err := helper.ParseUUIDParam(c, "gameLevelID", "invalid game level id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.GameConfigService.DeleteGameLevelByAdmin(adminUser.UserID, gameLevelID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "game level deleted", result)
}
