package service

import (
	"errors"
	"math"
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	constants "github.com/azmiagr/unesco-hackathon/pkg/constant"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
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
}

type UserService struct {
	db       *gorm.DB
	userRepo repository.IUserRepository
	roleRepo repository.IRoleRepository
}

func NewUserService(userRepo repository.IUserRepository, roleRepo repository.IRoleRepository) IUserService {
	return &UserService{
		db:       mariadb.Connection,
		userRepo: userRepo,
		roleRepo: roleRepo,
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
		role, err := s.roleRepo.GetRole(tx, user.RoleID)
		if err != nil {
			return nil, appErrors.InternalServer("failed to get current role")
		}
		finalRoleName = role.RoleName
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

	err = s.userRepo.HardDeleteUser(tx, user.UserID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete user")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &model.AdminDeleteUserResponse{
		UserID: user.UserID,
	}, nil

}
