package middleware

import (
	"goalkeeper-plan/internal/response"
	"goalkeeper-plan/internal/workspace/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthorizePageEdit ensures user has edit permission for the page
func AuthorizePageEdit(sharingService service.SharingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageIDStr := c.Param("id")
		if pageIDStr == "" {
			pageIDStr = c.Param("page_id")
		}

		if pageIDStr == "" {
			response.SimpleErrorResponse(c, http.StatusBadRequest, "Page ID is required")
			c.Abort()
			return
		}

		pageID, err := uuid.Parse(pageIDStr)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusBadRequest, "Invalid page ID")
			c.Abort()
			return
		}

		// Get user ID from header (would come from auth middleware in real app)
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			response.SimpleErrorResponse(c, http.StatusUnauthorized, "User ID is required")
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
			c.Abort()
			return
		}

		// Check if user has edit permission
		canEdit, err := sharingService.CanEdit(c, pageID, userID)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusInternalServerError, "Failed to check permissions")
			c.Abort()
			return
		}

		if !canEdit {
			response.SimpleErrorResponse(c, http.StatusForbidden, "You don't have permission to edit this page")
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthorizePageRead ensures user has at least read permission for the page
func AuthorizePageRead(sharingService service.SharingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageIDStr := c.Param("id")
		if pageIDStr == "" {
			pageIDStr = c.Param("page_id")
		}

		if pageIDStr == "" {
			response.SimpleErrorResponse(c, http.StatusBadRequest, "Page ID is required")
			c.Abort()
			return
		}

		pageID, err := uuid.Parse(pageIDStr)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusBadRequest, "Invalid page ID")
			c.Abort()
			return
		}

		// Get user ID from header (would come from auth middleware in real app)
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			response.SimpleErrorResponse(c, http.StatusUnauthorized, "User ID is required")
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
			c.Abort()
			return
		}

		// Check if user has read permission
		hasAccess, err := sharingService.HasAccess(c, pageID, userID)
		if err != nil {
			response.SimpleErrorResponse(c, http.StatusInternalServerError, "Failed to check permissions")
			c.Abort()
			return
		}

		if !hasAccess {
			response.SimpleErrorResponse(c, http.StatusForbidden, "You don't have permission to access this page")
			c.Abort()
			return
		}

		c.Next()
	}
}
