package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"
	redisClient "goalkeeper-plan/internal/redis"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	gpb "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewServerCmd is a function to create new server command
func NewServerCmd(configs config.Configurations, logger logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "run API server",
		Long:  "run GOALKeeper-Plan API server",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			defer func() {
				err := recover()
				if err != nil {
					logger.Error("panic recovered", zap.Any("error", err))
				}
			}()

			defer logger.Flush(5)

			// Initialize database connection
			gormDB, err := gorm.Open(gpb.Open(configs.PostgresConfig.ConnectionString), &gorm.Config{
				PrepareStmt: false,
			})

			if err != nil {
				logger.Error("Getting error create gorm db", zap.Error(err))
				os.Exit(1)
			}

			logger.Info("Connected to PostgreSQL database successfully")

			// Initialize Redis connection
			_, err = redisClient.GetRedisClient(configs.RedisConfig, logger)
			if err != nil {
				logger.Error("Getting error initialize Redis client", zap.Error(err))
				os.Exit(1)
			}
			defer redisClient.CloseRedis()

			// Initialize router
			router := NewRouter(gormDB, configs, logger)

			port := configs.AppConfig.Port
			if port == "" {
				port = "8000"
			}

			server := &http.Server{
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 30 * time.Second,
				Addr:         ":" + port,
				Handler:      router,
			}

			// Graceful shutdown
			idleConnectionsClosed := make(chan struct{})

			go func() {
				c := make(chan os.Signal, 1)
				signal.Notify(c, os.Interrupt)
				signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

				<-c

				ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
				defer cancel()
				logger.Info("Stopping server...")

				// Close Redis connection
				if err := redisClient.CloseRedis(); err != nil {
					logger.Error("Failed to close Redis connection", zap.Error(err))
				}

				if err := server.Shutdown(ctx); err != nil {
					logger.Error("Graceful shutdown has failed with error", zap.Error(err))
				}

				close(idleConnectionsClosed)
			}()

			go func() {
				logger.Info("Starting server", zap.String("port", port))
				if err := server.ListenAndServe(); err != http.ErrServerClosed {
					logger.Error("Run server has error", zap.Error(err))
					os.Exit(1)
				} else {
					logger.Info("Server was closed by shutdown gracefully")
				}
			}()

			<-idleConnectionsClosed
		},
	}
}
