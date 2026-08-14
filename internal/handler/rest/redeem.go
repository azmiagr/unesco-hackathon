package rest

import (
	"errors"
	"net/http"

	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

func (r *Rest) ListRedeemItemsForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.UserListRedeemItemsRequest
	err = c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list redeem items request"))
		return
	}

	result, err := r.service.RedeemService.ListRedeemItemsForUser(user.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "redeem items retrieved", result)
}

func (r *Rest) PurchaseRedeemItemForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	redeemItemID, err := helper.ParseUUIDParam(c, "redeemItemID", "invalid redeem item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.RedeemService.PurchaseRedeemItemForUser(user.UserID, redeemItemID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "redeem item purchased", result)
}

func (r *Rest) ListRedeemTypesByAdmin(c *gin.Context) {
	var req model.AdminListRedeemTypesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list redeem types request"))
		return
	}

	result, err := r.service.RedeemService.ListRedeemTypesByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "redeem types retrieved", result)
}

func (r *Rest) ListRedeemItemsByAdmin(c *gin.Context) {
	var req model.AdminListRedeemItemsRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list redeem items request"))
		return
	}

	result, err := r.service.RedeemService.ListRedeemItemsByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "redeem items retrieved", result)
}

func (r *Rest) ListRedeemCodesByAdmin(c *gin.Context) {
	var req model.AdminListRedeemCodesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list redeem codes request"))
		return
	}

	result, err := r.service.RedeemService.ListRedeemCodesByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "redeem codes retrieved", result)
}

func (r *Rest) GetRedeemItemDetailByAdmin(c *gin.Context) {
	redeemItemID, err := helper.ParseUUIDParam(c, "redeemItemID", "invalid redeem item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.RedeemService.GetRedeemItemDetailByAdmin(redeemItemID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "redeem item detail retrieved", result)
}

func (r *Rest) CreateRedeemItemByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateRedeemItemRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create redeem item request: "+err.Error()))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid redeem item image file"))
			return
		}
	} else {
		req.Image = image
	}

	result, err := r.service.RedeemService.CreateRedeemItemByAdmin(adminUser.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "redeem item created", result)
}

func (r *Rest) CreateRedeemCodeManualByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateRedeemCodeManualRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create redeem code request"))
		return
	}

	result, err := r.service.RedeemService.CreateRedeemCodeManualByAdmin(adminUser.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "redeem code created", result)
}

func (r *Rest) CreateRedeemCodesCSVByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("csv file is required"))
		return
	}

	result, err := r.service.RedeemService.CreateRedeemCodesCSVByAdmin(adminUser.UserID, model.AdminCreateRedeemCodeCSVRequest{
		File: file,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "redeem codes uploaded", result)
}

func (r *Rest) UpdateRedeemItemByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	redeemItemID, err := helper.ParseUUIDParam(c, "redeemItemID", "invalid redeem item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateRedeemItemRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update redeem item request: "+err.Error()))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid redeem item image file"))
			return
		}
	} else {
		req.Image = image
	}

	result, err := r.service.RedeemService.UpdateRedeemItemByAdmin(adminUser.UserID, redeemItemID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "redeem item updated", result)
}

func (r *Rest) DeleteRedeemItemByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	redeemItemID, err := helper.ParseUUIDParam(c, "redeemItemID", "invalid redeem item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.RedeemService.DeleteRedeemItemByAdmin(adminUser.UserID, redeemItemID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "redeem item deleted", result)
}

func (r *Rest) DeleteRedeemCodeByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	redeemCodeID, err := helper.ParseUUIDParam(c, "redeemCodeID", "invalid redeem code id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.RedeemService.DeleteRedeemCodeByAdmin(adminUser.UserID, redeemCodeID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "redeem code deleted", result)
}
