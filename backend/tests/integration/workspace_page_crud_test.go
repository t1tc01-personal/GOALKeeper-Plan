package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	apiRouter "goalkeeper-plan/cmd/api"
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newTestRouter now returns the actual engine created by your API package
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.GetConfig()
	log, err := logger.NewLogger(cfg.AppConfig)
	require.NoError(t, err)

	db, err := gorm.Open(postgres.Open(cfg.PostgresConfig.ConnectionString), &gorm.Config{})
	require.NoError(t, err)

	// Return the engine created by your router factory
	return apiRouter.NewRouter(db, *cfg, log)
}

func TestWorkspaceAndPageCRUD(t *testing.T) {
	if os.Getenv("NOTION_E2E_SKIP") == "1" {
		t.Skip("Skipping Notion workspace CRUD integration test")
	}

	// FIX 1: Capture the returned handler into a variable named 'router'
	// (or 'engine') so it's consistent throughout the test.
	router := newTestRouter(t)

	// 1) Create workspace
	createBody := map[string]any{
		"name": "Test Workspace",
	}
	bodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notion/workspaces", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req) // Now 'router' is defined

	require.Equal(t, http.StatusCreated, rec.Code, "expected workspace to be created")

	var created struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	require.True(t, created.Success, "response should indicate success")

	var workspace struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal(created.Data, &workspace))
	require.NotEmpty(t, workspace.ID, "workspace ID must be set")

	// 2) Create a page inside the workspace
	createPageBody := map[string]any{
		"workspaceId": workspace.ID,
		"title":       "First Page",
	}
	pageBytes, err := json.Marshal(createPageBody)
	require.NoError(t, err)

	pageReq := httptest.NewRequest(http.MethodPost, "/api/v1/notion/pages", bytes.NewReader(pageBytes))
	pageReq.Header.Set("Content-Type", "application/json")

	pageRec := httptest.NewRecorder()
	router.ServeHTTP(pageRec, pageReq) // Consistent naming

	require.Equal(t, http.StatusCreated, pageRec.Code, "expected page to be created")

	var pageCreated struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(pageRec.Body.Bytes(), &pageCreated))
	require.True(t, pageCreated.Success, "page response should indicate success")

	var page struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	require.NoError(t, json.Unmarshal(pageCreated.Data, &page))
	require.NotEmpty(t, page.ID, "page ID must be set")
	require.Equal(t, "First Page", page.Title)

	// 3) Fetch pages for workspace
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/notion/pages?workspaceId="+workspace.ID, nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq) // Used 'router' instead of undefined 'engine'

	require.Equal(t, http.StatusOK, listRec.Code, "expected page list to succeed")
}
