package meta

import (
	"context"
	"time"

	"maintenance-dashboard/internal/domain/commandlog"
	"maintenance-dashboard/internal/infra/cache"
)

type Service struct {
	repo  commandlog.MetaRepository
	cache cache.Cache
	ttl   time.Duration
}

func NewService(repo commandlog.MetaRepository, cache cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl}
}

func (s *Service) GetFilterOptions(ctx context.Context) (commandlog.FilterOptions, error) {
	const key = "meta:filters"
	var out commandlog.FilterOptions
	if s.cache != nil && s.cache.Get(key, &out) {
		return out, nil
	}
	out, err := s.repo.GetFilterOptions(ctx)
	if err != nil {
		return commandlog.FilterOptions{}, err
	}
	if s.cache != nil {
		s.cache.Set(key, out, s.ttl)
	}
	return out, nil
}
