package api

import (
	"fmt"
	"strings"
	"time"

	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"
	auth "goalkeeper-plan/internal/auth/app"
	rbac "goalkeeper-plan/internal/rbac/app"
	user "goalkeeper-plan/internal/user/app"
	_ "goalkeeper-plan/docs" // swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
		corsMiddleware(configs),
	)

	// Health check endpoints
	router.GET("/health", healthHandler(db, logger))
	router.GET("/liveness", livenessHandler(logger))
	router.GET("/ready/readiness", readinessHandler(db, logger))
	router.GET("/ready/liveliness", livenessHandler(logger))

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	apiV1 := router.Group("/api/v1")
	{
		// Register domain applications
		auth.NewApplication(db, apiV1, configs, logger)
		user.NewApplication(db, apiV1, configs, logger)
		rbac.NewRBACApplication(db, apiV1, configs, logger)
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

// corsMiddleware creates a CORS middleware with configurable origins
func corsMiddleware(configs config.Configurations) gin.HandlerFunc {
	allowedOrigins := configs.AppConfig.CORS.AllowedOrigins
	if len(allowedOrigins) == 0 {
		// Default: allow all in development
		allowedOrigins = []string{"*"}
	}

	allowedMethods := configs.AppConfig.CORS.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}

	allowedHeaders := configs.AppConfig.CORS.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{
			"Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "accept", "origin",
			"Cache-Control", "X-Requested-With",
		}
	}

	exposedHeaders := configs.AppConfig.CORS.ExposedHeaders
	allowCredentials := configs.AppConfig.CORS.AllowCredentials
	maxAge := configs.AppConfig.CORS.MaxAge
	if maxAge == 0 {
		maxAge = 86400 // Default 24 hours
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		var allowedOrigin string

		if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
			allowed = true
			allowedOrigin = "*"
		} else {
			for _, ao := range allowedOrigins {
				if origin == ao {
					allowed = true
					allowedOrigin = ao
					break
				}
			}
		}

		if allowed {
			if allowedOrigin == "*" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		if allowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Set allowed methods
		methods := strings.Join(allowedMethods, ", ")
		c.Writer.Header().Set("Access-Control-Allow-Methods", methods)

		// Set allowed headers
		headers := strings.Join(allowedHeaders, ", ")
		c.Writer.Header().Set("Access-Control-Allow-Headers", headers)

		// Set exposed headers
		if len(exposedHeaders) > 0 {
			exposed := strings.Join(exposedHeaders, ", ")
			c.Writer.Header().Set("Access-Control-Expose-Headers", exposed)
		}

		// Set max age for preflight
		c.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", maxAge))

		// Handle preflight requests
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
