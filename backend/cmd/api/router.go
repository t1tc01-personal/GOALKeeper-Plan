package api

import (
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"
	user "goalkeeper-plan/internal/user/app"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewRouter creates a new Gin router with all routes configured
func NewRouter(db *gorm.DB, configs config.Configurations, logger logger.Logger) *gin.Engine {
	// Set Gin mode
	if !configs.AppConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(
		ginLoggerMiddleware(logger),
		gin.Recovery(),
		corsMiddleware(),
	)

	// Health check endpoints
	router.GET("/health", healthHandler(db, logger))
	router.GET("/liveness", livenessHandler(logger))
	router.GET("/ready/readiness", readinessHandler(db, logger))
	router.GET("/ready/liveliness", livenessHandler(logger))

	// API v1 routes
	apiV1 := router.Group("/api/v1")
	{
		// Register domain applications
		user.NewApplication(db, apiV1, configs, logger)
	}

	return router
}

// ginLoggerMiddleware creates a Gin middleware for logging
func ginLoggerMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("error", errorMessage),
		)
	}
}

// corsMiddleware creates a CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// healthHandler checks if the service is healthy (includes database check)
func healthHandler(db *gorm.DB, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("Database connection error", zap.Error(err))
			c.JSON(503, gin.H{
				"status": "unhealthy",
				"error":  "database connection error",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			logger.Error("Database ping failed", zap.Error(err))
			c.JSON(503, gin.H{
				"status": "unhealthy",
				"error":  "database ping failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "healthy",
		})
	}
}

// readinessHandler checks if the service is ready to serve traffic
func readinessHandler(db *gorm.DB, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("Database connection error", zap.Error(err))
			c.JSON(503, gin.H{
				"status": "not ready",
				"error":  "database connection error",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			logger.Error("Database ping failed", zap.Error(err))
			c.JSON(503, gin.H{
				"status": "not ready",
				"error":  "database ping failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ready",
		})
	}
}

// livenessHandler checks if the service is alive
func livenessHandler(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "alive",
		})
	}
}
