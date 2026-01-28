package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware creates a Gin middleware that records HTTP metrics
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		// Calculate request size
		requestSize := c.Request.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get response size
		responseSize := int64(c.Writer.Size())
		statusCode := c.Writer.Status()

		// Record metrics
		metrics.RecordHTTPRequest(method, path, statusCode, duration, requestSize, responseSize)
	}
}

// PageLoadMiddleware creates a middleware that tracks page load metrics
// This should be used on page-related endpoints
func PageLoadMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		pageID := c.Param("id")
		if pageID == "" {
			pageID = c.Param("page_id")
		}

		// Get user ID from context or header
		userID := ""
		if userIDVal, exists := c.Get("user_id"); exists {
			if userIDStr, ok := userIDVal.(string); ok {
				userID = userIDStr
			}
		}
		if userID == "" {
			userID = c.GetHeader("X-User-ID")
		}
		if userID == "" {
			userID = "anonymous"
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		success := c.Writer.Status() < 400

		// Record page load metrics if page ID is present
		if pageID != "" {
			metrics.RecordPageLoad(pageID, userID, duration, success)
		}
	}
}

