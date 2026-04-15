package maintenance

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
	repo  commandlog.MaintenanceRepository
	cache cache.Cache
	ttl   time.Duration
}

func NewService(repo commandlog.MaintenanceRepository, cache cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl}
}

func (s *Service) GetMaintenanceReport(ctx context.Context, filters commandlog.QueryFilters, granularity string) (commandlog.MaintenanceReport, error) {
	key := cacheKey(filters, granularity)
	var out commandlog.MaintenanceReport
	if s.cache != nil && s.cache.Get(key, &out) {
		return out, nil
	}
	out, err := s.repo.GetMaintenanceReport(ctx, filters, granularity)
	if err != nil {
		return commandlog.MaintenanceReport{}, err
	}
	if s.cache != nil {
		s.cache.Set(key, out, s.ttl)
	}
	return out, nil
}

func cacheKey(filters commandlog.QueryFilters, granularity string) string {
	b, _ := json.Marshal(struct {
		F commandlog.QueryFilters
		G string
	}{filters, granularity})
	sum := sha1.Sum(b)
	return fmt.Sprintf("maintenance:report:%x", sum)
}
