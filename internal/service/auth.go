package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/bcrypt"
	constants "github.com/azmiagr/unesco-hackathon/pkg/constant"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/jwt"
	"github.com/azmiagr/unesco-hackathon/pkg/mail"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	registerSessionTTL = 24 * time.Hour
	registerOtpTTL     = 10 * time.Minute
	adminLoginOtpTTL   = 10 * time.Minute
)

type IAuthService interface {
	StartRegister(req model.StartRegisterRequest) (*model.RegisterAuthResult, error)
	VerifyRegisterOtp(sessionToken string, req model.VerifyRegisterOtpRequest) (*model.RegisterAuthResult, error)
	SelectRegisterAvatar(sessionToken string, req model.SelectRegisterAvatarRequest) (*model.RegisterAuthResult, error)
	CompleteRegisterProfile(sessionToken string, req model.CompleteRegisterProfileRequest) (*model.CompleteRegisterResult, error)
	GetRegisterSession(sessionToken string) (*model.RegisterSessionResponse, error)
	Login(req model.LoginRequest) (*model.LoginResponse, error)
	VerifyAdminLoginOtp(sessionToken string, req model.VerifyAdminLoginOtpRequest) (*model.LoginResponse, error)
}

type AuthService struct {
	db              *gorm.DB
	userRepo        repository.IUserRepository
	avatarRepo      repository.IAvatarRepository
	userProfileRepo repository.IUserProfileRepository
	sessionRepo     repository.IRegistrationSessionRepository
	roleRepo        repository.IRoleRepository
	jwtAuth         jwt.Interface
	bcrypt          bcrypt.Interface
}

func NewAuthService(
	userRepo repository.IUserRepository,
	avatarRepo repository.IAvatarRepository,
	userProfileRepo repository.IUserProfileRepository,
	sessionRepo repository.IRegistrationSessionRepository,
	roleRepo repository.IRoleRepository,
	jwtAuth jwt.Interface,
	bcrypt bcrypt.Interface,
) IAuthService {
	return &AuthService{
		db:              mariadb.Connection,
		userRepo:        userRepo,
		avatarRepo:      avatarRepo,
		userProfileRepo: userProfileRepo,
		sessionRepo:     sessionRepo,
		roleRepo:        roleRepo,
		jwtAuth:         jwtAuth,
		bcrypt:          bcrypt,
	}
}

func (s *AuthService) StartRegister(req model.StartRegisterRequest) (*model.RegisterAuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	_, err := s.userRepo.GetUser(s.db, model.GetUserParam{
		Email: email,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return nil, appErrors.Conflict("email already exists")
	}

	passwordHash, err := s.bcrypt.GenerateFromPassword(req.Password)
	if err != nil {
		return nil, appErrors.InternalServer("failed to hash password")
	}

	sessionToken, err := generateSecureToken(32)
	if err != nil {
		return nil, appErrors.InternalServer("failed to generate session token")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	otpCode := mail.GenerateCode()
	now := time.Now().UTC()
	session := &entity.RegistrationSession{
		RegistrationSessionID: uuid.New(),
		SessionTokenHash:      hashString(sessionToken),
		Email:                 email,
		PasswordHash:          passwordHash,
		OtpCodeHash:           hashString(otpCode),
		OtpSentAt:             &now,
		OtpExpiresAt:          ptrTime(now.Add(registerOtpTTL)),
		CurrentStep:           model.RegisterStepEmailSubmitted,
		ExpiresAt:             now.Add(registerSessionTTL),
	}

	existing, err := s.sessionRepo.GetRegistrationSession(tx, model.GetRegistrationSessionParam{
		Email: email,
	})
	if err == nil && existing.CompletedAt == nil && existing.ExpiresAt.After(now) {
		existing.SessionTokenHash = session.SessionTokenHash
		existing.PasswordHash = session.PasswordHash
		existing.OtpCodeHash = session.OtpCodeHash
		existing.OtpSentAt = session.OtpSentAt
		existing.OtpExpiresAt = session.OtpExpiresAt
		existing.EmailVerifiedAt = nil
		existing.AvatarID = nil
		existing.Title = ""
		existing.CurrentStep = model.RegisterStepEmailSubmitted
		existing.ExpiresAt = session.ExpiresAt
		session = existing
	}

	err = s.sessionRepo.CreateRegistrationSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create registration session")
	}

	err = mail.SendVerificationEmail(email, email, otpCode, "")
	if err != nil {
		return nil, appErrors.InternalServer("failed to send verification email")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.RegisterAuthResult{
		SessionToken: sessionToken,
		State:        mapRegisterSession(session),
	}, nil
}

func (s *AuthService) VerifyRegisterOtp(sessionToken string, req model.VerifyRegisterOtpRequest) (*model.RegisterAuthResult, error) {
	var session *entity.RegistrationSession

	tx := s.db.Begin()
	defer tx.Rollback()

	lockedSession, err := s.getValidSessionForUpdate(tx, sessionToken)
	if err != nil {
		return nil, err
	}

	if lockedSession.CurrentStep != model.RegisterStepEmailSubmitted {
		return nil, appErrors.Conflict("registration step is not waiting for otp verification")
	}

	if lockedSession.OtpExpiresAt == nil || lockedSession.OtpExpiresAt.Before(time.Now().UTC()) {
		return nil, appErrors.BadRequest("registration session expired")
	}

	if lockedSession.OtpCodeHash != hashString(req.Code) {
		return nil, appErrors.BadRequest("invalid otp code")
	}

	now := time.Now().UTC()
	lockedSession.EmailVerifiedAt = &now
	lockedSession.CurrentStep = model.RegisterStepEmailVerified
	lockedSession.OtpCodeHash = ""
	session = lockedSession

	err = s.sessionRepo.UpdateRegistrationSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update registration session")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.RegisterAuthResult{
		SessionToken: sessionToken,
		State:        mapRegisterSession(session),
	}, nil
}

func (s *AuthService) SelectRegisterAvatar(sessionToken string, req model.SelectRegisterAvatarRequest) (*model.RegisterAuthResult, error) {
	var session *entity.RegistrationSession

	tx := s.db.Begin()
	defer tx.Rollback()

	lockedSession, err := s.getValidSessionForUpdate(tx, sessionToken)
	if err != nil {
		return nil, err
	}

	if lockedSession.CurrentStep != model.RegisterStepEmailVerified {
		return nil, appErrors.Conflict("registration step is not waiting for avatar selection")
	}

	avatar, err := s.avatarRepo.GetAvatar(tx, model.GetAvatarParam{
		AvatarID: req.AvatarID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Conflict("avatar not found")
		}
		return nil, err
	}

	if avatar.Status != "active" || avatar.UnlockLevel > 0 {
		return nil, appErrors.Forbidden("avatar is not available for registration")
	}

	lockedSession.AvatarID = &req.AvatarID
	lockedSession.CurrentStep = model.RegisterStepAvatarSelected
	session = lockedSession

	err = s.sessionRepo.UpdateRegistrationSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update registration session")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.RegisterAuthResult{
		SessionToken: sessionToken,
		State:        mapRegisterSession(session),
	}, nil
}

func (s *AuthService) CompleteRegisterProfile(sessionToken string, req model.CompleteRegisterProfileRequest) (*model.CompleteRegisterResult, error) {
	username := strings.TrimSpace(req.Username)
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = model.DefaultRegisterTitle
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	session, err := s.getValidSessionForUpdate(tx, sessionToken)
	if err != nil {
		return nil, err
	}

	if session.CurrentStep != model.RegisterStepAvatarSelected {
		return nil, appErrors.Conflict("registration step is not waiting for profile completion")
	}

	if session.AvatarID == nil {
		return nil, appErrors.BadRequest("avatar must be selected")
	}

	_, err = s.userRepo.GetUser(tx, model.GetUserParam{
		Email: session.Email,
	})
	if err == nil {
		return nil, appErrors.Conflict("email already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to check email existence")
	}

	_, err = s.userRepo.GetUser(tx, model.GetUserParam{
		Username: username,
	})
	if err == nil {
		return nil, appErrors.Conflict("username already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to check username existence")
	}

	role, err := s.roleRepo.GetRoleByName(tx, constants.RoleUser)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get user role")
	}

	user := &entity.User{
		UserID:   uuid.New(),
		RoleID:   role.RoleID,
		Username: username,
		Email:    session.Email,
		Password: session.PasswordHash,
		Status:   "active",
	}

	err = s.userRepo.CreateUser(tx, user)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create user")
	}

	profile := &entity.UserProfile{
		UserProfileID:     uuid.New(),
		UserID:            user.UserID,
		AvatarID:          *session.AvatarID,
		Title:             title,
		CurrentLevel:      1,
		CurrentXP:         0,
		AuditorReputation: 100,
	}

	err = s.userProfileRepo.CreateUserProfile(tx, profile)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create user profile")
	}

	now := time.Now().UTC()
	session.Title = title
	session.CurrentStep = model.RegisterStepCompleted
	session.CompletedAt = &now

	err = s.sessionRepo.UpdateRegistrationSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update registration session")
	}

	token, err := s.jwtAuth.CreateJWTToken(user.UserID, role.RoleName)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create jwt token")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.CompleteRegisterResult{
		Token: token,
	}, nil
}

func (s *AuthService) GetRegisterSession(sessionToken string) (*model.RegisterSessionResponse, error) {
	sessionToken = strings.TrimSpace(sessionToken)
	if sessionToken == "" {
		return nil, appErrors.Unauthorized("missing registration session token")
	}

	session, err := s.sessionRepo.GetRegistrationSession(s.db, model.GetRegistrationSessionParam{
		SessionTokenHash: hashString(sessionToken),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Unauthorized("invalid registration session token")
		}
		return nil, err
	}

	if session.CompletedAt != nil {
		return nil, appErrors.Conflict("registration already completed")
	}

	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, appErrors.Unauthorized("registration session expired")
	}

	response := mapRegisterSession(session)
	return &response, nil
}

func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.GetUser(s.db, model.GetUserParam{
		Email: email,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Unauthorized("invalid email or password")
		}
		return nil, appErrors.InternalServer("failed to get user")
	}

	if user.Status != "active" {
		return nil, appErrors.Unauthorized("account is not active")
	}

	err = s.bcrypt.CompareAndHashPassword(user.Password, req.Password)
	if err != nil {
		return nil, appErrors.Unauthorized("invalid email or password")
	}

	role, err := s.roleRepo.GetRole(s.db, user.RoleID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get role")
	}

	if role.RoleName == constants.RoleAdmin {
		return s.startAdminLoginOtp(user)
	}

	token, err := s.jwtAuth.CreateJWTToken(user.UserID, role.RoleName)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create jwt token")
	}

	return &model.LoginResponse{
		Token:       token,
		RequiresOtp: false,
	}, nil
}

func (s *AuthService) VerifyAdminLoginOtp(sessionToken string, req model.VerifyAdminLoginOtpRequest) (*model.LoginResponse, error) {
	sessionToken = strings.TrimSpace(sessionToken)
	if sessionToken == "" {
		return nil, appErrors.Unauthorized("missing admin login session token")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	session, err := s.userRepo.GetAdminLoginOtpSessionForUpdate(tx, model.GetAdminLoginOtpSessionParam{
		SessionTokenHash: hashString(sessionToken),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Unauthorized("invalid admin login session token")
		}
		return nil, appErrors.InternalServer("failed to get admin login otp session")
	}

	if session.RevokedAt != nil {
		return nil, appErrors.Unauthorized("admin login session revoked")
	}
	if session.VerifiedAt != nil {
		return nil, appErrors.Conflict("admin login session already verified")
	}
	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, appErrors.Unauthorized("admin login otp expired")
	}
	if session.OtpCodeHash != hashString(req.Code) {
		return nil, appErrors.BadRequest("invalid otp code")
	}

	user, err := s.userRepo.GetUser(tx, model.GetUserParam{
		UserID: session.UserID,
	})
	if err != nil {
		return nil, appErrors.Unauthorized("admin user not found")
	}
	if user.Status != "active" {
		return nil, appErrors.Unauthorized("account is not active")
	}

	role, err := s.roleRepo.GetRole(tx, user.RoleID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get role")
	}
	if role.RoleName != constants.RoleAdmin {
		return nil, appErrors.Forbidden("user is not admin")
	}

	now := time.Now().UTC()
	session.VerifiedAt = &now
	session.OtpCodeHash = ""
	if err := s.userRepo.UpdateAdminLoginOtpSession(tx, session); err != nil {
		return nil, appErrors.InternalServer("failed to verify admin login otp session")
	}

	token, err := s.jwtAuth.CreateJWTToken(user.UserID, role.RoleName)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create jwt token")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:       token,
		RequiresOtp: false,
	}, nil
}

func (s *AuthService) startAdminLoginOtp(user *entity.User) (*model.LoginResponse, error) {
	sessionToken, err := generateSecureToken(32)
	if err != nil {
		return nil, appErrors.InternalServer("failed to generate session token")
	}

	otpCode := mail.GenerateCode()
	now := time.Now().UTC()
	session := &entity.AdminLoginOtpSession{
		AdminLoginOtpSessionID: uuid.New(),
		UserID:                 user.UserID,
		SessionTokenHash:       hashString(sessionToken),
		OtpCodeHash:            hashString(otpCode),
		ExpiresAt:              now.Add(adminLoginOtpTTL),
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	err = s.userRepo.RevokeActiveAdminLoginOtpSessions(tx, user.UserID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to revoke previous admin login otp")
	}

	err = s.userRepo.CreateAdminLoginOtpSession(tx, session)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create admin login otp session")
	}

	err = mail.SendAdminLoginOtpEmail(user.Email, user.Email, otpCode)
	if err != nil {
		return nil, appErrors.InternalServer("failed to send admin login otp")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		RequiresOtp:  true,
		Email:        user.Email,
		SessionToken: sessionToken,
	}, nil
}

func (s *AuthService) getValidSessionForUpdate(tx *gorm.DB, sessionToken string) (*entity.RegistrationSession, error) {
	sessionToken = strings.TrimSpace(sessionToken)
	if sessionToken == "" {
		return nil, appErrors.Unauthorized("missing registration session token")
	}

	session, err := s.sessionRepo.GetRegistrationSessionForUpdate(tx, hashString(sessionToken))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Unauthorized("invalid registration session token")
		}
		return nil, err
	}

	if session.CompletedAt != nil {
		return nil, appErrors.Conflict("registration already completed")
	}

	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, appErrors.Unauthorized("registration session expired")
	}

	return session, nil
}

func mapRegisterSession(session *entity.RegistrationSession) model.RegisterSessionResponse {
	return model.RegisterSessionResponse{
		Email:       session.Email,
		CurrentStep: session.CurrentStep,
		AvatarID:    session.AvatarID,
		Title:       session.Title,
		ExpiresAt:   session.ExpiresAt,
	}
}

func generateSecureToken(byteLength int) (string, error) {
	token := make([]byte, byteLength)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
