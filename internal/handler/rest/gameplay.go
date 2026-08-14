package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) StartCaseSessionForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	caseID, err := helper.ParseUUIDParam(c, "caseID", "invalid case id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.StartCaseSessionRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.HandleError(c, appErrors.BadRequest("invalid start session request"))
			return
		}
	}

	result, err := r.service.GameplayService.StartCaseSessionForUser(user.UserID, caseID, req, c.GetHeader("Idempotency-Key"))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "case session started", result)
}

func (r *Rest) GetGameplayStateForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	sessionID, err := helper.ParseUUIDParam(c, "sessionID", "invalid session id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.GameplayService.GetGameplayStateForUser(user.UserID, sessionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "gameplay retrieved", result)
}

func (r *Rest) OpenEvidenceForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	sessionID, err := helper.ParseUUIDParam(c, "sessionID", "invalid session id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	evidenceID, err := helper.ParseUUIDParam(c, "caseEvidenceID", "invalid case evidence id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.OpenCaseSessionEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid open evidence request"))
		return
	}

	result, err := r.service.GameplayService.OpenEvidenceForUser(
		user.UserID,
		sessionID,
		evidenceID,
		req,
		c.GetHeader("Idempotency-Key"),
		c.Request.Method,
		c.FullPath(),
	)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "evidence opened", result)
}

func (r *Rest) SaveAnswersForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	sessionID, err := helper.ParseUUIDParam(c, "sessionID", "invalid session id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.SaveCaseSessionAnswersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid save answers request"))
		return
	}

	result, err := r.service.GameplayService.SaveAnswersForUser(
		user.UserID,
		sessionID,
		req,
		c.GetHeader("Idempotency-Key"),
		c.Request.Method,
		c.FullPath(),
	)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "answers saved", result)
}

func (r *Rest) SubmitCaseSessionForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	sessionID, err := helper.ParseUUIDParam(c, "sessionID", "invalid session id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.SubmitCaseSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid submit session request"))
		return
	}

	result, err := r.service.GameplayService.SubmitCaseSessionForUser(
		user.UserID,
		sessionID,
		req,
		c.GetHeader("Idempotency-Key"),
		c.Request.Method,
		c.FullPath(),
	)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "case session submitted", result)
}
