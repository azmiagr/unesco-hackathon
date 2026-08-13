package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) ListLeaderboard(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.ListLeaderboardRequest
	err = c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid leaderboard request"))
		return
	}

	result, err := r.service.LeaderboardService.ListLeaderboard(user.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "leaderboard retrieved", result)
}
