package indexes

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
	repo  commandlog.IndexRepository
	cache cache.Cache
	ttl   time.Duration
}

func NewService(repo commandlog.IndexRepository, cache cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl}
}

func (s *Service) GetTopFragmentedIndexes(ctx context.Context, filters commandlog.QueryFilters, limit int) ([]commandlog.IndexRow, error) {
	key := cacheKey(filters, limit)
	var out []commandlog.IndexRow
	if s.cache != nil && s.cache.Get(key, &out) {
		return out, nil
	}
	out, err := s.repo.GetTopFragmentedIndexes(ctx, filters, limit)
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
	return fmt.Sprintf("indexes:top-fragmented:%x", sum)
}
