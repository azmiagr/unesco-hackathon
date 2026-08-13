package service

import (
	"encoding/csv"
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
	defaultAdminRedeemItemLimit = 10
	maxAdminRedeemItemLimit     = 100
	maxRedeemItemImageSize      = 2 * 1024 * 1024
	defaultAdminRedeemCodeLimit = 10
	maxAdminRedeemCodeLimit     = 100
	maxRedeemCodeCSVSize        = 2 * 1024 * 1024
	redeemCodeDateLayout        = "2006-01-02"
)

var allowedRedeemItemStatuses = map[string]bool{
	model.RedeemItemStatusActive:   true,
	model.RedeemItemStatusInactive: true,
	model.RedeemItemStatusRetired:  true,
}

var allowedRedeemClaimPeriods = map[string]bool{
	model.RedeemClaimPeriodDaily:   true,
	model.RedeemClaimPeriodWeekly:  true,
	model.RedeemClaimPeriodMonthly: true,
}

var allowedRedeemCodeStatuses = map[string]bool{
	model.RedeemCodeStatusAvailable: true,
	model.RedeemCodeStatusClaimed:   true,
	model.RedeemCodeStatusExpired:   true,
}

type IRedeemService interface {
	ListRedeemTypesByAdmin(req model.AdminListRedeemTypesRequest) (*model.AdminListRedeemTypesResponse, error)
	ListRedeemItemsByAdmin(req model.AdminListRedeemItemsRequest) (*model.AdminListRedeemItemsResponse, error)
	ListRedeemCodesByAdmin(req model.AdminListRedeemCodesRequest) (*model.AdminListRedeemCodesResponse, error)
	GetRedeemItemDetailByAdmin(redeemItemID uuid.UUID) (*model.AdminGetRedeemItemDetailResponse, error)
	CreateRedeemItemByAdmin(adminUserID uuid.UUID, req model.AdminCreateRedeemItemRequest) (*model.AdminCreateRedeemItemResponse, error)
	CreateRedeemCodeManualByAdmin(adminUserID uuid.UUID, req model.AdminCreateRedeemCodeManualRequest) (*model.AdminCreateRedeemCodeResponse, error)
	CreateRedeemCodesCSVByAdmin(adminUserID uuid.UUID, req model.AdminCreateRedeemCodeCSVRequest) (*model.AdminCreateRedeemCodesCSVResponse, error)
	UpdateRedeemItemByAdmin(adminUserID uuid.UUID, redeemItemID uuid.UUID, req model.AdminUpdateRedeemItemRequest) (*model.AdminUpdateRedeemItemResponse, error)
	DeleteRedeemItemByAdmin(adminUserID uuid.UUID, redeemItemID uuid.UUID) (*model.AdminDeleteRedeemItemResponse, error)
	DeleteRedeemCodeByAdmin(adminUserID uuid.UUID, redeemCodeID uuid.UUID) (*model.AdminDeleteRedeemCodeResponse, error)
}

type RedeemService struct {
	db             *gorm.DB
	redeemItemRepo repository.IRedeemItemRepository
	redeemTypeRepo repository.IRedeemTypeRepository
	redeemCodeRepo repository.IRedeemCodeRepository
	userRepo       repository.IUserRepository
	auditLogRepo   repository.IAuditLogRepository
	storage        supabase.Interface
}

func NewRedeemService(
	redeemItemRepo repository.IRedeemItemRepository,
	redeemTypeRepo repository.IRedeemTypeRepository,
	redeemCodeRepo repository.IRedeemCodeRepository,
	userRepo repository.IUserRepository,
	auditLogRepo repository.IAuditLogRepository,
	storage supabase.Interface,
) IRedeemService {
	return &RedeemService{
		db:             mariadb.Connection,
		redeemItemRepo: redeemItemRepo,
		redeemTypeRepo: redeemTypeRepo,
		redeemCodeRepo: redeemCodeRepo,
		userRepo:       userRepo,
		auditLogRepo:   auditLogRepo,
		storage:        storage,
	}
}

func (s *RedeemService) ListRedeemTypesByAdmin(req model.AdminListRedeemTypesRequest) (*model.AdminListRedeemTypesResponse, error) {
	redeemTypes, err := s.redeemTypeRepo.ListRedeemTypes(s.db, model.ListRedeemTypesParam{
		Search: strings.TrimSpace(req.Search),
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list redeem types")
	}

	responses := make([]model.AdminRedeemTypeResponse, 0, len(redeemTypes))
	for _, redeemType := range redeemTypes {
		responses = append(responses, mapAdminRedeemTypeResponse(redeemType))
	}

	return &model.AdminListRedeemTypesResponse{
		Types: responses,
	}, nil
}

func (s *RedeemService) ListRedeemItemsByAdmin(req model.AdminListRedeemItemsRequest) (*model.AdminListRedeemItemsResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultAdminRedeemItemLimit
	}
	if limit > maxAdminRedeemItemLimit {
		limit = maxAdminRedeemItemLimit
	}

	redeemTypeID := uuid.Nil
	if strings.TrimSpace(req.RedeemTypeID) != "" {
		parsedID, err := uuid.Parse(strings.TrimSpace(req.RedeemTypeID))
		if err != nil {
			return nil, appErrors.BadRequest("invalid redeem type id")
		}
		redeemTypeID = parsedID
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "" && !allowedRedeemItemStatuses[status] {
		return nil, appErrors.BadRequest("invalid status filter")
	}

	claimPeriod := strings.ToLower(strings.TrimSpace(req.ClaimPeriod))
	if claimPeriod != "" && !allowedRedeemClaimPeriods[claimPeriod] {
		return nil, appErrors.BadRequest("invalid claim period filter")
	}

	redeemItems, total, err := s.redeemItemRepo.ListRedeemItems(s.db, model.ListRedeemItemsParam{
		Search:       strings.TrimSpace(req.Search),
		RedeemTypeID: redeemTypeID,
		TypeCode:     normalizeRedeemTypeCode(req.TypeCode),
		Status:       status,
		ClaimPeriod:  claimPeriod,
		Limit:        limit,
		Offset:       (page - 1) * limit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list redeem items")
	}

	responses := make([]model.AdminRedeemItemResponse, 0, len(redeemItems))
	for _, redeemItem := range redeemItems {
		redeemItemResponse, err := s.mapAdminRedeemItemResponse(s.db, redeemItem)
		if err != nil {
			return nil, err
		}
		responses = append(responses, redeemItemResponse)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.AdminListRedeemItemsResponse{
		Items: responses,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *RedeemService) ListRedeemCodesByAdmin(req model.AdminListRedeemCodesRequest) (*model.AdminListRedeemCodesResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultAdminRedeemCodeLimit
	}
	if limit > maxAdminRedeemCodeLimit {
		limit = maxAdminRedeemCodeLimit
	}

	redeemItemID := uuid.Nil
	if strings.TrimSpace(req.RedeemItemID) != "" {
		parsedID, err := uuid.Parse(strings.TrimSpace(req.RedeemItemID))
		if err != nil {
			return nil, appErrors.BadRequest("invalid redeem item id")
		}
		redeemItemID = parsedID
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "" && !allowedRedeemCodeStatuses[status] {
		return nil, appErrors.BadRequest("invalid redeem code status")
	}

	rows, total, err := s.redeemCodeRepo.ListRedeemCodes(s.db, model.ListRedeemCodesParam{
		Search:       strings.TrimSpace(req.Search),
		RedeemItemID: redeemItemID,
		Status:       status,
		Limit:        limit,
		Offset:       (page - 1) * limit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list redeem codes")
	}

	responses := make([]model.AdminRedeemCodeResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, mapAdminRedeemCodeResponse(row))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.AdminListRedeemCodesResponse{
		RedeemCodes: responses,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *RedeemService) GetRedeemItemDetailByAdmin(redeemItemID uuid.UUID) (*model.AdminGetRedeemItemDetailResponse, error) {
	if redeemItemID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid redeem item id")
	}

	redeemItem, err := s.redeemItemRepo.GetRedeemItem(s.db, model.GetRedeemItemParam{RedeemItemID: redeemItemID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("redeem item not found")
		}
		return nil, appErrors.InternalServer("failed to get redeem item detail")
	}

	redeemItemResponse, err := s.mapAdminRedeemItemResponse(s.db, *redeemItem)
	if err != nil {
		return nil, err
	}

	return &model.AdminGetRedeemItemDetailResponse{
		Item: redeemItemResponse,
	}, nil
}

func (s *RedeemService) CreateRedeemItemByAdmin(
	adminUserID uuid.UUID,
	req model.AdminCreateRedeemItemRequest,
) (*model.AdminCreateRedeemItemResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	payload, err := validateAdminRedeemItemPayload(
		req.Name,
		req.Type,
		req.PartnerName,
		req.Description,
		req.PriceCoin,
		req.MaxClaimPerPeriod,
		req.ClaimPeriod,
		req.MinimumLevel,
		req.Status,
	)
	if err != nil {
		return nil, err
	}
	if req.Image == nil {
		return nil, appErrors.BadRequest("redeem item image is required")
	}

	imageURL, err := s.uploadRedeemItemImage(req.Image)
	if err != nil {
		return nil, err
	}

	shouldDeleteImage := true
	defer func() {
		if shouldDeleteImage && imageURL != "" {
			_ = supabase.DeleteFileIfPresent(s.storage, imageURL)
		}
	}()

	isStockVisible := true
	if req.IsStockVisible != nil {
		isStockVisible = *req.IsStockVisible
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	redeemType, err := s.getOrCreateRedeemType(tx, payload.TypeCode)
	if err != nil {
		return nil, err
	}

	redeemItem := &entity.RedeemItem{
		RedeemItemID:      uuid.New(),
		RedeemTypeID:      redeemType.RedeemTypeID,
		Name:              payload.Name,
		PartnerName:       payload.PartnerName,
		Description:       payload.Description,
		PriceCoin:         payload.PriceCoin,
		MaxClaimPerPeriod: payload.MaxClaimPerPeriod,
		ClaimPeriod:       payload.ClaimPeriod,
		MinimumLevel:      payload.MinimumLevel,
		ImageURL:          imageURL,
		IsStockVisible:    isStockVisible,
		Status:            payload.Status,
	}

	err = s.redeemItemRepo.CreateRedeemItem(tx, redeemItem)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create redeem item")
	}

	savedRedeemItem, err := s.redeemItemRepo.GetRedeemItem(tx, model.GetRedeemItemParam{RedeemItemID: redeemItem.RedeemItemID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get created redeem item")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID: adminUserID,
		ActionType:   model.AuditActionCreate,
		Module:       model.AuditModuleShop,
		TargetType:   "redeem_item",
		TargetID:     redeemItem.RedeemItemID.String(),
		TargetLabel:  redeemItem.Name,
		Detail:       fmt.Sprintf("Created redeem item %s", redeemItem.Name),
		PayloadAfter: newAuditRedeemItemSnapshot(savedRedeemItem),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteImage = false

	redeemItemResponse, err := s.mapAdminRedeemItemResponse(s.db, *savedRedeemItem)
	if err != nil {
		return nil, err
	}

	return &model.AdminCreateRedeemItemResponse{
		Item: redeemItemResponse,
	}, nil
}

func (s *RedeemService) CreateRedeemCodeManualByAdmin(
	adminUserID uuid.UUID,
	req model.AdminCreateRedeemCodeManualRequest,
) (*model.AdminCreateRedeemCodeResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	redeemItemID, err := uuid.Parse(strings.TrimSpace(req.RedeemItemID))
	if err != nil {
		return nil, appErrors.BadRequest("invalid redeem item id")
	}

	code, err := normalizeRedeemCode(req.Code)
	if err != nil {
		return nil, err
	}

	expiresAt, err := parseRedeemCodeExpiry(req.ExpiresAt)
	if err != nil {
		return nil, appErrors.BadRequest("invalid expires_at")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	redeemItem, err := s.redeemItemRepo.GetRedeemItem(tx, model.GetRedeemItemParam{RedeemItemID: redeemItemID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("redeem item not found")
		}
		return nil, appErrors.InternalServer("failed to get redeem item")
	}

	redeemCode := &entity.RedeemCode{
		RedeemCodeID:     uuid.New(),
		RedeemItemID:     redeemItem.RedeemItemID,
		Code:             code,
		ExpiresAt:        expiresAt,
		CreatedByAdminID: &adminUserID,
	}

	err = s.redeemCodeRepo.CreateRedeemCode(tx, redeemCode)
	if err != nil {
		return nil, appErrors.Conflict("redeem code already exists")
	}

	row, err := s.redeemCodeRepo.GetRedeemCodeRow(tx, redeemCode.RedeemCodeID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get created redeem code")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateRedeemCodeResponse{
		RedeemCode: mapAdminRedeemCodeResponse(*row),
	}, nil
}

func (s *RedeemService) CreateRedeemCodesCSVByAdmin(
	adminUserID uuid.UUID,
	req model.AdminCreateRedeemCodeCSVRequest,
) (*model.AdminCreateRedeemCodesCSVResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if req.File == nil {
		return nil, appErrors.BadRequest("csv file is required")
	}
	if req.File.Size > maxRedeemCodeCSVSize {
		return nil, appErrors.BadRequest("csv file is too large")
	}

	file, err := req.File.Open()
	if err != nil {
		return nil, appErrors.BadRequest("failed to open csv file")
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, appErrors.BadRequest("invalid csv file")
	}
	if len(records) < 2 {
		return nil, appErrors.BadRequest("csv file is empty")
	}

	if err := validateRedeemCodeCSVHeader(records[0]); err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	redeemCodes := make([]entity.RedeemCode, 0, len(records)-1)
	seenCodes := map[string]bool{}
	redeemItemLookup, availableRedeemItemNames, err := s.buildRedeemItemNameLookup(tx)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get redeem items")
	}

	for index, record := range records[1:] {
		lineNumber := index + 2
		if len(record) != 3 {
			return nil, appErrors.BadRequest(fmt.Sprintf("invalid csv row at line %d", lineNumber))
		}

		code, err := normalizeRedeemCode(record[0])
		if err != nil {
			return nil, appErrors.BadRequest(fmt.Sprintf("invalid code at line %d", lineNumber))
		}
		if seenCodes[code] {
			return nil, appErrors.BadRequest(fmt.Sprintf("duplicate code in csv at line %d", lineNumber))
		}
		seenCodes[code] = true

		redeemItemName := strings.TrimSpace(record[1])
		if redeemItemName == "" {
			return nil, appErrors.BadRequest(fmt.Sprintf("redeem item name is required at line %d", lineNumber))
		}

		redeemItem, ok := redeemItemLookup[normalizeLookupName(redeemItemName)]
		if !ok {
			return nil, appErrors.BadRequest(fmt.Sprintf(
				"redeem item %q not found at line %d. available redeem items: %s",
				redeemItemName,
				lineNumber,
				strings.Join(availableRedeemItemNames, ", "),
			))
		}

		expiresAt, err := parseRedeemCodeExpiry(record[2])
		if err != nil {
			return nil, appErrors.BadRequest(fmt.Sprintf("invalid expires_at at line %d", lineNumber))
		}

		redeemCodes = append(redeemCodes, entity.RedeemCode{
			RedeemCodeID:     uuid.New(),
			RedeemItemID:     redeemItem.RedeemItemID,
			Code:             code,
			ExpiresAt:        expiresAt,
			CreatedByAdminID: &adminUserID,
		})
	}

	err = s.redeemCodeRepo.CreateRedeemCodes(tx, redeemCodes)
	if err != nil {
		return nil, appErrors.Conflict("one or more redeem codes already exist")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateRedeemCodesCSVResponse{
		CreatedCount: len(redeemCodes),
	}, nil
}

func (s *RedeemService) UpdateRedeemItemByAdmin(
	adminUserID uuid.UUID,
	redeemItemID uuid.UUID,
	req model.AdminUpdateRedeemItemRequest,
) (*model.AdminUpdateRedeemItemResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if redeemItemID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid redeem item id")
	}

	payload, err := validateAdminRedeemItemPayload(
		req.Name,
		req.Type,
		req.PartnerName,
		req.Description,
		req.PriceCoin,
		req.MaxClaimPerPeriod,
		req.ClaimPeriod,
		req.MinimumLevel,
		req.Status,
	)
	if err != nil {
		return nil, err
	}

	imageURL := ""
	if req.Image != nil {
		imageURL, err = s.uploadRedeemItemImage(req.Image)
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

	redeemType, err := s.getOrCreateRedeemType(tx, payload.TypeCode)
	if err != nil {
		return nil, err
	}

	redeemItem, err := s.redeemItemRepo.GetRedeemItemForUpdate(tx, redeemItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("redeem item not found")
		}
		return nil, appErrors.InternalServer("failed to get redeem item")
	}

	before := newAuditRedeemItemSnapshot(redeemItem)
	oldImageURL := redeemItem.ImageURL
	redeemItem.RedeemTypeID = redeemType.RedeemTypeID
	redeemItem.Name = payload.Name
	redeemItem.PartnerName = payload.PartnerName
	redeemItem.Description = payload.Description
	redeemItem.PriceCoin = payload.PriceCoin
	redeemItem.MaxClaimPerPeriod = payload.MaxClaimPerPeriod
	redeemItem.ClaimPeriod = payload.ClaimPeriod
	redeemItem.MinimumLevel = payload.MinimumLevel
	redeemItem.Status = payload.Status
	if req.IsStockVisible != nil {
		redeemItem.IsStockVisible = *req.IsStockVisible
	}
	if imageURL != "" {
		redeemItem.ImageURL = imageURL
	}

	err = s.redeemItemRepo.UpdateRedeemItem(tx, redeemItem)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update redeem item")
	}

	savedRedeemItem, err := s.redeemItemRepo.GetRedeemItem(tx, model.GetRedeemItemParam{RedeemItemID: redeemItem.RedeemItemID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get updated redeem item")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionUpdate,
		Module:        model.AuditModuleShop,
		TargetType:    "redeem_item",
		TargetID:      redeemItem.RedeemItemID.String(),
		TargetLabel:   redeemItem.Name,
		Detail:        fmt.Sprintf("Updated redeem item %s", redeemItem.Name),
		PayloadBefore: before,
		PayloadAfter:  newAuditRedeemItemSnapshot(savedRedeemItem),
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

	redeemItemResponse, err := s.mapAdminRedeemItemResponse(s.db, *savedRedeemItem)
	if err != nil {
		return nil, err
	}

	return &model.AdminUpdateRedeemItemResponse{
		Item: redeemItemResponse,
	}, nil
}

func (s *RedeemService) DeleteRedeemItemByAdmin(
	adminUserID uuid.UUID,
	redeemItemID uuid.UUID,
) (*model.AdminDeleteRedeemItemResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if redeemItemID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid redeem item id")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	redeemItem, err := s.redeemItemRepo.GetRedeemItemForUpdate(tx, redeemItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("redeem item not found")
		}
		return nil, appErrors.InternalServer("failed to get redeem item")
	}

	before := newAuditRedeemItemSnapshot(redeemItem)
	err = s.redeemItemRepo.DeleteRedeemItem(tx, redeemItem.RedeemItemID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete redeem item")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionDelete,
		Module:        model.AuditModuleShop,
		TargetType:    "redeem_item",
		TargetID:      redeemItem.RedeemItemID.String(),
		TargetLabel:   redeemItem.Name,
		Detail:        fmt.Sprintf("Deleted redeem item %s", redeemItem.Name),
		PayloadBefore: before,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminDeleteRedeemItemResponse{
		RedeemItemID: redeemItem.RedeemItemID,
	}, nil
}

func (s *RedeemService) buildRedeemItemNameLookup(tx *gorm.DB) (map[string]entity.RedeemItem, []string, error) {
	redeemItems, err := s.redeemItemRepo.ListRedeemItemsByNormalizedNameLookup(tx)
	if err != nil {
		return nil, nil, err
	}

	lookup := make(map[string]entity.RedeemItem, len(redeemItems))
	names := make([]string, 0, len(redeemItems))
	for _, redeemItem := range redeemItems {
		lookup[normalizeLookupName(redeemItem.Name)] = redeemItem
		names = append(names, redeemItem.Name)
	}

	if len(names) == 0 {
		names = append(names, "-")
	}

	return lookup, names, nil
}

func (s *RedeemService) DeleteRedeemCodeByAdmin(
	adminUserID uuid.UUID,
	redeemCodeID uuid.UUID,
) (*model.AdminDeleteRedeemCodeResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if redeemCodeID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid redeem code id")
	}

	_, err := s.redeemCodeRepo.GetRedeemCode(s.db, model.GetRedeemCodeParam{RedeemCodeID: redeemCodeID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("redeem code not found")
		}
		return nil, appErrors.InternalServer("failed to get redeem code")
	}

	err = s.redeemCodeRepo.DeleteRedeemCode(s.db, redeemCodeID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete redeem code")
	}

	return &model.AdminDeleteRedeemCodeResponse{
		RedeemCodeID: redeemCodeID,
	}, nil
}

type adminRedeemItemPayload struct {
	Name              string
	TypeCode          string
	PartnerName       string
	Description       string
	PriceCoin         int
	MaxClaimPerPeriod int
	ClaimPeriod       string
	MinimumLevel      int
	Status            string
}

func validateAdminRedeemItemPayload(
	nameRaw string,
	typeRaw string,
	partnerNameRaw string,
	descriptionRaw string,
	priceCoin int,
	maxClaimPerPeriod int,
	claimPeriodRaw string,
	minimumLevel int,
	statusRaw string,
) (adminRedeemItemPayload, error) {
	name, err := helper.RequireTrimmedString(nameRaw, "name is required")
	if err != nil {
		return adminRedeemItemPayload{}, err
	}
	if len(name) > 150 {
		return adminRedeemItemPayload{}, appErrors.BadRequest("name is too long")
	}

	typeCode := normalizeRedeemTypeCode(typeRaw)
	if typeCode == "" {
		return adminRedeemItemPayload{}, appErrors.BadRequest("type is required")
	}
	if len(typeCode) > 80 {
		return adminRedeemItemPayload{}, appErrors.BadRequest("type is too long")
	}

	partnerName, err := helper.RequireTrimmedString(partnerNameRaw, "partner name is required")
	if err != nil {
		return adminRedeemItemPayload{}, err
	}
	if len(partnerName) > 150 {
		return adminRedeemItemPayload{}, appErrors.BadRequest("partner name is too long")
	}

	description, err := helper.RequireTrimmedString(descriptionRaw, "description is required")
	if err != nil {
		return adminRedeemItemPayload{}, err
	}

	if priceCoin < 1 {
		return adminRedeemItemPayload{}, appErrors.BadRequest("price coin must be greater than 0")
	}
	if maxClaimPerPeriod < 1 {
		return adminRedeemItemPayload{}, appErrors.BadRequest("max claim per period must be greater than 0")
	}

	claimPeriod := strings.ToLower(strings.TrimSpace(claimPeriodRaw))
	if !allowedRedeemClaimPeriods[claimPeriod] {
		return adminRedeemItemPayload{}, appErrors.BadRequest("invalid claim period")
	}

	if minimumLevel < 1 {
		minimumLevel = 1
	}

	status := strings.ToLower(strings.TrimSpace(statusRaw))
	if status == "" {
		status = model.RedeemItemStatusActive
	}
	if !allowedRedeemItemStatuses[status] {
		return adminRedeemItemPayload{}, appErrors.BadRequest("invalid redeem item status")
	}

	return adminRedeemItemPayload{
		Name:              name,
		TypeCode:          typeCode,
		PartnerName:       partnerName,
		Description:       description,
		PriceCoin:         priceCoin,
		MaxClaimPerPeriod: maxClaimPerPeriod,
		ClaimPeriod:       claimPeriod,
		MinimumLevel:      minimumLevel,
		Status:            status,
	}, nil
}

func (s *RedeemService) getOrCreateRedeemType(tx *gorm.DB, typeCode string) (*entity.RedeemType, error) {
	redeemType, err := s.redeemTypeRepo.GetRedeemType(tx, model.GetRedeemTypeParam{Code: typeCode})
	if err == nil {
		return redeemType, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to get redeem type")
	}

	redeemType = &entity.RedeemType{
		RedeemTypeID: uuid.New(),
		Code:         typeCode,
		Name:         formatRedeemTypeName(typeCode),
	}
	err = s.redeemTypeRepo.CreateRedeemType(tx, redeemType)
	if err != nil {
		redeemType, getErr := s.redeemTypeRepo.GetRedeemType(tx, model.GetRedeemTypeParam{Code: typeCode})
		if getErr == nil {
			return redeemType, nil
		}
		return nil, appErrors.InternalServer("failed to create redeem type")
	}

	return redeemType, nil
}

func normalizeRedeemTypeCode(raw string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(raw))), "_")
}

func formatRedeemTypeName(code string) string {
	parts := strings.Split(code, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, " ")
}

func normalizeRedeemCode(raw string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if code == "" {
		return "", appErrors.BadRequest("redeem code is required")
	}
	if len(code) > 120 {
		return "", appErrors.BadRequest("redeem code is too long")
	}

	return code, nil
}

func parseRedeemCodeExpiry(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, appErrors.BadRequest("expires_at is required")
	}

	parsed, err := time.Parse(redeemCodeDateLayout, value)
	if err != nil {
		return time.Time{}, err
	}

	return parsed.UTC(), nil
}

func validateRedeemCodeCSVHeader(header []string) error {
	if len(header) != 3 {
		return appErrors.BadRequest("invalid csv header")
	}

	expected := []string{"code", "redeem_item_name", "expires_at"}
	for index, value := range expected {
		if strings.ToLower(strings.TrimSpace(header[index])) != value {
			return appErrors.BadRequest("invalid csv header")
		}
	}

	return nil
}

func normalizeLookupName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}

func (s *RedeemService) uploadRedeemItemImage(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", nil
	}

	url, err := supabase.UploadOptionalImage(
		s.storage,
		file,
		maxRedeemItemImageSize,
		"redeem item image size exceeds 2MB limit",
	)
	if err != nil {
		return "", appErrors.BadRequest("failed to upload redeem item image")
	}

	return url, nil
}

func (s *RedeemService) mapAdminRedeemItemResponse(tx *gorm.DB, redeemItem entity.RedeemItem) (model.AdminRedeemItemResponse, error) {
	redeemType, err := s.redeemTypeRepo.GetRedeemType(tx, model.GetRedeemTypeParam{
		RedeemTypeID: redeemItem.RedeemTypeID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.AdminRedeemItemResponse{}, appErrors.InternalServer("redeem type not found")
		}
		return model.AdminRedeemItemResponse{}, appErrors.InternalServer("failed to get redeem type")
	}

	return model.AdminRedeemItemResponse{
		RedeemItemID:      redeemItem.RedeemItemID,
		RedeemTypeID:      redeemItem.RedeemTypeID,
		Type:              mapAdminRedeemTypeResponse(*redeemType),
		Name:              redeemItem.Name,
		PartnerName:       redeemItem.PartnerName,
		Description:       redeemItem.Description,
		PriceCoin:         redeemItem.PriceCoin,
		MaxClaimPerPeriod: redeemItem.MaxClaimPerPeriod,
		ClaimPeriod:       redeemItem.ClaimPeriod,
		MinimumLevel:      redeemItem.MinimumLevel,
		ImageURL:          redeemItem.ImageURL,
		IsStockVisible:    redeemItem.IsStockVisible,
		Status:            redeemItem.Status,
		CreatedAt:         redeemItem.CreatedAt,
		UpdatedAt:         redeemItem.UpdatedAt,
	}, nil
}

func mapAdminRedeemCodeResponse(row model.AdminRedeemCodeListRow) model.AdminRedeemCodeResponse {
	return model.AdminRedeemCodeResponse{
		RedeemCodeID:    row.RedeemCodeID,
		RedeemItemID:    row.RedeemItemID,
		RedeemItemName:  row.RedeemItemName,
		Code:            row.Code,
		Status:          row.Status,
		ClaimedByUserID: row.ClaimedByUserID,
		ClaimedBy:       row.ClaimedBy,
		ClaimedAt:       row.ClaimedAt,
		ExpiresAt:       row.ExpiresAt,
		CreatedAt:       row.CreatedAt,
	}
}

func mapAdminRedeemTypeResponse(redeemType entity.RedeemType) model.AdminRedeemTypeResponse {
	return model.AdminRedeemTypeResponse{
		RedeemTypeID: redeemType.RedeemTypeID,
		Code:         redeemType.Code,
		Name:         redeemType.Name,
		CreatedAt:    redeemType.CreatedAt,
		UpdatedAt:    redeemType.UpdatedAt,
	}
}
