package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService provides caching functionality for frequently accessed data
type CacheService interface {
	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get retrieves a value from cache
	Get(ctx context.Context, key string, dest interface{}) error

	// Delete removes a key from cache
	Delete(ctx context.Context, keys ...string) error

	// InvalidatePattern deletes all keys matching a pattern
	InvalidatePattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// GetOrSet gets a value from cache, or computes and sets it if missing
	GetOrSet(ctx context.Context, key string, ttl time.Duration, compute func() (interface{}, error)) (interface{}, error)
}

type cacheService struct {
	client *redis.Client
}

// NewCacheService creates a new cache service
func NewCacheService(client *redis.Client) CacheService {
	return &cacheService{client: client}
}

// Set stores a value in cache with TTL
func (c *cacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a value from cache
func (c *cacheService) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheKeyNotFound
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	return json.Unmarshal(data, dest)
}

// Delete removes keys from cache
func (c *cacheService) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// InvalidatePattern deletes all keys matching a pattern
func (c *cacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan pattern: %w", err)
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists
func (c *cacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetOrSet gets a value from cache, or computes and sets it if missing
func (c *cacheService) GetOrSet(ctx context.Context, key string, ttl time.Duration, compute func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	data, err := c.client.Get(ctx, key).Bytes()
	if err == nil {
		return data, nil
	}

	// If not found, compute the value
	value, err := compute()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail if caching fails
		fmt.Printf("warning: failed to cache value for key %s: %v\n", key, err)
	}

	return value, nil
}

// Cache key patterns for different entities
const (
	// Block cache patterns
	BlockByIDPattern      = "block:id:%s"
	BlocksByPagePattern   = "blocks:page:%s"
	BlocksByParentPattern = "blocks:parent:%s"

	// Page cache patterns
	PageByIDPattern         = "page:id:%s"
	PagesByWorkspacePattern = "pages:workspace:%s"
	PageHierarchyPattern    = "pages:hierarchy:%s"

	// Workspace cache patterns
	WorkspaceByIDPattern = "workspace:id:%s"
	WorkspacesPattern    = "workspaces:list"

	// Default TTL for different cache types
	DefaultBlockTTL     = 10 * time.Minute
	DefaultPageTTL      = 10 * time.Minute
	DefaultWorkspaceTTL = 15 * time.Minute
	DefaultShortTTL     = 5 * time.Minute
	DefaultLongTTL      = 30 * time.Minute
)
