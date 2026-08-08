package rest

import (
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

const registerSessionHeader = "X-Session-Token"

func (r *Rest) StartRegister(c *gin.Context) {
	var req model.StartRegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, errors.BadRequest("invalid register request"))
		return
	}

	result, err := r.service.AuthService.StartRegister(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Header(registerSessionHeader, result.SessionToken)
	response.Success(c, http.StatusCreated, "verification code sent", result.State)
}

func (r *Rest) VerifyRegisterOtp(c *gin.Context) {
	sessionToken := c.GetHeader(registerSessionHeader)
	if sessionToken == "" {
		response.HandleError(c, errors.Unauthorized("missing registration session token"))
		return
	}

	var req model.VerifyRegisterOtpRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, errors.BadRequest("invalid otp verification request"))
		return
	}

	result, err := r.service.AuthService.VerifyRegisterOtp(sessionToken, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Header(registerSessionHeader, result.SessionToken)
	response.Success(c, http.StatusOK, "email verified", result.State)
}

func (r *Rest) SelectRegisterAvatar(c *gin.Context) {
	sessionToken := c.GetHeader(registerSessionHeader)
	if sessionToken == "" {
		response.HandleError(c, errors.Unauthorized("missing registration session token"))
		return
	}

	var req model.SelectRegisterAvatarRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, errors.BadRequest("invalid avatar selection request"))
		return
	}

	result, err := r.service.AuthService.SelectRegisterAvatar(sessionToken, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Header(registerSessionHeader, result.SessionToken)
	response.Success(c, http.StatusOK, "avatar selected", result.State)
}

func (r *Rest) CompleteRegisterProfile(c *gin.Context) {
	sessionToken := c.GetHeader(registerSessionHeader)
	if sessionToken == "" {
		response.HandleError(c, errors.Unauthorized("missing registration session token"))
		return
	}

	var req model.CompleteRegisterProfileRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, errors.BadRequest("invalid profile completion request"))
		return
	}

	result, err := r.service.AuthService.CompleteRegisterProfile(sessionToken, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "registration completed", result)
}

func (r *Rest) GetRegisterSession(c *gin.Context) {
	sessionToken := c.GetHeader(registerSessionHeader)
	if sessionToken == "" {
		response.HandleError(c, errors.Unauthorized("missing registration session token"))
		return
	}

	result, err := r.service.AuthService.GetRegisterSession(sessionToken)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "registration session retrieved", result)
}

func (r *Rest) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, errors.BadRequest("invalid login request"))
		return
	}

	result, err := r.service.AuthService.Login(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if result.SessionToken != "" {
		c.Header(registerSessionHeader, result.SessionToken)
	}

	response.Success(c, http.StatusOK, "login success", result)
}

func (r *Rest) VerifyAdminLoginOtp(c *gin.Context) {
	sessionToken := c.GetHeader(registerSessionHeader)
	if sessionToken == "" {
		response.HandleError(c, errors.Unauthorized("missing admin login session token"))
		return
	}

	var req model.VerifyAdminLoginOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, errors.BadRequest("invalid admin login otp request"))
		return
	}

	result, err := r.service.AuthService.VerifyAdminLoginOtp(sessionToken, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "admin login verified", result)
}
