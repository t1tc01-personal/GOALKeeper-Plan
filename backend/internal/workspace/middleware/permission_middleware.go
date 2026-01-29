package middleware

import (
	"time"

	"goalkeeper-plan/internal/metrics"
	"goalkeeper-plan/internal/response"
	sharingService "goalkeeper-plan/internal/sharing/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthorizePageEdit ensures user has edit permission for the page
func AuthorizePageEdit(sharingService sharingService.SharingService, metricsInstance *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		pageIDStr := c.Param("id")
		if pageIDStr == "" {
			pageIDStr = c.Param("page_id")
		}

		if pageIDStr == "" {
			response.SimpleErrorResponse(c, http.StatusBadRequest, "Page ID is required")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "edit", false, time.Since(start), "missing_page_id")
			}
			return
		}

		pageID, err := uuid.Parse(pageIDStr)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusBadRequest, "Invalid page ID")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "edit", false, time.Since(start), "invalid_page_id")
			}
			return
		}

		// Get user ID from header (would come from auth middleware in real app)
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			response.SimpleErrorResponse(c, http.StatusUnauthorized, "User ID is required")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "edit", false, time.Since(start), "missing_user_id")
			}
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "edit", false, time.Since(start), "invalid_user_id")
			}
			return
		}

		// Check if user has edit permission
		canEdit, err := sharingService.CanEdit(c, pageID, userID)
		duration := time.Since(start)
		
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusInternalServerError, "Failed to check permissions")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "edit", false, duration, "service_error")
			}
			return
		}

		if !canEdit {
			response.SimpleErrorResponse(c, http.StatusForbidden, "You don't have permission to edit this page")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "edit", false, duration, "insufficient_permissions")
			}
			return
		}

		if metricsInstance != nil {
			metricsInstance.RecordPermissionCheck("page", "edit", true, duration, "")
		}
		c.Next()
	}
}

// AuthorizePageRead ensures user has at least read permission for the page
func AuthorizePageRead(sharingService sharingService.SharingService, metricsInstance *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		pageIDStr := c.Param("id")
		if pageIDStr == "" {
			pageIDStr = c.Param("page_id")
		}

		if pageIDStr == "" {
			response.SimpleErrorResponse(c, http.StatusBadRequest, "Page ID is required")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "read", false, time.Since(start), "missing_page_id")
			}
			return
		}

		pageID, err := uuid.Parse(pageIDStr)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusBadRequest, "Invalid page ID")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "read", false, time.Since(start), "invalid_page_id")
			}
			return
		}

		// Get user ID from header (would come from auth middleware in real app)
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			response.SimpleErrorResponse(c, http.StatusUnauthorized, "User ID is required")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "read", false, time.Since(start), "missing_user_id")
			}
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "read", false, time.Since(start), "invalid_user_id")
			}
			return
		}

		// Check if user has read permission
		hasAccess, err := sharingService.HasAccess(c, pageID, userID)
		duration := time.Since(start)
		
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusInternalServerError, "Failed to check permissions")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "read", false, duration, "service_error")
			}
			return
		}

		if !hasAccess {
			response.SimpleErrorResponse(c, http.StatusForbidden, "You don't have permission to access this page")
			c.Abort()
			if metricsInstance != nil {
				metricsInstance.RecordPermissionCheck("page", "read", false, duration, "insufficient_permissions")
			}
			return
		}

		if metricsInstance != nil {
			metricsInstance.RecordPermissionCheck("page", "read", true, duration, "")
		}
		c.Next()
	}
}
