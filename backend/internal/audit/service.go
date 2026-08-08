package audit

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const maxAuditRangeDays = 90

type Repository interface {
	List(ctx context.Context, filter listQuery) ([]Log, error)
	Count(ctx context.Context, filter listQuery) (int, error)
}

type Service struct {
	repo     Repository
	location *time.Location
}

func NewService(repo Repository, location *time.Location) Service {
	return Service{repo: repo, location: location}
}

func (s Service) List(ctx context.Context, filter ListFilter) (List, error) {
	query, err := s.normalize(filter)
	if err != nil {
		return List{}, err
	}

	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return List{}, err
	}
	items, err := s.repo.List(ctx, query)
	if err != nil {
		return List{}, err
	}
	pages := 0
	if total > 0 {
		pages = (total + query.PageSize - 1) / query.PageSize
	}
	return List{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: pages,
	}, nil
}

func (s Service) normalize(filter ListFilter) (listQuery, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	action := Action(strings.TrimSpace(string(filter.Action)))
	entityType := EntityType(strings.TrimSpace(string(filter.EntityType)))
	entityID := strings.TrimSpace(filter.EntityID)
	if action != "" && action != ActionAttendanceCorrected && action != ActionFaceEnrollmentReset {
		return listQuery{}, ErrInvalidFilter
	}
	if entityType != "" && entityType != EntityAttendanceRecord && entityType != EntityFaceProfile {
		return listQuery{}, ErrInvalidFilter
	}
	if entityID != "" {
		var parsed pgtype.UUID
		if err := parsed.Scan(entityID); err != nil || !parsed.Valid {
			return listQuery{}, ErrInvalidFilter
		}
	}

	from, to, err := s.parseRange(filter.DateFrom, filter.DateTo)
	if err != nil {
		return listQuery{}, err
	}
	return listQuery{
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		DateFrom:   from,
		DateTo:     to,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s Service) parseRange(fromValue, toValue string) (*time.Time, *time.Time, error) {
	fromValue = strings.TrimSpace(fromValue)
	toValue = strings.TrimSpace(toValue)
	if fromValue == "" && toValue == "" {
		return nil, nil, nil
	}
	if fromValue == "" || toValue == "" {
		return nil, nil, ErrInvalidFilter
	}

	from, err := time.ParseInLocation("2006-01-02", fromValue, s.location)
	if err != nil {
		return nil, nil, ErrInvalidFilter
	}
	to, err := time.ParseInLocation("2006-01-02", toValue, s.location)
	if err != nil || to.Before(from) {
		return nil, nil, ErrInvalidFilter
	}
	if int(to.Sub(from).Hours()/24)+1 > maxAuditRangeDays {
		return nil, nil, ErrInvalidFilter
	}
	toExclusive := to.AddDate(0, 0, 1)
	return &from, &toExclusive, nil
}
