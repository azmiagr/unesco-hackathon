package service

import (
	"math"
	"strings"
	"time"

	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultAdminAuditLogLimit = 10
	maxAdminAuditLogLimit     = 100
)

var allowedAuditActions = map[string]bool{
	model.AuditActionLogin:        true,
	model.AuditActionCreate:       true,
	model.AuditActionUpdate:       true,
	model.AuditActionDelete:       true,
	model.AuditActionConfigChange: true,
}

var allowedAuditModules = map[string]bool{
	model.AuditModuleAuth:   true,
	model.AuditModuleCMS:    true,
	model.AuditModuleUsers:  true,
	model.AuditModuleShop:   true,
	model.AuditModuleConfig: true,
	model.AuditModuleSystem: true,
}

type IAuditLogService interface {
	ListAuditLogsByAdmin(req model.AdminListAuditLogsRequest) (*model.AdminListAuditLogsResponse, error)
}

type AuditLogService struct {
	db           *gorm.DB
	auditLogRepo repository.IAuditLogRepository
}

func NewAuditLogService(auditLogRepo repository.IAuditLogRepository) IAuditLogService {
	return &AuditLogService{
		db:           mariadb.Connection,
		auditLogRepo: auditLogRepo,
	}
}

func (s *AuditLogService) ListAuditLogsByAdmin(req model.AdminListAuditLogsRequest) (*model.AdminListAuditLogsResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultAdminAuditLogLimit
	}
	if limit > maxAdminAuditLogLimit {
		limit = maxAdminAuditLogLimit
	}

	actorAdminID := uuid.Nil
	actorAdminIDRaw := strings.TrimSpace(req.ActorAdminID)
	if actorAdminIDRaw != "" {
		parsedActorAdminID, err := uuid.Parse(actorAdminIDRaw)
		if err != nil {
			return nil, appErrors.BadRequest("invalid actor admin id")
		}
		actorAdminID = parsedActorAdminID
	}

	actionType := strings.ToLower(strings.TrimSpace(req.ActionType))
	if actionType == "" {
		actionType = strings.ToLower(strings.TrimSpace(req.Action))
	}
	if actionType != "" && !allowedAuditActions[actionType] {
		return nil, appErrors.BadRequest("invalid audit action")
	}

	module := strings.ToLower(strings.TrimSpace(req.Module))
	if module != "" && !allowedAuditModules[module] {
		return nil, appErrors.BadRequest("invalid audit module")
	}

	from, to, err := parseAuditLogRange(req)
	if err != nil {
		return nil, err
	}
	if from != nil && to != nil && from.After(*to) {
		return nil, appErrors.BadRequest("from must be before to")
	}

	auditLogs, total, err := s.auditLogRepo.ListAdminAuditLogs(s.db, model.ListAdminAuditLogsParam{
		ActorAdminID: actorAdminID,
		ActionType:   actionType,
		Module:       module,
		TargetType:   strings.TrimSpace(req.TargetType),
		From:         from,
		To:           to,
		Limit:        limit,
		Offset:       (page - 1) * limit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list audit logs")
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.AdminListAuditLogsResponse{
		AuditLogs: auditLogs,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func parseAuditLogRange(req model.AdminListAuditLogsRequest) (*time.Time, *time.Time, error) {
	fromRaw := strings.TrimSpace(req.From)
	toRaw := strings.TrimSpace(req.To)

	if fromRaw != "" || toRaw != "" {
		from, err := parseOptionalAuditLogTime(fromRaw, false)
		if err != nil {
			return nil, nil, appErrors.BadRequest("invalid from date")
		}
		to, err := parseOptionalAuditLogTime(toRaw, true)
		if err != nil {
			return nil, nil, appErrors.BadRequest("invalid to date")
		}
		return from, to, nil
	}

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeValue := strings.ToLower(strings.TrimSpace(req.Range))

	switch rangeValue {
	case "", "all":
		return nil, nil, nil
	case "today":
		end := startOfToday.AddDate(0, 0, 1).Add(-time.Nanosecond)
		return &startOfToday, &end, nil
	case "yesterday":
		start := startOfToday.AddDate(0, 0, -1)
		end := startOfToday.Add(-time.Nanosecond)
		return &start, &end, nil
	case "7d", "last_7_days":
		start := startOfToday.AddDate(0, 0, -6)
		end := startOfToday.AddDate(0, 0, 1).Add(-time.Nanosecond)
		return &start, &end, nil
	case "30d", "last_30_days":
		start := startOfToday.AddDate(0, 0, -29)
		end := startOfToday.AddDate(0, 0, 1).Add(-time.Nanosecond)
		return &start, &end, nil
	default:
		return nil, nil, appErrors.BadRequest("invalid audit log range")
	}
}

func parseOptionalAuditLogTime(raw string, endOfDay bool) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	if parsedTime, err := time.Parse(time.RFC3339, raw); err == nil {
		return &parsedTime, nil
	}

	parsedDate, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		parsedDate = parsedDate.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}

	return &parsedDate, nil
}
