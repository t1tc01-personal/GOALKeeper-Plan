package cache

import "errors"

var (
	// ErrCacheKeyNotFound is returned when a cache key is not found
	ErrCacheKeyNotFound = errors.New("cache key not found")

	// ErrInvalidCacheData is returned when cached data is invalid
	ErrInvalidCacheData = errors.New("invalid cache data")
)
