package dashboard

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
	repo  commandlog.DashboardRepository
	cache cache.Cache
	ttl   time.Duration
}

func NewService(repo commandlog.DashboardRepository, cache cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl}
}

func (s *Service) GetSummary(ctx context.Context, filters commandlog.QueryFilters) (commandlog.DashboardSummary, error) {
	key := cacheKey(filters)
	var out commandlog.DashboardSummary
	if s.cache != nil && s.cache.Get(key, &out) {
		return out, nil
	}
	out, err := s.repo.GetSummary(ctx, filters)
	if err != nil {
		return commandlog.DashboardSummary{}, err
	}
	if s.cache != nil {
		s.cache.Set(key, out, s.ttl)
	}
	return out, nil
}

func cacheKey(filters commandlog.QueryFilters) string {
	b, _ := json.Marshal(filters)
	sum := sha1.Sum(b)
	return fmt.Sprintf("dashboard:summary:%x", sum)
}
