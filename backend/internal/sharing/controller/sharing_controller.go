package controller

import (
	"fmt"
	"goalkeeper-plan/internal/response"
	"goalkeeper-plan/internal/sharing/model"
	"goalkeeper-plan/internal/sharing/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SharingController interface {
	GrantAccess(c *gin.Context)
	RevokeAccess(c *gin.Context)
	ListCollaborators(c *gin.Context)
}

type sharingController struct {
	sharingService service.SharingService
}

func NewSharingController(sharingService service.SharingService) SharingController {
	return &sharingController{
		sharingService: sharingService,
	}
}

// GrantAccessRequest represents the payload for granting access
type GrantAccessRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Role   string    `json:"role" binding:"required,oneof=viewer editor owner"`
}

// CollaboratorResponse represents a collaborator in the response
type CollaboratorResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PageID    uuid.UUID `json:"page_id"`
	Role      string    `json:"role"`
	CreatedAt string    `json:"created_at"`
}

// GrantAccess grants access to a page for a user
// POST /api/v1/notion/pages/:page_id/share
func (c *sharingController) GrantAccess(ctx *gin.Context) {
	pageIDStr := ctx.Param("page_id")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		response.SimpleErrorResponse(ctx, http.StatusBadRequest, "Invalid page ID")
		return
	}

	var req GrantAccessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.SimpleErrorResponse(ctx, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	role := model.SharePermissionRole(req.Role)

	err = c.sharingService.GrantAccess(ctx, pageID, req.UserID, role)
	if err != nil {
		response.SimpleErrorResponse(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to grant access: %v", err))
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Access granted successfully", map[string]interface{}{
		"page_id": pageID,
		"user_id": req.UserID,
		"role":    role,
	})
}

// RevokeAccess revokes access to a page for a user
// DELETE /api/v1/notion/pages/:page_id/share/:user_id
func (c *sharingController) RevokeAccess(ctx *gin.Context) {
	pageIDStr := ctx.Param("page_id")
	userIDStr := ctx.Param("user_id")

	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		response.SimpleErrorResponse(ctx, http.StatusBadRequest, "Invalid page ID")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.SimpleErrorResponse(ctx, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = c.sharingService.RevokeAccess(ctx, pageID, userID)
	if err != nil {
		response.SimpleErrorResponse(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to revoke access: %v", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ListCollaborators lists all collaborators for a page
// GET /api/v1/notion/pages/:page_id/collaborators
func (c *sharingController) ListCollaborators(ctx *gin.Context) {
	pageIDStr := ctx.Param("page_id")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		response.SimpleErrorResponse(ctx, http.StatusBadRequest, "Invalid page ID")
		return
	}

	collaborators, err := c.sharingService.ListCollaborators(ctx, pageID)
	if err != nil {
		response.SimpleErrorResponse(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to list collaborators: %v", err))
		return
	}

	// Convert to response format
	responses := make([]CollaboratorResponse, len(collaborators))
	for i, collab := range collaborators {
		responses[i] = CollaboratorResponse{
			ID:        collab.ID,
			UserID:    collab.UserID,
			PageID:    collab.PageID,
			Role:      string(collab.Role),
			CreatedAt: collab.CreatedAt.String(),
		}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Collaborators retrieved successfully", map[string]interface{}{
		"collaborators": responses,
		"count":         len(responses),
	})
}
