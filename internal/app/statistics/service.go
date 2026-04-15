package statistics

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
	repo  commandlog.StatisticsRepository
	cache cache.Cache
	ttl   time.Duration
}

func NewService(repo commandlog.StatisticsRepository, cache cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl}
}

func (s *Service) GetMostModifiedStatistics(ctx context.Context, filters commandlog.QueryFilters, limit int) ([]commandlog.StatisticsRow, error) {
	key := cacheKey(filters, limit)
	var out []commandlog.StatisticsRow
	if s.cache != nil && s.cache.Get(key, &out) {
		return out, nil
	}
	out, err := s.repo.GetMostModifiedStatistics(ctx, filters, limit)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Set(key, out, s.ttl)
	}
	return out, nil
}

func cacheKey(filters commandlog.QueryFilters, limit int) string {
	b, _ := json.Marshal(struct {
		F commandlog.QueryFilters
		L int
	}{filters, limit})
	sum := sha1.Sum(b)
	return fmt.Sprintf("statistics:most-modified:%x", sum)
}
