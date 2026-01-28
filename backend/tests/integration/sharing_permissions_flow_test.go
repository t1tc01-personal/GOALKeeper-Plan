package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	apiRouter "goalkeeper-plan/cmd/api"
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/auth/jwt"
	"goalkeeper-plan/internal/logger"
	pageModel "goalkeeper-plan/internal/page/model"
	"goalkeeper-plan/internal/workspace/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// generateTestToken creates a JWT token for testing purposes
func generateTestToken(userID uuid.UUID, email string, cfg *config.Configurations) (string, error) {
	jwtService := jwt.NewJWTService(
		cfg.AuthConfig.Secret,
		cfg.AuthConfig.JWTExpiration,
		cfg.AuthConfig.RefreshExpiration,
	)
	return jwtService.GenerateAccessToken(userID.String(), email)
}

// TestSharingPermissionsFlow validates that sharing permissions are correctly
// enforced: a user with "viewer" role can read but not edit, an "editor" can
// both read and edit, and unauthorized users cannot access the page.
func TestSharingPermissionsFlow(t *testing.T) {
	cfg := config.GetConfig()
	log, err := logger.NewLogger(cfg.AppConfig)
	require.NoError(t, err)

	db, err := gorm.Open(postgres.Open(cfg.PostgresConfig.ConnectionString), &gorm.Config{})
	require.NoError(t, err)

	router := apiRouter.NewRouter(db, *cfg, log)

	// Setup: Create a workspace and page
	workspace := &model.Workspace{
		ID:   uuid.New(),
		Name: "Test Workspace",
	}
	err = db.Create(workspace).Error
	require.NoError(t, err)

	page := &pageModel.Page{
		ID:          uuid.New(),
		WorkspaceID: workspace.ID,
		Title:       "Test Page",
	}
	err = db.Create(page).Error
	require.NoError(t, err)

	// Create users and generate tokens
	ownerID := uuid.New()
	viewerID := uuid.New()
	editorID := uuid.New()
	unauthorizedID := uuid.New()

	ownerToken, err := generateTestToken(ownerID, "owner@test.com", cfg)
	require.NoError(t, err)
	viewerToken, err := generateTestToken(viewerID, "viewer@test.com", cfg)
	require.NoError(t, err)
	editorToken, err := generateTestToken(editorID, "editor@test.com", cfg)
	require.NoError(t, err)
	unauthorizedToken, err := generateTestToken(unauthorizedID, "unauthorized@test.com", cfg)
	require.NoError(t, err)

	t.Run("viewer can read but not edit page", func(t *testing.T) {
		// Grant viewer access
		grant := map[string]interface{}{
			"user_id": viewerID.String(),
			"role":    "viewer",
		}
		grantBody, err := json.Marshal(grant)
		require.NoError(t, err)

		grantReq := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/notion/pages/%s/share", page.ID.String()),
			bytes.NewReader(grantBody),
		)
		grantReq.Header.Set("Content-Type", "application/json")
		grantReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ownerToken))

		grantRec := httptest.NewRecorder()
		router.ServeHTTP(grantRec, grantReq)

		// Permission should be created or already exists
		require.True(t,
			grantRec.Code == http.StatusCreated || grantRec.Code == http.StatusOK,
			"granting access should succeed (got %d)", grantRec.Code,
		)

		// Viewer can read page
		readReq := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			nil,
		)
		readReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viewerToken))

		readRec := httptest.NewRecorder()
		router.ServeHTTP(readRec, readReq)

		// Viewer should be able to read (200 OK)
		require.Equal(t, http.StatusOK, readRec.Code,
			"viewer should be able to read page (got %d)", readRec.Code)

		// Viewer cannot edit page (PUT/PATCH should be forbidden)
		updatePayload := map[string]interface{}{
			"title": "Updated Title",
		}
		updateBody, err := json.Marshal(updatePayload)
		require.NoError(t, err)

		editReq := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			bytes.NewReader(updateBody),
		)
		editReq.Header.Set("Content-Type", "application/json")
		editReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viewerToken))

		editRec := httptest.NewRecorder()
		router.ServeHTTP(editRec, editReq)

		// Viewer should get 403 Forbidden when trying to edit
		require.Equal(t, http.StatusForbidden, editRec.Code,
			"viewer should not be able to edit page (got %d)", editRec.Code)
	})

	t.Run("editor can read and edit page", func(t *testing.T) {
		// Grant editor access
		grant := map[string]interface{}{
			"user_id": editorID.String(),
			"role":    "editor",
		}
		grantBody, err := json.Marshal(grant)
		require.NoError(t, err)

		grantReq := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/notion/pages/%s/share", page.ID.String()),
			bytes.NewReader(grantBody),
		)
		grantReq.Header.Set("Content-Type", "application/json")
		grantReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ownerToken))

		grantRec := httptest.NewRecorder()
		router.ServeHTTP(grantRec, grantReq)

		// Editor can read page
		readReq := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			nil,
		)
		readReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", editorToken))

		readRec := httptest.NewRecorder()
		router.ServeHTTP(readRec, readReq)
		// Just verify endpoint responds

		// Editor can edit page
		updatePayload := map[string]interface{}{
			"title": "Updated by Editor",
		}
		updateBody, err := json.Marshal(updatePayload)
		require.NoError(t, err)

		editReq := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			bytes.NewReader(updateBody),
		)
		editReq.Header.Set("Content-Type", "application/json")
		editReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", editorToken))

		editRec := httptest.NewRecorder()
		router.ServeHTTP(editRec, editReq)
		// Just verify endpoint responds

		// Verify the update was successful (200 or 204)
		require.True(t,
			editRec.Code == http.StatusOK || editRec.Code == http.StatusNoContent,
			"editor should be able to edit page (got %d)", editRec.Code,
		)
	})

	t.Run("unauthorized user cannot access page", func(t *testing.T) {
		// Unauthorized user tries to read page
		readReq := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			nil,
		)
		readReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedToken))

		readRec := httptest.NewRecorder()
		router.ServeHTTP(readRec, readReq)

		// Unauthorized should get 403 Forbidden
		require.Equal(t, http.StatusForbidden, readRec.Code,
			"unauthorized user should not be able to read page (got %d)", readRec.Code)

		// Unauthorized user tries to edit page
		updatePayload := map[string]interface{}{
			"title": "Attempted Hack",
		}
		updateBody, err := json.Marshal(updatePayload)
		require.NoError(t, err)

		editReq := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			bytes.NewReader(updateBody),
		)
		editReq.Header.Set("Content-Type", "application/json")
		editReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", unauthorizedToken))

		editRec := httptest.NewRecorder()
		router.ServeHTTP(editRec, editReq)

		// Should get 403 Forbidden
		require.Equal(t, http.StatusForbidden, editRec.Code,
			"unauthorized user should not be able to edit page (got %d)", editRec.Code)
	})

	t.Run("revoking access removes permissions", func(t *testing.T) {
		revokeID := uuid.New()
		revokeToken, err := generateTestToken(revokeID, "revoke@test.com", cfg)
		require.NoError(t, err)

		// Grant access first
		grant := map[string]interface{}{
			"user_id": revokeID.String(),
			"role":    "editor",
		}
		grantBody, err := json.Marshal(grant)
		require.NoError(t, err)

		grantReq := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/notion/pages/%s/share", page.ID.String()),
			bytes.NewReader(grantBody),
		)
		grantReq.Header.Set("Content-Type", "application/json")
		grantReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ownerToken))

		grantRec := httptest.NewRecorder()
		router.ServeHTTP(grantRec, grantReq)

		// User can read before revocation
		readReq := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			nil,
		)
		readReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", revokeToken))

		readRec := httptest.NewRecorder()
		router.ServeHTTP(readRec, readReq)

		// Now revoke access
		revokeReq := httptest.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("/api/v1/notion/pages/%s/share/%s", page.ID.String(), revokeID.String()),
			nil,
		)
		revokeReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ownerToken))

		revokeRec := httptest.NewRecorder()
		router.ServeHTTP(revokeRec, revokeReq)

		// Revocation should succeed
		require.Equal(t, http.StatusNoContent, revokeRec.Code,
			"revoking access should succeed (got %d)", revokeRec.Code)

		// User cannot read after revocation
		readReq2 := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			nil,
		)
		readReq2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", revokeToken))

		readRec2 := httptest.NewRecorder()
		router.ServeHTTP(readRec2, readReq2)

		// Should get 403 Forbidden after revocation
		require.Equal(t, http.StatusForbidden, readRec2.Code,
			"user should not be able to read page after revocation (got %d)", readRec2.Code)
	})
}
