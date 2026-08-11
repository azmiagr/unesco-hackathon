package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) ListAvailableAvatars(c *gin.Context) {
	result, err := r.service.AvatarService.ListAvailableAvatars()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "avatars retrieved", result)
}
