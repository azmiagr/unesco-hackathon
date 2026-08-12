package service

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/bcrypt"
	constants "github.com/azmiagr/unesco-hackathon/pkg/constant"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultAdminUserLimit = 10
	maxAdminUserLimit     = 100
)

var allowedUserStatuses = map[string]bool{
	"active":    true,
	"inactive":  true,
	"suspended": true,
	"banned":    true,
}

var allowedAdminRoles = map[string]bool{
	constants.RoleUser:  true,
	constants.RoleAdmin: true,
}

type IUserService interface {
	GetUser(param model.GetUserParam) (*entity.User, error)
	GetUserRoleName(user *entity.User) (string, error)
	ListUsers(req model.AdminListUsersRequest) (*model.AdminListUsersResponse, error)
	GetUserDetail(userID uuid.UUID) (*model.AdminUserDetailResponse, error)
	UpdateUserAccess(adminUserID uuid.UUID, targetUserID uuid.UUID, req model.AdminUpdateUserAccessRequest) (*model.AdminUpdateUserAccessResponse, error)
	HardDeleteUser(adminUserID uuid.UUID, targetUserID uuid.UUID) (*model.AdminDeleteUserResponse, error)
	CreateUserByAdmin(adminUserID uuid.UUID, req model.AdminCreateUserRequest) (*model.AdminCreateUserResponse, error)
	UpdateUserByAdmin(adminUserID uuid.UUID, targetUserID uuid.UUID, req model.AdminUpdateUserRequest) (*model.AdminUpdateUserResponse, error)
}

type UserService struct {
	db              *gorm.DB
	userRepo        repository.IUserRepository
	roleRepo        repository.IRoleRepository
	userProfileRepo repository.IUserProfileRepository
	auditLogRepo    repository.IAuditLogRepository
	bcrypt          bcrypt.Interface
}

func NewUserService(
	userRepo repository.IUserRepository,
	roleRepo repository.IRoleRepository,
	userProfileRepo repository.IUserProfileRepository,
	auditLogRepo repository.IAuditLogRepository,
	bcrypt bcrypt.Interface,
) IUserService {
	return &UserService{
		db:              mariadb.Connection,
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		userProfileRepo: userProfileRepo,
		auditLogRepo:    auditLogRepo,
		bcrypt:          bcrypt,
	}
}

func (s *UserService) GetUser(param model.GetUserParam) (*entity.User, error) {
	return s.userRepo.GetUser(s.db, param)
}

func (s *UserService) GetUserRoleName(user *entity.User) (string, error) {
	role, err := s.roleRepo.GetRole(s.db, user.RoleID)
	if err != nil {
		return "", err
	}

	return role.RoleName, nil
}

func (s *UserService) ListUsers(req model.AdminListUsersRequest) (*model.AdminListUsersResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultAdminUserLimit
	}
	if limit > maxAdminUserLimit {
		limit = maxAdminUserLimit
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role != "" && !allowedAdminRoles[role] {
		return nil, appErrors.BadRequest("invalid role filter")
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "" && !allowedUserStatuses[status] {
		return nil, appErrors.BadRequest("invalid status filter")
	}

	param := model.AdminListUsersParam{
		Search: strings.TrimSpace(req.Search),
		Role:   role,
		Status: status,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	users, total, err := s.userRepo.ListUsers(s.db, param)
	if err != nil {
		return nil, appErrors.InternalServer("failed to list users")
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.AdminListUsersResponse{
		Users: users,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *UserService) GetUserDetail(userID uuid.UUID) (*model.AdminUserDetailResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid user id")
	}

	user, err := s.userRepo.GetUserDetail(s.db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user not found")
		}
		return nil, appErrors.InternalServer("failed to get user detail")
	}

	return &model.AdminUserDetailResponse{
		User: *user,
		RecentProgress: model.AdminUserRecentProgressResponse{
			Items: []any{},
		},
	}, nil
}

func (s *UserService) UpdateUserAccess(adminUserID uuid.UUID, targetUserID uuid.UUID, req model.AdminUpdateUserAccessRequest) (*model.AdminUpdateUserAccessResponse, error) {
	if targetUserID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid user id")
	}

	if adminUserID == targetUserID {
		return nil, appErrors.Forbidden("admin cannot update own access")
	}

	roleName := ""
	if req.RoleName != nil {
		roleName = strings.ToLower(strings.TrimSpace(*req.RoleName))
		if roleName == "" || !allowedAdminRoles[roleName] {
			return nil, appErrors.BadRequest("invalid role")
		}
	}

	status := ""
	if req.Status != nil {
		status = strings.ToLower(strings.TrimSpace(*req.Status))
		if status == "" || !allowedUserStatuses[status] {
			return nil, appErrors.BadRequest("invalid status")
		}
	}

	if roleName == "" && status == "" {
		return nil, appErrors.BadRequest("role_name or status is required")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	user, err := s.userRepo.GetUserForUpdate(tx, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user not found")
		}
		return nil, appErrors.InternalServer("failed to get user")
	}

	currentRole, err := s.roleRepo.GetRole(tx, user.RoleID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get current role")
	}
	before := newAuditUserSnapshot(user, currentRole.RoleName)

	roleID := uuid.Nil
	finalRoleName := ""
	if roleName != "" {
		role, err := s.roleRepo.GetRoleByName(tx, roleName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, appErrors.BadRequest("role not found")
			}
			return nil, appErrors.InternalServer("failed to get role")
		}

		roleID = role.RoleID
		finalRoleName = role.RoleName
	} else {
		finalRoleName = currentRole.RoleName
	}

	finalStatus := user.Status
	if status != "" {
		finalStatus = status
	}

	err = s.userRepo.UpdateUserAccess(tx, model.AdminUpdateUserAccessParam{
		UserID: user.UserID,
		RoleID: roleID,
		Status: status,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to update user access")
	}

	after := auditUserSnapshot{
		UserID:   user.UserID,
		RoleID:   user.RoleID,
		RoleName: finalRoleName,
		Username: user.Username,
		Email:    user.Email,
		Status:   finalStatus,
	}
	if roleID != uuid.Nil {
		after.RoleID = roleID
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionUpdate,
		Module:        model.AuditModuleUsers,
		TargetType:    "user",
		TargetID:      user.UserID.String(),
		TargetLabel:   user.Email,
		Detail:        fmt.Sprintf("Updated user access for %s", user.Email),
		PayloadBefore: before,
		PayloadAfter:  after,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.AdminUpdateUserAccessResponse{
		UserID:   user.UserID,
		RoleName: finalRoleName,
		Status:   finalStatus,
	}, nil
}

func (s *UserService) HardDeleteUser(adminUserID uuid.UUID, targetUserID uuid.UUID) (*model.AdminDeleteUserResponse, error) {
	if targetUserID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid user id")
	}

	if adminUserID == targetUserID {
		return nil, appErrors.Forbidden("admin cannot delete own account")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	user, err := s.userRepo.GetUserForUpdate(tx, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user not found")
		}
		return nil, appErrors.InternalServer("failed to get user")
	}

	role, err := s.roleRepo.GetRole(tx, user.RoleID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get current role")
	}
	before := newAuditUserSnapshot(user, role.RoleName)

	err = s.userRepo.HardDeleteUser(tx, user.UserID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete user")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionDelete,
		Module:        model.AuditModuleUsers,
		TargetType:    "user",
		TargetID:      user.UserID.String(),
		TargetLabel:   user.Email,
		Detail:        fmt.Sprintf("Hard deleted user %s", user.Email),
		PayloadBefore: before,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.AdminDeleteUserResponse{
		UserID: user.UserID,
	}, nil

}

func (s *UserService) CreateUserByAdmin(adminUserID uuid.UUID, req model.AdminCreateUserRequest) (*model.AdminCreateUserResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	roleName := strings.ToLower(strings.TrimSpace(req.RoleName))
	status := strings.ToLower(strings.TrimSpace(req.Status))

	username, err := helper.RequireTrimmedString(req.Username, "username is required")
	if err != nil {
		return nil, err
	}

	email, err := helper.RequireTrimmedString(req.Email, "email is required")
	if err != nil {
		return nil, err
	}
	email = strings.ToLower(email)

	if req.Password != req.PasswordConfirmation {
		return nil, appErrors.BadRequest("password confirmation does not match")
	}

	if roleName == "" || !allowedAdminRoles[roleName] {
		return nil, appErrors.BadRequest("invalid role")
	}

	if status == "" || !allowedUserStatuses[status] {
		return nil, appErrors.BadRequest("invalid status")
	}

	passwordHash, err := s.bcrypt.GenerateFromPassword(req.Password)
	if err != nil {
		return nil, appErrors.InternalServer("failed to hash password")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	emailExists, err := s.userRepo.UserExists(tx, model.GetUserParam{
		Email: email,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to check email existence")
	}
	if emailExists {
		return nil, appErrors.Conflict("email already exists")
	}

	usernameExists, err := s.userRepo.UserExists(tx, model.GetUserParam{
		Username: username,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to check username existence")
	}
	if usernameExists {
		return nil, appErrors.Conflict("username already exists")
	}

	role, err := s.roleRepo.GetRoleByName(tx, roleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("role not found")
		}
		return nil, appErrors.InternalServer("failed to get role")
	}

	user := &entity.User{
		UserID:   uuid.New(),
		RoleID:   role.RoleID,
		Username: username,
		Email:    email,
		Password: passwordHash,
		Status:   status,
	}

	err = s.userRepo.CreateUser(tx, user)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create user")
	}

	profile := &entity.UserProfile{
		UserProfileID:     uuid.New(),
		UserID:            user.UserID,
		Title:             model.DefaultRegisterTitle,
		CurrentLevel:      1,
		CurrentXP:         0,
		AuditorReputation: 100,
	}

	err = s.userProfileRepo.CreateUserProfile(tx, profile)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create user profile")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID: adminUserID,
		ActionType:   model.AuditActionCreate,
		Module:       model.AuditModuleUsers,
		TargetType:   "user",
		TargetID:     user.UserID.String(),
		TargetLabel:  user.Email,
		Detail:       fmt.Sprintf("Created user %s", user.Email),
		PayloadAfter: newAuditUserSnapshot(user, role.RoleName),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	createdUser, err := s.userRepo.GetUserDetail(s.db, user.UserID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get user detail")
	}

	return &model.AdminCreateUserResponse{
		User: *createdUser,
	}, nil
}

func (s *UserService) UpdateUserByAdmin(adminUserID uuid.UUID, targetUserID uuid.UUID, req model.AdminUpdateUserRequest) (*model.AdminUpdateUserResponse, error) {
	if targetUserID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid user id")
	}

	if adminUserID == targetUserID {
		return nil, appErrors.Forbidden("admin cannot update own account")
	}

	roleName := strings.ToLower(strings.TrimSpace(req.RoleName))
	status := strings.ToLower(strings.TrimSpace(req.Status))
	password := strings.TrimSpace(req.Password)
	passwordConfirmation := strings.TrimSpace(req.PasswordConfirmation)

	email, err := helper.RequireTrimmedString(req.Email, "email is required")
	if err != nil {
		return nil, err
	}
	email = strings.ToLower(email)

	username, err := helper.RequireTrimmedString(req.Username, "username is required")
	if err != nil {
		return nil, err
	}

	if roleName == "" || !allowedAdminRoles[roleName] {
		return nil, appErrors.BadRequest("invalid role")
	}

	if status == "" || !allowedUserStatuses[status] {
		return nil, appErrors.BadRequest("invalid status")
	}

	if password != "" {
		if len(password) < 8 {
			return nil, appErrors.BadRequest("password must be at least 8 characters")
		}

		if password != passwordConfirmation {
			return nil, appErrors.BadRequest("password confirmation does not match")
		}
	} else if passwordConfirmation != "" {
		return nil, appErrors.BadRequest("password is required when password confirmation is provided")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	user, err := s.userRepo.GetUserForUpdate(tx, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user not found")
		}
		return nil, appErrors.InternalServer("failed to get user")
	}

	currentRole, err := s.roleRepo.GetRole(tx, user.RoleID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get current role")
	}
	before := newAuditUserSnapshot(user, currentRole.RoleName)

	if user.Email != email {
		emailExists, err := s.userRepo.UserExists(tx, model.GetUserParam{
			Email: email,
		})
		if err != nil {
			return nil, appErrors.InternalServer("failed to check email existence")
		}
		if emailExists {
			return nil, appErrors.Conflict("email already exists")
		}
	}

	if user.Username != username {
		usernameExists, err := s.userRepo.UserExists(tx, model.GetUserParam{
			Username: username,
		})
		if err != nil {
			return nil, appErrors.InternalServer("failed to check username existence")
		}
		if usernameExists {
			return nil, appErrors.Conflict("username already exists")
		}
	}

	role, err := s.roleRepo.GetRoleByName(tx, roleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("role not found")
		}
		return nil, appErrors.InternalServer("failed to get role")
	}

	user.Email = email
	user.Username = username
	user.RoleID = role.RoleID
	user.Status = status

	if password != "" {
		passwordHash, err := s.bcrypt.GenerateFromPassword(password)
		if err != nil {
			return nil, appErrors.InternalServer("failed to hash password")
		}
		user.Password = passwordHash
	}

	err = s.userRepo.UpdateUser(tx, user)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update user")
	}

	after := newAuditUserSnapshot(user, role.RoleName)
	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionUpdate,
		Module:        model.AuditModuleUsers,
		TargetType:    "user",
		TargetID:      user.UserID.String(),
		TargetLabel:   user.Email,
		Detail:        fmt.Sprintf("Updated user %s", user.Email),
		PayloadBefore: before,
		PayloadAfter: map[string]any{
			"user":             after,
			"password_changed": password != "",
		},
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	updatedUser, err := s.userRepo.GetUserDetail(s.db, user.UserID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get user detail")
	}

	return &model.AdminUpdateUserResponse{
		User: *updatedUser,
	}, nil
}
