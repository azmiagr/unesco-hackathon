package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) GetAllRoles(c *gin.Context) {
	result, err := r.service.RoleService.GetAllRoles()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get all roles", result)
}
