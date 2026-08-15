package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ITitleRepository interface {
	GetTitle(tx *gorm.DB, param model.GetTitleParam) (*entity.Title, error)
	GetTitleForUpdate(tx *gorm.DB, titleID uuid.UUID) (*entity.Title, error)
	ListActiveTitles(tx *gorm.DB) ([]entity.Title, error)
	ListTitlesForUser(tx *gorm.DB, userID uuid.UUID) ([]model.UserTitleRow, error)
	CreateTitle(tx *gorm.DB, title *entity.Title) error
	ListTitles(tx *gorm.DB, search string, limit int, offset int) ([]entity.Title, int64, error)
	UpdateTitle(tx *gorm.DB, title *entity.Title) error
	DeleteTitle(tx *gorm.DB, titleID uuid.UUID) error
	CountTitleOwnerships(tx *gorm.DB, titleID uuid.UUID) (int64, error)
}

type TitleRepository struct {
	db *gorm.DB
}

func NewTitleRepository(db *gorm.DB) ITitleRepository {
	return &TitleRepository{db: db}
}

func (r *TitleRepository) CreateTitle(tx *gorm.DB, title *entity.Title) error {
	err := tx.Create(title).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *TitleRepository) ListTitles(tx *gorm.DB, search string, limit int, offset int) ([]entity.Title, int64, error) {
	var titles []entity.Title
	var total int64
	query := tx.Model(&entity.Title{})
	if search != "" {
		query = query.Where("title LIKE ?", "%"+search+"%")
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = query.Order("unlock_level ASC").Order("created_at ASC").Limit(limit).Offset(offset).Find(&titles).Error
	if err != nil {
		return nil, 0, err
	}
	return titles, total, nil
}

func (r *TitleRepository) UpdateTitle(tx *gorm.DB, title *entity.Title) error {
	err := tx.Save(title).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *TitleRepository) DeleteTitle(tx *gorm.DB, titleID uuid.UUID) error {
	err := tx.Where("title_id = ?", titleID).Delete(&entity.Title{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *TitleRepository) CountTitleOwnerships(tx *gorm.DB, titleID uuid.UUID) (int64, error) {
	var count int64
	err := tx.Model(&entity.UserItem{}).Where("title_id = ?", titleID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TitleRepository) GetTitle(tx *gorm.DB, param model.GetTitleParam) (*entity.Title, error) {
	var title entity.Title
	err := tx.Where("title_id = ?", param.TitleID).First(&title).Error
	if err != nil {
		return nil, err
	}
	return &title, nil
}

func (r *TitleRepository) GetTitleForUpdate(tx *gorm.DB, titleID uuid.UUID) (*entity.Title, error) {
	var title entity.Title
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("title_id = ?", titleID).First(&title).Error
	if err != nil {
		return nil, err
	}
	return &title, nil
}

func (r *TitleRepository) ListActiveTitles(tx *gorm.DB) ([]entity.Title, error) {
	var titles []entity.Title
	err := tx.Where("status = ?", model.TitleStatusActive).Order("unlock_level ASC").Order("created_at ASC").Find(&titles).Error
	if err != nil {
		return nil, err
	}
	return titles, nil
}

func (r *TitleRepository) ListTitlesForUser(tx *gorm.DB, userID uuid.UUID) ([]model.UserTitleRow, error) {
	var rows []model.UserTitleRow
	err := tx.Table("titles").
		Joins("LEFT JOIN user_items ON user_items.title_id = titles.title_id AND user_items.user_id = ?", userID).
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = ?", userID).
		Where("titles.status = ?", model.TitleStatusActive).
		Select(`
			titles.title_id,
			titles.title,
			titles.unlock_level,
			titles.image_border,
			titles.status,
			user_items.user_item_id,
			user_items.equipped_at,
			user_profiles.title_id AS current_title_id
		`).
		Order("titles.unlock_level ASC").
		Order("titles.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
