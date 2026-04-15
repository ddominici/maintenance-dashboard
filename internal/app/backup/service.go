package backup

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
	repo  commandlog.BackupRepository
	cache cache.Cache
	ttl   time.Duration
}

func NewService(repo commandlog.BackupRepository, cache cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl}
}

func (s *Service) GetBackupReport(ctx context.Context, filters commandlog.QueryFilters, granularity string) (commandlog.BackupReport, error) {
	key := cacheKey(filters, granularity)
	var out commandlog.BackupReport
	if s.cache != nil && s.cache.Get(key, &out) {
		return out, nil
	}
	out, err := s.repo.GetBackupReport(ctx, filters, granularity)
	if err != nil {
		return commandlog.BackupReport{}, err
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
	return fmt.Sprintf("backup:report:%x", sum)
}
