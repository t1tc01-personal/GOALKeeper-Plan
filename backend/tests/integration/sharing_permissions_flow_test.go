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
	"goalkeeper-plan/internal/logger"
	pageModel "goalkeeper-plan/internal/page/model"
	"goalkeeper-plan/internal/workspace/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

	// Create users
	ownerID := uuid.New()
	viewerID := uuid.New()
	editorID := uuid.New()
	unauthorizedID := uuid.New()

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
		grantReq.Header.Set("X-User-ID", ownerID.String()) // Owner grants access

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
		readReq.Header.Set("X-User-ID", viewerID.String())

		readRec := httptest.NewRecorder()
		router.ServeHTTP(readReq, readRec)

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
		editReq.Header.Set("X-User-ID", viewerID.String())

		editRec := httptest.NewRecorder()
		router.ServeHTTP(editReq, editRec)

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
		grantReq.Header.Set("X-User-ID", ownerID.String())

		grantRec := httptest.NewRecorder()
		router.ServeHTTP(grantRec, grantReq)

		// Editor can read page
		readReq := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			nil,
		)
		readReq.Header.Set("X-User-ID", editorID.String())

		readRec := httptest.NewRecorder()
		router.ServeHTTP(readReq, readReq)
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
		editReq.Header.Set("X-User-ID", editorID.String())

		editRec := httptest.NewRecorder()
		router.ServeHTTP(editReq, editReq)
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
		readReq.Header.Set("X-User-ID", unauthorizedID.String())

		readRec := httptest.NewRecorder()
		router.ServeHTTP(readReq, readReq)

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
		editReq.Header.Set("X-User-ID", unauthorizedID.String())

		editRec := httptest.NewRecorder()
		router.ServeHTTP(editReq, editReq)

		// Should get 403 Forbidden
		require.Equal(t, http.StatusForbidden, editRec.Code,
			"unauthorized user should not be able to edit page (got %d)", editRec.Code)
	})

	t.Run("revoking access removes permissions", func(t *testing.T) {
		revokeID := uuid.New()

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
		grantReq.Header.Set("X-User-ID", ownerID.String())

		grantRec := httptest.NewRecorder()
		router.ServeHTTP(grantReq, grantReq)

		// User can read before revocation
		readReq := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			nil,
		)
		readReq.Header.Set("X-User-ID", revokeID.String())

		readRec := httptest.NewRecorder()
		router.ServeHTTP(readReq, readReq)

		// Now revoke access
		revokeReq := httptest.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("/api/v1/notion/pages/%s/share/%s", page.ID.String(), revokeID.String()),
			nil,
		)
		revokeReq.Header.Set("X-User-ID", ownerID.String())

		revokeRec := httptest.NewRecorder()
		router.ServeHTTP(revokeReq, revokeReq)

		// Revocation should succeed
		require.Equal(t, http.StatusNoContent, revokeRec.Code,
			"revoking access should succeed (got %d)", revokeRec.Code)

		// User cannot read after revocation
		readReq2 := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/notion/pages/%s", page.ID.String()),
			nil,
		)
		readReq2.Header.Set("X-User-ID", revokeID.String())

		readRec2 := httptest.NewRecorder()
		router.ServeHTTP(readReq2, readReq2)

		// Should get 403 Forbidden after revocation
		require.Equal(t, http.StatusForbidden, readRec2.Code,
			"user should not be able to read page after revocation (got %d)", readRec2.Code)
	})
}
