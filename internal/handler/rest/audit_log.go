package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) ListAuditLogsByAdmin(c *gin.Context) {
	var req model.AdminListAuditLogsRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list audit logs request"))
		return
	}

	result, err := r.service.AuditLogService.ListAuditLogsByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "audit logs retrieved", result)
}
