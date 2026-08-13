package service

import (
	"errors"
	"math"

	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultLeaderboardLimit = 10
	maxLeaderboardLimit     = 100
)

type ILeaderboardService interface {
	ListLeaderboard(userID uuid.UUID, req model.ListLeaderboardRequest) (*model.ListLeaderboardResponse, error)
}

type LeaderboardService struct {
	db              *gorm.DB
	userProfileRepo repository.IUserProfileRepository
}

func NewLeaderboardService(userProfileRepo repository.IUserProfileRepository) ILeaderboardService {
	return &LeaderboardService{
		db:              mariadb.Connection,
		userProfileRepo: userProfileRepo,
	}
}

func (s *LeaderboardService) ListLeaderboard(userID uuid.UUID, req model.ListLeaderboardRequest) (*model.ListLeaderboardResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultLeaderboardLimit
	}
	if limit > maxLeaderboardLimit {
		limit = maxLeaderboardLimit
	}

	entries, total, err := s.userProfileRepo.ListLeaderboard(s.db, model.ListLeaderboardParam{
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list leaderboard")
	}

	var me *model.LeaderboardEntryRow
	for i := range entries {
		if entries[i].UserID == userID {
			entries[i].IsCurrentUser = true
			me = &entries[i]
			break
		}
	}

	if me == nil {
		currentUserEntry, err := s.userProfileRepo.GetUserLeaderboardRank(s.db, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, appErrors.NotFound("leaderboard profile not found")
			}
			return nil, appErrors.InternalServer("failed to get current user leaderboard rank")
		}
		currentUserEntry.IsCurrentUser = true
		me = currentUserEntry
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.ListLeaderboardResponse{
		Entries: entries,
		Me:      me,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
