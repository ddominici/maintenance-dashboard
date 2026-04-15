package errors

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"time"

	"maintenance-dashboard/internal/domain/commandlog"
	"maintenance-dashboard/internal/infra/cache"
)

type Service struct {
	repo  commandlog.MaintenanceErrorsRepository
	cache cache.Cache
	ttl   time.Duration
}

func NewService(repo commandlog.MaintenanceErrorsRepository, cache cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl}
}

func (s *Service) GetMaintenanceErrorsReport(ctx context.Context, filters commandlog.QueryFilters, granularity string, limit int) (commandlog.MaintenanceErrorsReport, error) {
	key := cacheKey(filters, granularity, limit)
	var out commandlog.MaintenanceErrorsReport
	if s.cache != nil && s.cache.Get(key, &out) {
		return out, nil
	}
	out, err := s.repo.GetMaintenanceErrorsReport(ctx, filters, granularity, limit)
	if err != nil {
		return commandlog.MaintenanceErrorsReport{}, err
	}
	if s.cache != nil {
		s.cache.Set(key, out, s.ttl)
	}
	return out, nil
}

func cacheKey(filters commandlog.QueryFilters, granularity string, limit int) string {
	b, _ := json.Marshal(struct {
		F commandlog.QueryFilters
		G string
		L int
	}{filters, granularity, limit})
	sum := sha1.Sum(b)
	return fmt.Sprintf("errors:report:%x", sum)
}
