package service

import (
	"errors"
	"math"
	"mime/multipart"
	"strings"
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ITitleService interface {
	ListTitlesForUser(userID uuid.UUID) (*model.ListUserTitlesResponse, error)
	EquipTitleForUser(userID uuid.UUID, titleID uuid.UUID) (*model.EquipTitleResponse, error)
	ListTitlesByAdmin(req model.AdminListTitlesRequest) (*model.AdminListTitlesResponse, error)
	GetTitleByAdmin(titleID uuid.UUID) (*model.AdminGetTitleResponse, error)
	CreateTitleByAdmin(req model.AdminCreateTitleRequest) (*model.AdminCreateTitleResponse, error)
	UpdateTitleByAdmin(titleID uuid.UUID, req model.AdminUpdateTitleRequest) (*model.AdminUpdateTitleResponse, error)
	DeleteTitleByAdmin(titleID uuid.UUID) (*model.AdminDeleteTitleResponse, error)
}

type TitleService struct {
	db              *gorm.DB
	titleRepo       repository.ITitleRepository
	userItemRepo    repository.IUserItemRepository
	userProfileRepo repository.IUserProfileRepository
	storage         supabase.Interface
}

func NewTitleService(titleRepo repository.ITitleRepository, userItemRepo repository.IUserItemRepository, userProfileRepo repository.IUserProfileRepository, storage supabase.Interface) ITitleService {
	return &TitleService{
		db:              mariadb.Connection,
		titleRepo:       titleRepo,
		userItemRepo:    userItemRepo,
		userProfileRepo: userProfileRepo,
		storage:         storage,
	}
}

func (s *TitleService) ListTitlesByAdmin(req model.AdminListTitlesRequest) (*model.AdminListTitlesResponse, error) {
	page, limit := normalizeTitlePagination(req.Page, req.Limit)
	titles, total, err := s.titleRepo.ListTitles(s.db, strings.TrimSpace(req.Search), limit, (page-1)*limit)
	if err != nil {
		return nil, appErrors.InternalServer("failed to list titles")
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.AdminListTitlesResponse{
		Titles: mapAdminTitleResponses(titles),
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *TitleService) GetTitleByAdmin(titleID uuid.UUID) (*model.AdminGetTitleResponse, error) {
	if titleID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid title id")
	}

	title, err := s.titleRepo.GetTitle(s.db, model.GetTitleParam{TitleID: titleID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("title not found")
		}
		return nil, appErrors.InternalServer("failed to get title")
	}

	return &model.AdminGetTitleResponse{
		Title: mapAdminTitleResponse(*title),
	}, nil
}

func (s *TitleService) CreateTitleByAdmin(req model.AdminCreateTitleRequest) (*model.AdminCreateTitleResponse, error) {
	title, err := newTitleEntity(req.Title, req.UnlockLevel, req.Status)
	if err != nil {
		return nil, err
	}
	if req.Image == nil {
		return nil, appErrors.BadRequest("title image is required")
	}

	imageURL, err := s.uploadTitleImage(req.Image)
	if err != nil {
		return nil, err
	}
	shouldDeleteImage := true
	defer func() {
		if shouldDeleteImage {
			_ = supabase.DeleteFileIfPresent(s.storage, imageURL)
		}
	}()
	title.ImageBorder = imageURL

	err = s.titleRepo.CreateTitle(s.db, title)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create title")
	}
	shouldDeleteImage = false

	return &model.AdminCreateTitleResponse{
		Title: mapAdminTitleResponse(*title),
	}, nil
}

func (s *TitleService) UpdateTitleByAdmin(titleID uuid.UUID, req model.AdminUpdateTitleRequest) (*model.AdminUpdateTitleResponse, error) {
	if titleID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid title id")
	}

	title, err := s.titleRepo.GetTitleForUpdate(s.db, titleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("title not found")
		}
		return nil, appErrors.InternalServer("failed to get title")
	}

	updated, err := newTitleEntity(req.Title, req.UnlockLevel, req.Status)
	if err != nil {
		return nil, err
	}

	imageURL := ""
	if req.Image != nil {
		imageURL, err = s.uploadTitleImage(req.Image)
		if err != nil {
			return nil, err
		}
	}
	shouldDeleteNewImage := imageURL != ""
	defer func() {
		if shouldDeleteNewImage {
			_ = supabase.DeleteFileIfPresent(s.storage, imageURL)
		}
	}()

	oldImageURL := title.ImageBorder
	title.Title = updated.Title
	title.UnlockLevel = updated.UnlockLevel
	title.Status = updated.Status
	if imageURL != "" {
		title.ImageBorder = imageURL
	}

	err = s.titleRepo.UpdateTitle(s.db, title)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update title")
	}
	shouldDeleteNewImage = false
	if imageURL != "" && oldImageURL != "" && oldImageURL != imageURL {
		_ = supabase.DeleteFileIfPresent(s.storage, oldImageURL)
	}

	return &model.AdminUpdateTitleResponse{
		Title: mapAdminTitleResponse(*title),
	}, nil
}

func (s *TitleService) DeleteTitleByAdmin(titleID uuid.UUID) (*model.AdminDeleteTitleResponse, error) {
	if titleID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid title id")
	}

	if _, err := s.titleRepo.GetTitle(s.db, model.GetTitleParam{TitleID: titleID}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("title not found")
		}
		return nil, appErrors.InternalServer("failed to get title")
	}

	ownerships, err := s.titleRepo.CountTitleOwnerships(s.db, titleID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to check title ownership")
	}

	if ownerships > 0 {
		return nil, appErrors.Conflict("granted title cannot be deleted; set it inactive instead")
	}

	err = s.titleRepo.DeleteTitle(s.db, titleID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete title")
	}

	return &model.AdminDeleteTitleResponse{
		TitleID: titleID,
	}, nil
}

func (s *TitleService) ListTitlesForUser(userID uuid.UUID) (*model.ListUserTitlesResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	tx := s.db.Begin()
	defer tx.Rollback()
	profile, err := s.userProfileRepo.GetUserProfileForUpdate(tx, model.GetUserProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user profile not found")
		}
		return nil, appErrors.InternalServer("failed to get user profile")
	}

	currentLevel := max(profile.CurrentLevel, 1)
	err = grantUnlockedTitles(tx, s.titleRepo, s.userItemRepo, userID, currentLevel, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	rows, err := s.titleRepo.ListTitlesForUser(tx, userID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to list titles")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.ListUserTitlesResponse{
		Titles: mapUserTitleResponses(rows),
	}, nil
}

func (s *TitleService) EquipTitleForUser(userID uuid.UUID, titleID uuid.UUID) (*model.EquipTitleResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if titleID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid title id")
	}

	now := time.Now().UTC()
	tx := s.db.Begin()
	defer tx.Rollback()
	profile, err := s.userProfileRepo.GetUserProfileForUpdate(tx, model.GetUserProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user profile not found")
		}
		return nil, appErrors.InternalServer("failed to get user profile")
	}

	title, err := s.titleRepo.GetTitleForUpdate(tx, titleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("title not found")
		}
		return nil, appErrors.InternalServer("failed to get title")
	}

	if title.Status != model.TitleStatusActive {
		return nil, appErrors.Conflict("title is not available")
	}

	currentLevel := max(profile.CurrentLevel, 1)
	if currentLevel < title.UnlockLevel {
		return nil, appErrors.Conflict("title level requirement not met")
	}

	err = grantUnlockedTitles(tx, s.titleRepo, s.userItemRepo, userID, currentLevel, now)
	if err != nil {
		return nil, err
	}

	userItem, err := s.userItemRepo.GetUserItem(tx, model.GetUserItemParam{UserID: userID, TitleID: titleID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get title ownership")
	}

	err = s.userItemRepo.ClearEquippedTitlesByUser(tx, userID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to clear equipped title")
	}

	userItem.EquippedAt = &now
	err = s.userItemRepo.UpdateUserItem(tx, userItem)
	if err != nil {
		return nil, appErrors.InternalServer("failed to equip title")
	}

	profile.TitleID = &title.TitleID
	profile.Title = title.Title
	err = s.userProfileRepo.UpdateUserProfile(tx, profile)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update user profile")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.EquipTitleResponse{
		Title: model.UserTitleResponse{
			TitleID:     title.TitleID,
			Title:       title.Title,
			UnlockLevel: title.UnlockLevel,
			ImageBorder: title.ImageBorder,
			IsOwned:     true,
			IsEquipped:  true,
			CanEquip:    false,
		}}, nil
}

func grantUnlockedTitles(tx *gorm.DB, titleRepo repository.ITitleRepository, userItemRepo repository.IUserItemRepository, userID uuid.UUID, currentLevel int, now time.Time) error {
	titles, err := titleRepo.ListActiveTitles(tx)
	if err != nil {
		return appErrors.InternalServer("failed to list titles")
	}
	for _, title := range titles {
		if title.UnlockLevel > currentLevel {
			continue
		}
		_, err := userItemRepo.GetUserItem(tx, model.GetUserItemParam{UserID: userID, TitleID: title.TitleID})
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.InternalServer("failed to get title ownership")
		}
		if err := userItemRepo.CreateUserItem(tx, &entity.UserItem{
			UserItemID:   uuid.New(),
			UserID:       userID,
			TitleID:      &title.TitleID,
			PurchaseType: model.UserItemPurchaseTypeGrant,
			PurchasedAt:  now,
		}); err != nil {
			return appErrors.InternalServer("failed to grant title")
		}
	}
	return nil
}

func mapUserTitleResponses(rows []model.UserTitleRow) []model.UserTitleResponse {
	items := make([]model.UserTitleResponse, 0, len(rows))
	for _, row := range rows {
		isOwned := row.UserItemID != nil
		isEquipped := row.CurrentTitleID != nil && *row.CurrentTitleID == row.TitleID
		items = append(items, model.UserTitleResponse{
			TitleID:     row.TitleID,
			Title:       row.Title,
			UnlockLevel: row.UnlockLevel,
			ImageBorder: row.ImageBorder,
			IsOwned:     isOwned,
			IsEquipped:  isEquipped,
			CanEquip:    isOwned && !isEquipped,
		})
	}
	return items
}

func (s *TitleService) uploadTitleImage(file *multipart.FileHeader) (string, error) {
	url, err := supabase.UploadOptionalImage(s.storage, file, 2*1024*1024, "title image size exceeds 2MB limit")
	if err != nil {
		return "", appErrors.BadRequest(err.Error())
	}
	return url, nil
}

func newTitleEntity(name string, unlockLevel int, status string) (*entity.Title, error) {
	name = strings.TrimSpace(name)
	if name == "" || unlockLevel < 1 {
		return nil, appErrors.BadRequest("title and unlock level are required")
	}
	if status == "" {
		status = model.TitleStatusActive
	}
	if status != model.TitleStatusActive && status != model.TitleStatusInactive {
		return nil, appErrors.BadRequest("invalid title status")
	}
	return &entity.Title{
		TitleID:     uuid.New(),
		Title:       name,
		UnlockLevel: unlockLevel,
		Status:      status,
	}, nil
}

func mapAdminTitleResponses(titles []entity.Title) []model.AdminTitleResponse {
	items := make([]model.AdminTitleResponse, 0, len(titles))
	for _, title := range titles {
		items = append(items, mapAdminTitleResponse(title))
	}
	return items
}

func mapAdminTitleResponse(title entity.Title) model.AdminTitleResponse {
	return model.AdminTitleResponse{
		TitleID:     title.TitleID,
		Title:       title.Title,
		UnlockLevel: title.UnlockLevel,
		ImageBorder: title.ImageBorder,
		Status:      title.Status,
		CreatedAt:   title.CreatedAt,
		UpdatedAt:   title.UpdatedAt,
	}
}

func normalizeTitlePagination(page int, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
