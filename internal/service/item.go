package service

import (
	"errors"
	"math"
	"mime/multipart"
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultAdminItemLimit = 10
	maxAdminItemLimit     = 100
	maxItemImageSize      = 5 * 1024 * 1024
)

var allowedItemStatuses = map[string]bool{
	model.ItemStatusActive:   true,
	model.ItemStatusInactive: true,
	model.ItemStatusRetired:  true,
}

type IItemService interface {
	ListItemCategoriesByAdmin(req model.AdminListItemCategoriesRequest) (*model.AdminListItemCategoriesResponse, error)
	ListItemsByAdmin(req model.AdminListItemsRequest) (*model.AdminListItemsResponse, error)
	GetItemDetailByAdmin(itemID uuid.UUID) (*model.AdminGetItemDetailResponse, error)
	CreateItemByAdmin(adminUserID uuid.UUID, req model.AdminCreateItemRequest) (*model.AdminCreateItemResponse, error)
	UpdateItemByAdmin(adminUserID uuid.UUID, itemID uuid.UUID, req model.AdminUpdateItemRequest) (*model.AdminUpdateItemResponse, error)
	DeleteItemByAdmin(adminUserID uuid.UUID, itemID uuid.UUID) (*model.AdminDeleteItemResponse, error)
}

type ItemService struct {
	db               *gorm.DB
	itemRepo         repository.IItemRepository
	itemCategoryRepo repository.IItemCategoryRepository
	storage          supabase.Interface
}

func NewItemService(
	itemRepo repository.IItemRepository,
	itemCategoryRepo repository.IItemCategoryRepository,
	storage supabase.Interface,
) IItemService {
	return &ItemService{
		db:               mariadb.Connection,
		itemRepo:         itemRepo,
		itemCategoryRepo: itemCategoryRepo,
		storage:          storage,
	}
}

func (s *ItemService) ListItemCategoriesByAdmin(req model.AdminListItemCategoriesRequest) (*model.AdminListItemCategoriesResponse, error) {
	categories, err := s.itemCategoryRepo.ListItemCategories(s.db, model.ListItemCategoriesParam{
		Search: strings.TrimSpace(req.Search),
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list item categories")
	}

	responses := make([]model.AdminItemCategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, mapAdminItemCategoryResponse(category))
	}

	return &model.AdminListItemCategoriesResponse{
		Categories: responses,
	}, nil
}

func (s *ItemService) ListItemsByAdmin(req model.AdminListItemsRequest) (*model.AdminListItemsResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultAdminItemLimit
	}
	if limit > maxAdminItemLimit {
		limit = maxAdminItemLimit
	}

	itemCategoryID := uuid.Nil
	if strings.TrimSpace(req.ItemCategoryID) != "" {
		parsedID, err := uuid.Parse(strings.TrimSpace(req.ItemCategoryID))
		if err != nil {
			return nil, appErrors.BadRequest("invalid item category id")
		}
		itemCategoryID = parsedID
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "" && !allowedItemStatuses[status] {
		return nil, appErrors.BadRequest("invalid status filter")
	}

	items, total, err := s.itemRepo.ListItems(s.db, model.ListItemsParam{
		Search:         strings.TrimSpace(req.Search),
		ItemCategoryID: itemCategoryID,
		CategoryCode:   strings.ToLower(strings.TrimSpace(req.CategoryCode)),
		Status:         status,
		IsVisible:      req.IsVisible,
		IsFeatured:     req.IsFeatured,
		Limit:          limit,
		Offset:         (page - 1) * limit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list items")
	}

	responses := make([]model.AdminItemResponse, 0, len(items))
	for _, item := range items {
		itemResponse, err := s.mapAdminItemResponse(s.db, item)
		if err != nil {
			return nil, err
		}
		responses = append(responses, itemResponse)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.AdminListItemsResponse{
		Items: responses,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *ItemService) GetItemDetailByAdmin(itemID uuid.UUID) (*model.AdminGetItemDetailResponse, error) {
	if itemID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid item id")
	}

	item, err := s.itemRepo.GetItem(s.db, model.GetItemParam{ItemID: itemID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("item not found")
		}
		return nil, appErrors.InternalServer("failed to get item detail")
	}

	itemResponse, err := s.mapAdminItemResponse(s.db, *item)
	if err != nil {
		return nil, err
	}

	return &model.AdminGetItemDetailResponse{
		Item: itemResponse,
	}, nil
}

func (s *ItemService) CreateItemByAdmin(
	adminUserID uuid.UUID,
	req model.AdminCreateItemRequest,
) (*model.AdminCreateItemResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	name, description, status, err := validateAdminItemPayload(req.Name, req.Description, req.PriceCoin, req.Status)
	if err != nil {
		return nil, err
	}
	itemCategoryID, err := parseItemCategoryID(req.ItemCategoryID)
	if err != nil {
		return nil, err
	}
	if req.Image == nil {
		return nil, appErrors.BadRequest("item image is required")
	}

	imageURL, err := s.uploadItemImage(req.Image)
	if err != nil {
		return nil, err
	}

	shouldDeleteImage := true
	defer func() {
		if shouldDeleteImage && imageURL != "" {
			_ = supabase.DeleteFileIfPresent(s.storage, imageURL)
		}
	}()

	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}

	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	_, err = s.itemCategoryRepo.GetItemCategory(tx, model.GetItemCategoryParam{
		ItemCategoryID: itemCategoryID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("item category not found")
		}
		return nil, appErrors.InternalServer("failed to get item category")
	}

	item := &entity.Item{
		ItemID:         uuid.New(),
		ItemCategoryID: itemCategoryID,
		Name:           name,
		Description:    description,
		PriceCoin:      req.PriceCoin,
		ImageURL:       imageURL,
		IsVisible:      isVisible,
		IsFeatured:     isFeatured,
		Status:         status,
	}

	err = s.itemRepo.CreateItem(tx, item)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create item")
	}

	savedItem, err := s.itemRepo.GetItem(tx, model.GetItemParam{ItemID: item.ItemID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get created item")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteImage = false

	itemResponse, err := s.mapAdminItemResponse(s.db, *savedItem)
	if err != nil {
		return nil, err
	}

	return &model.AdminCreateItemResponse{
		Item: itemResponse,
	}, nil
}

func (s *ItemService) UpdateItemByAdmin(
	adminUserID uuid.UUID,
	itemID uuid.UUID,
	req model.AdminUpdateItemRequest,
) (*model.AdminUpdateItemResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if itemID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid item id")
	}

	name, description, status, err := validateAdminItemPayload(req.Name, req.Description, req.PriceCoin, req.Status)
	if err != nil {
		return nil, err
	}
	itemCategoryID, err := parseItemCategoryID(req.ItemCategoryID)
	if err != nil {
		return nil, err
	}

	imageURL := ""
	if req.Image != nil {
		imageURL, err = s.uploadItemImage(req.Image)
		if err != nil {
			return nil, err
		}
	}

	shouldDeleteNewImage := imageURL != ""
	defer func() {
		if shouldDeleteNewImage && imageURL != "" {
			_ = supabase.DeleteFileIfPresent(s.storage, imageURL)
		}
	}()

	tx := s.db.Begin()
	defer tx.Rollback()

	_, err = s.itemCategoryRepo.GetItemCategory(tx, model.GetItemCategoryParam{
		ItemCategoryID: itemCategoryID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("item category not found")
		}
		return nil, appErrors.InternalServer("failed to get item category")
	}

	item, err := s.itemRepo.GetItemForUpdate(tx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("item not found")
		}
		return nil, appErrors.InternalServer("failed to get item")
	}

	oldImageURL := item.ImageURL
	item.ItemCategoryID = itemCategoryID
	item.Name = name
	item.Description = description
	item.PriceCoin = req.PriceCoin
	item.Status = status
	if req.IsVisible != nil {
		item.IsVisible = *req.IsVisible
	}
	if req.IsFeatured != nil {
		item.IsFeatured = *req.IsFeatured
	}
	if imageURL != "" {
		item.ImageURL = imageURL
	}

	err = s.itemRepo.UpdateItem(tx, item)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update item")
	}

	savedItem, err := s.itemRepo.GetItem(tx, model.GetItemParam{ItemID: item.ItemID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get updated item")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteNewImage = false
	if imageURL != "" && oldImageURL != "" && oldImageURL != imageURL {
		_ = supabase.DeleteFileIfPresent(s.storage, oldImageURL)
	}

	itemResponse, err := s.mapAdminItemResponse(s.db, *savedItem)
	if err != nil {
		return nil, err
	}

	return &model.AdminUpdateItemResponse{
		Item: itemResponse,
	}, nil
}

func (s *ItemService) DeleteItemByAdmin(
	adminUserID uuid.UUID,
	itemID uuid.UUID,
) (*model.AdminDeleteItemResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if itemID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid item id")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	item, err := s.itemRepo.GetItemForUpdate(tx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("item not found")
		}
		return nil, appErrors.InternalServer("failed to get item")
	}

	err = s.itemRepo.DeleteItem(tx, item.ItemID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete item")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminDeleteItemResponse{
		ItemID: item.ItemID,
	}, nil
}

func validateAdminItemPayload(nameRaw string, descriptionRaw string, priceCoin int, statusRaw string) (string, string, string, error) {
	name, err := helper.RequireTrimmedString(nameRaw, "name is required")
	if err != nil {
		return "", "", "", err
	}
	if len(name) > 150 {
		return "", "", "", appErrors.BadRequest("name is too long")
	}

	description, err := helper.RequireTrimmedString(descriptionRaw, "description is required")
	if err != nil {
		return "", "", "", err
	}

	if priceCoin < 1 {
		return "", "", "", appErrors.BadRequest("price coin must be greater than 0")
	}

	status := strings.ToLower(strings.TrimSpace(statusRaw))
	if status == "" {
		status = model.ItemStatusActive
	}
	if !allowedItemStatuses[status] {
		return "", "", "", appErrors.BadRequest("invalid item status")
	}

	return name, description, status, nil
}

func parseItemCategoryID(raw string) (uuid.UUID, error) {
	itemCategoryID, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, appErrors.BadRequest("invalid item category id")
	}

	return itemCategoryID, nil
}

func (s *ItemService) uploadItemImage(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", nil
	}

	url, err := supabase.UploadOptionalImage(
		s.storage,
		file,
		maxItemImageSize,
		"item image size exceeds 5MB limit",
	)
	if err != nil {
		return "", appErrors.BadRequest("failed to upload item image")
	}

	return url, nil
}

func (s *ItemService) mapAdminItemResponse(tx *gorm.DB, item entity.Item) (model.AdminItemResponse, error) {
	category, err := s.itemCategoryRepo.GetItemCategory(tx, model.GetItemCategoryParam{
		ItemCategoryID: item.ItemCategoryID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.AdminItemResponse{}, appErrors.InternalServer("item category not found")
		}
		return model.AdminItemResponse{}, appErrors.InternalServer("failed to get item category")
	}

	return model.AdminItemResponse{
		ItemID:         item.ItemID,
		ItemCategoryID: item.ItemCategoryID,
		Category:       mapAdminItemCategoryResponse(*category),
		Name:           item.Name,
		Description:    item.Description,
		PriceCoin:      item.PriceCoin,
		ImageURL:       item.ImageURL,
		IsVisible:      item.IsVisible,
		IsFeatured:     item.IsFeatured,
		Status:         item.Status,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}, nil
}

func mapAdminItemCategoryResponse(category entity.ItemCategory) model.AdminItemCategoryResponse {
	return model.AdminItemCategoryResponse{
		ItemCategoryID: category.ItemCategoryID,
		Code:           category.Code,
		Name:           category.Name,
		Description:    category.Description,
		CreatedAt:      category.CreatedAt,
		UpdatedAt:      category.UpdatedAt,
	}
}
