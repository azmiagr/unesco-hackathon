package service

import (
	"errors"
	"fmt"
	"math"
	"mime/multipart"
	"strings"
	"time"

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
	defaultUserShopLimit  = 10
	maxUserShopLimit      = 100
	maxItemImageSize      = 5 * 1024 * 1024
)

var allowedItemStatuses = map[string]bool{
	model.ItemStatusActive:   true,
	model.ItemStatusInactive: true,
	model.ItemStatusRetired:  true,
}

type IItemService interface {
	ListItemCategoriesByAdmin(req model.AdminListItemCategoriesRequest) (*model.AdminListItemCategoriesResponse, error)
	ListItemCategoriesForUser(req model.UserListItemCategoriesRequest) (*model.UserListItemCategoriesResponse, error)
	ListItemsByAdmin(req model.AdminListItemsRequest) (*model.AdminListItemsResponse, error)
	ListShopItemsForUser(userID uuid.UUID, req model.UserListShopItemsRequest) (*model.UserListShopItemsResponse, error)
	GetItemDetailByAdmin(itemID uuid.UUID) (*model.AdminGetItemDetailResponse, error)
	CreateItemByAdmin(adminUserID uuid.UUID, req model.AdminCreateItemRequest) (*model.AdminCreateItemResponse, error)
	PurchaseShopItemForUser(userID uuid.UUID, itemID uuid.UUID) (*model.UserPurchaseShopItemResponse, error)
	EquipShopItemForUser(userID uuid.UUID, itemID uuid.UUID) (*model.UserEquipShopItemResponse, error)
	UpdateItemByAdmin(adminUserID uuid.UUID, itemID uuid.UUID, req model.AdminUpdateItemRequest) (*model.AdminUpdateItemResponse, error)
	DeleteItemByAdmin(adminUserID uuid.UUID, itemID uuid.UUID) (*model.AdminDeleteItemResponse, error)
}

type ItemService struct {
	db               *gorm.DB
	itemRepo         repository.IItemRepository
	itemCategoryRepo repository.IItemCategoryRepository
	userItemRepo     repository.IUserItemRepository
	userProfileRepo  repository.IUserProfileRepository
	avatarRepo       repository.IAvatarRepository
	userRepo         repository.IUserRepository
	auditLogRepo     repository.IAuditLogRepository
	storage          supabase.Interface
}

func NewItemService(
	itemRepo repository.IItemRepository,
	itemCategoryRepo repository.IItemCategoryRepository,
	userItemRepo repository.IUserItemRepository,
	userProfileRepo repository.IUserProfileRepository,
	avatarRepo repository.IAvatarRepository,
	userRepo repository.IUserRepository,
	auditLogRepo repository.IAuditLogRepository,
	storage supabase.Interface,
) IItemService {
	return &ItemService{
		db:               mariadb.Connection,
		itemRepo:         itemRepo,
		itemCategoryRepo: itemCategoryRepo,
		userItemRepo:     userItemRepo,
		userProfileRepo:  userProfileRepo,
		avatarRepo:       avatarRepo,
		userRepo:         userRepo,
		auditLogRepo:     auditLogRepo,
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

func (s *ItemService) ListItemCategoriesForUser(req model.UserListItemCategoriesRequest) (*model.UserListItemCategoriesResponse, error) {
	categories, err := s.itemCategoryRepo.ListItemCategories(s.db, model.ListItemCategoriesParam{
		Search: strings.TrimSpace(req.Search),
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list item categories")
	}

	responses := make([]model.UserItemCategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, mapUserItemCategoryResponse(category))
	}

	return &model.UserListItemCategoriesResponse{
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

func (s *ItemService) ListShopItemsForUser(userID uuid.UUID, req model.UserListShopItemsRequest) (*model.UserListShopItemsResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultUserShopLimit
	}
	if limit > maxUserShopLimit {
		limit = maxUserShopLimit
	}

	categoryCode := strings.ToLower(strings.TrimSpace(req.CategoryCode))

	rows, total, err := s.itemRepo.ListVisibleShopItems(s.db, model.ListVisibleShopItemsParam{
		UserID:       userID,
		Search:       strings.TrimSpace(req.Search),
		CategoryCode: categoryCode,
		Limit:        limit,
		Offset:       (page - 1) * limit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list shop items")
	}

	items := make([]model.UserShopItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapUserShopItemResponse(row))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.UserListShopItemsResponse{
		Items: items,
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

	category, err := s.itemCategoryRepo.GetItemCategory(tx, model.GetItemCategoryParam{
		ItemCategoryID: itemCategoryID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("item category not found")
		}
		return nil, appErrors.InternalServer("failed to get item category")
	}

	avatarID, err := s.resolveItemAvatarIDForCreate(tx, req.AvatarID, category.Code, imageURL)
	if err != nil {
		return nil, err
	}

	item := &entity.Item{
		ItemID:         uuid.New(),
		ItemCategoryID: itemCategoryID,
		AvatarID:       avatarID,
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

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID: adminUserID,
		ActionType:   model.AuditActionCreate,
		Module:       model.AuditModuleShop,
		TargetType:   "item",
		TargetID:     item.ItemID.String(),
		TargetLabel:  item.Name,
		Detail:       fmt.Sprintf("Created shop item %s", item.Name),
		PayloadAfter: newAuditItemSnapshot(savedItem),
	})
	if err != nil {
		return nil, err
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

func (s *ItemService) PurchaseShopItemForUser(userID uuid.UUID, itemID uuid.UUID) (*model.UserPurchaseShopItemResponse, error) {
	if userID == uuid.Nil {
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
	if item.Status != model.ItemStatusActive || !item.IsVisible {
		return nil, appErrors.Conflict("item is not available")
	}

	_, err = s.userItemRepo.GetUserItem(tx, model.GetUserItemParam{UserID: userID, ItemID: itemID})
	if err == nil {
		return nil, appErrors.Conflict("item already owned")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to check user item")
	}

	now := time.Now().UTC()
	userItem := &entity.UserItem{
		UserItemID:  uuid.New(),
		UserID:      userID,
		ItemID:      itemID,
		PurchasedAt: now,
	}
	if err := s.userItemRepo.CreateUserItem(tx, userItem); err != nil {
		return nil, appErrors.Conflict("item already owned")
	}

	responseItem, err := s.getUserShopItemResponse(tx, userID, itemID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.UserPurchaseShopItemResponse{
		Item: responseItem,
	}, nil
}

func (s *ItemService) EquipShopItemForUser(userID uuid.UUID, itemID uuid.UUID) (*model.UserEquipShopItemResponse, error) {
	if userID == uuid.Nil {
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
	if item.Status != model.ItemStatusActive || !item.IsVisible {
		return nil, appErrors.Conflict("item is not available")
	}

	category, err := s.itemCategoryRepo.GetItemCategory(tx, model.GetItemCategoryParam{ItemCategoryID: item.ItemCategoryID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get item category")
	}
	if category.Code != model.ItemCategoryAvatar {
		return nil, appErrors.BadRequest("only avatar items can be equipped")
	}
	if item.AvatarID == nil {
		return nil, appErrors.Conflict("item is not linked to an avatar")
	}

	userItem, err := s.userItemRepo.GetUserItem(tx, model.GetUserItemParam{UserID: userID, ItemID: itemID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Conflict("item is not owned")
		}
		return nil, appErrors.InternalServer("failed to get user item")
	}

	profile, err := s.userProfileRepo.GetUserProfileForUpdate(tx, model.GetUserProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user profile not found")
		}
		return nil, appErrors.InternalServer("failed to get user profile")
	}

	now := time.Now().UTC()
	if err := s.userItemRepo.ClearEquippedItemsByUserAndCategory(tx, userID, model.ItemCategoryAvatar); err != nil {
		return nil, appErrors.InternalServer("failed to clear equipped item")
	}

	userItem.EquippedAt = &now
	if err := s.userItemRepo.UpdateUserItem(tx, userItem); err != nil {
		return nil, appErrors.InternalServer("failed to equip item")
	}

	profile.AvatarID = item.AvatarID
	if err := s.userProfileRepo.UpdateUserProfile(tx, profile); err != nil {
		return nil, appErrors.InternalServer("failed to update user profile avatar")
	}

	responseItem, err := s.getUserShopItemResponse(tx, userID, itemID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.UserEquipShopItemResponse{
		Item:     responseItem,
		AvatarID: *item.AvatarID,
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

	category, err := s.itemCategoryRepo.GetItemCategory(tx, model.GetItemCategoryParam{
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

	before := newAuditItemSnapshot(item)
	oldImageURL := item.ImageURL
	finalImageURL := item.ImageURL
	if imageURL != "" {
		finalImageURL = imageURL
	}

	avatarID, err := s.resolveItemAvatarIDForUpdate(tx, req.AvatarID, category.Code, item.AvatarID, finalImageURL)
	if err != nil {
		return nil, err
	}

	item.ItemCategoryID = itemCategoryID
	item.AvatarID = avatarID
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

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionUpdate,
		Module:        model.AuditModuleShop,
		TargetType:    "item",
		TargetID:      item.ItemID.String(),
		TargetLabel:   item.Name,
		Detail:        fmt.Sprintf("Updated shop item %s", item.Name),
		PayloadBefore: before,
		PayloadAfter:  newAuditItemSnapshot(savedItem),
	})
	if err != nil {
		return nil, err
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

	before := newAuditItemSnapshot(item)
	err = s.itemRepo.DeleteItem(tx, item.ItemID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete item")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionDelete,
		Module:        model.AuditModuleShop,
		TargetType:    "item",
		TargetID:      item.ItemID.String(),
		TargetLabel:   item.Name,
		Detail:        fmt.Sprintf("Deleted shop item %s", item.Name),
		PayloadBefore: before,
	})
	if err != nil {
		return nil, err
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

	if priceCoin < 0 {
		return "", "", "", appErrors.BadRequest("price coin cannot be negative")
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

func (s *ItemService) resolveItemAvatarIDForCreate(tx *gorm.DB, raw string, categoryCode string, imageURL string) (*uuid.UUID, error) {
	avatarIDRaw := strings.TrimSpace(raw)
	if categoryCode != model.ItemCategoryAvatar {
		if avatarIDRaw != "" {
			return nil, appErrors.BadRequest("avatar id is only allowed for avatar item")
		}
		return nil, nil
	}

	if avatarIDRaw != "" {
		return s.validateExistingAvatarID(tx, avatarIDRaw)
	}

	avatar := &entity.Avatar{
		AvatarID:    uuid.New(),
		ImageURL:    imageURL,
		UnlockLevel: 1,
		Status:      "active",
	}
	if err := s.avatarRepo.CreateAvatar(tx, avatar); err != nil {
		return nil, appErrors.InternalServer("failed to create avatar")
	}

	return &avatar.AvatarID, nil
}

func (s *ItemService) resolveItemAvatarIDForUpdate(tx *gorm.DB, raw string, categoryCode string, currentAvatarID *uuid.UUID, imageURL string) (*uuid.UUID, error) {
	avatarIDRaw := strings.TrimSpace(raw)
	if categoryCode != model.ItemCategoryAvatar {
		if avatarIDRaw != "" {
			return nil, appErrors.BadRequest("avatar id is only allowed for avatar item")
		}
		return nil, nil
	}

	if avatarIDRaw != "" {
		avatarID, err := s.validateExistingAvatarID(tx, avatarIDRaw)
		if err != nil {
			return nil, err
		}
		if err := s.syncAvatarImageURL(tx, avatarID, imageURL); err != nil {
			return nil, err
		}
		return avatarID, nil
	}

	if currentAvatarID != nil {
		if err := s.syncAvatarImageURL(tx, currentAvatarID, imageURL); err != nil {
			return nil, err
		}
		return currentAvatarID, nil
	}

	avatar := &entity.Avatar{
		AvatarID:    uuid.New(),
		ImageURL:    imageURL,
		UnlockLevel: 1,
		Status:      "active",
	}
	if err := s.avatarRepo.CreateAvatar(tx, avatar); err != nil {
		return nil, appErrors.InternalServer("failed to create avatar")
	}

	return &avatar.AvatarID, nil
}

func (s *ItemService) validateExistingAvatarID(tx *gorm.DB, raw string) (*uuid.UUID, error) {
	avatarID, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, appErrors.BadRequest("invalid avatar id")
	}

	avatar, err := s.avatarRepo.GetAvatar(tx, model.GetAvatarParam{AvatarID: avatarID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.BadRequest("avatar not found")
		}
		return nil, appErrors.InternalServer("failed to get avatar")
	}
	if avatar.Status != "active" {
		return nil, appErrors.BadRequest("avatar is not active")
	}

	return &avatarID, nil
}

func (s *ItemService) syncAvatarImageURL(tx *gorm.DB, avatarID *uuid.UUID, imageURL string) error {
	if avatarID == nil {
		return nil
	}

	avatar, err := s.avatarRepo.GetAvatar(tx, model.GetAvatarParam{AvatarID: *avatarID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.BadRequest("avatar not found")
		}
		return appErrors.InternalServer("failed to get avatar")
	}
	if avatar.ImageURL == imageURL {
		return nil
	}

	avatar.ImageURL = imageURL
	if err := s.avatarRepo.UpdateAvatar(tx, avatar); err != nil {
		return appErrors.InternalServer("failed to update avatar")
	}

	return nil
}

func (s *ItemService) getUserShopItemResponse(tx *gorm.DB, userID uuid.UUID, itemID uuid.UUID) (model.UserShopItemResponse, error) {
	rows, _, err := s.itemRepo.ListVisibleShopItems(tx, model.ListVisibleShopItemsParam{
		UserID: userID,
		ItemID: itemID,
		Limit:  1,
	})
	if err != nil {
		return model.UserShopItemResponse{}, appErrors.InternalServer("failed to get shop item")
	}
	if len(rows) == 0 {
		return model.UserShopItemResponse{}, appErrors.NotFound("shop item not found")
	}

	return mapUserShopItemResponse(rows[0]), nil
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
		AvatarID:       item.AvatarID,
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

func mapUserShopItemResponse(row model.UserShopItemRow) model.UserShopItemResponse {
	isOwned := row.UserItemID != nil
	isEquipped := row.AvatarID != nil && row.CurrentAvatarID != nil && *row.AvatarID == *row.CurrentAvatarID

	ownershipStatus := model.UserShopItemOwnershipNotOwned
	if isOwned {
		ownershipStatus = model.UserShopItemOwnershipOwned
	}
	if isEquipped {
		ownershipStatus = model.UserShopItemOwnershipEquipped
		isOwned = true
	}

	return model.UserShopItemResponse{
		ItemID:          row.ItemID,
		ItemCategoryID:  row.ItemCategoryID,
		CategoryCode:    row.CategoryCode,
		CategoryName:    row.CategoryName,
		AvatarID:        row.AvatarID,
		Name:            row.Name,
		Description:     row.Description,
		PriceCoin:       row.PriceCoin,
		ImageURL:        row.ImageURL,
		OwnershipStatus: ownershipStatus,
		IsOwned:         isOwned,
		IsEquipped:      isEquipped,
		CanPurchase:     !isOwned,
		CanEquip:        row.CategoryCode == model.ItemCategoryAvatar && isOwned && !isEquipped && row.AvatarID != nil,
	}
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

func mapUserItemCategoryResponse(category entity.ItemCategory) model.UserItemCategoryResponse {
	return model.UserItemCategoryResponse{
		ItemCategoryID: category.ItemCategoryID,
		Code:           category.Code,
		Name:           category.Name,
		Description:    category.Description,
	}
}
