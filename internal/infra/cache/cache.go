package cache

import "time"

type Cache interface {
	Get(key string, dest any) bool
	Set(key string, value any, ttl time.Duration)
	Delete(key string)
	Clear()
}
