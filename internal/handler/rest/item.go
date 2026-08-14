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

func (r *Rest) ListItemCategoriesByAdmin(c *gin.Context) {
	var req model.AdminListItemCategoriesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list item categories request"))
		return
	}

	result, err := r.service.ItemService.ListItemCategoriesByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "item categories retrieved", result)
}

func (r *Rest) ListItemCategoriesForUser(c *gin.Context) {
	var req model.UserListItemCategoriesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list item categories request"))
		return
	}

	result, err := r.service.ItemService.ListItemCategoriesForUser(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "item categories retrieved", result)
}

func (r *Rest) ListItemsByAdmin(c *gin.Context) {
	var req model.AdminListItemsRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list items request"))
		return
	}

	result, err := r.service.ItemService.ListItemsByAdmin(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "items retrieved", result)
}

func (r *Rest) ListShopItemsForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.UserListShopItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid list shop items request"))
		return
	}

	result, err := r.service.ItemService.ListShopItemsForUser(user.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "shop items retrieved", result)
}

func (r *Rest) GetShopItemDetailForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	itemID, err := helper.ParseUUIDParam(c, "itemID", "invalid item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.ItemService.GetShopItemDetailForUser(user.UserID, itemID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "shop item detail retrieved", result)
}

func (r *Rest) PurchaseShopItemForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	itemID, err := helper.ParseUUIDParam(c, "itemID", "invalid item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.ItemService.PurchaseShopItemForUser(user.UserID, itemID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "shop item purchased", result)
}

func (r *Rest) EquipShopItemForUser(c *gin.Context) {
	user, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	itemID, err := helper.ParseUUIDParam(c, "itemID", "invalid item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.ItemService.EquipShopItemForUser(user.UserID, itemID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "shop item equipped", result)
}

func (r *Rest) GetItemDetailByAdmin(c *gin.Context) {
	itemID, err := helper.ParseUUIDParam(c, "itemID", "invalid item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.ItemService.GetItemDetailByAdmin(itemID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "item detail retrieved", result)
}

func (r *Rest) CreateItemByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminCreateItemRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid create item request: "+err.Error()))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid item image file"))
			return
		}
	} else {
		req.Image = image
	}

	result, err := r.service.ItemService.CreateItemByAdmin(adminUser.UserID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "item created", result)
}

func (r *Rest) UpdateItemByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	itemID, err := helper.ParseUUIDParam(c, "itemID", "invalid item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req model.AdminUpdateItemRequest
	err = c.ShouldBind(&req)
	if err != nil {
		response.HandleError(c, appErrors.BadRequest("invalid update item request: "+err.Error()))
		return
	}

	image, err := c.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			response.HandleError(c, appErrors.BadRequest("invalid item image file"))
			return
		}
	} else {
		req.Image = image
	}

	result, err := r.service.ItemService.UpdateItemByAdmin(adminUser.UserID, itemID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "item updated", result)
}

func (r *Rest) DeleteItemByAdmin(c *gin.Context) {
	adminUser, err := helper.GetAuthenticatedUser(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	itemID, err := helper.ParseUUIDParam(c, "itemID", "invalid item id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := r.service.ItemService.DeleteItemByAdmin(adminUser.UserID, itemID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "item deleted", result)
}
