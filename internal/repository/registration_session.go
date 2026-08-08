package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IRegistrationSessionRepository interface {
	GetRegistrationSession(tx *gorm.DB, param model.GetRegistrationSessionParam) (*entity.RegistrationSession, error)
	GetRegistrationSessionForUpdate(tx *gorm.DB, sessionTokenHash string) (*entity.RegistrationSession, error)
	CreateRegistrationSession(tx *gorm.DB, session *entity.RegistrationSession) error
	UpdateRegistrationSession(tx *gorm.DB, session *entity.RegistrationSession) error
	MarkRegistrationSessionCompleted(tx *gorm.DB, sessionID uuid.UUID) error
	DeleteExpiredRegistrationSessions(tx *gorm.DB) error
}

type RegistrationSessionRepository struct {
	db *gorm.DB
}

func NewRegistrationSessionRepository(db *gorm.DB) IRegistrationSessionRepository {
	return &RegistrationSessionRepository{db: db}
}

func (r *RegistrationSessionRepository) GetRegistrationSession(tx *gorm.DB, param model.GetRegistrationSessionParam) (*entity.RegistrationSession, error) {
	var session entity.RegistrationSession
	query := tx.Where(&param)

	err := query.First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RegistrationSessionRepository) GetRegistrationSessionForUpdate(tx *gorm.DB, sessionTokenHash string) (*entity.RegistrationSession, error) {
	var session entity.RegistrationSession

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_token_hash = ?", sessionTokenHash).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RegistrationSessionRepository) CreateRegistrationSession(tx *gorm.DB, session *entity.RegistrationSession) error {
	err := tx.Create(session).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *RegistrationSessionRepository) UpdateRegistrationSession(tx *gorm.DB, session *entity.RegistrationSession) error {
	err := tx.Save(session).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *RegistrationSessionRepository) MarkRegistrationSessionCompleted(tx *gorm.DB, sessionID uuid.UUID) error {
	err := tx.Model(&entity.RegistrationSession{}).
		Where("registration_session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"current_step": model.RegisterStepCompleted,
			"completed_at": gorm.Expr("UTC_TIMESTAMP()"),
		}).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RegistrationSessionRepository) DeleteExpiredRegistrationSessions(tx *gorm.DB) error {
	err := tx.
		Where("expires_at < UTC_TIMESTAMP() AND completed_at IS NULL").
		Delete(&entity.RegistrationSession{}).Error
	if err != nil {
		return err
	}
	return nil
}
