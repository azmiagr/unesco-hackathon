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
