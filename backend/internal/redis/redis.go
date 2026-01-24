package redis

import (
	"context"
	"fmt"
	"sync"
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// GetRedisClient returns a singleton Redis client instance
func GetRedisClient(cfg config.RedisConfig, logger logger.Logger) (*redis.Client, error) {
	var err error
	redisOnce.Do(func() {
		addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
		
		opts := &redis.Options{
			Addr:     addr,
			Password: cfg.RedisPassword,
			DB:       0,
		}

		redisClient = redis.NewClient(opts)

		// Test connection
		ctx := context.Background()
		if err = redisClient.Ping(ctx).Err(); err != nil {
			logger.Error("Failed to connect to Redis", zap.Error(err))
			return
		}

		logger.Info("Connected to Redis successfully", 
			zap.String("addr", addr))
	})

	return redisClient, err
}

// CloseRedis closes the Redis client connection
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

