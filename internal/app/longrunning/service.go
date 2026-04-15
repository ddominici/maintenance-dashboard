package longrunning

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
	repo  commandlog.LongRunningRepository
	cache cache.Cache
	ttl   time.Duration
}

func NewService(repo commandlog.LongRunningRepository, cache cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl}
}

func (s *Service) GetLongRunningReport(ctx context.Context, filters commandlog.QueryFilters, granularity string, minDurationSeconds int, limit int) (commandlog.LongRunningReport, error) {
	key := cacheKey(filters, granularity, minDurationSeconds, limit)
	var out commandlog.LongRunningReport
	if s.cache != nil && s.cache.Get(key, &out) {
		return out, nil
	}
	out, err := s.repo.GetLongRunningReport(ctx, filters, granularity, minDurationSeconds, limit)
	if err != nil {
		return commandlog.LongRunningReport{}, err
	}
	if s.cache != nil {
		s.cache.Set(key, out, s.ttl)
	}
	return out, nil
}

func cacheKey(filters commandlog.QueryFilters, granularity string, minDuration int, limit int) string {
	b, _ := json.Marshal(struct {
		F commandlog.QueryFilters
		G string
		M int
		L int
	}{filters, granularity, minDuration, limit})
	sum := sha1.Sum(b)
	return fmt.Sprintf("longrunning:report:%x", sum)
}
