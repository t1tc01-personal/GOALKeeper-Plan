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

func TestBlockFlow(t *testing.T) {
	if os.Getenv("NOTION_E2E_SKIP") == "1" {
		t.Skip("Skipping block flow integration test")
	}

	cfg := config.GetConfig()
	log, err := logger.NewLogger(cfg.AppConfig)
	require.NoError(t, err)

	db, err := gorm.Open(postgres.Open(cfg.PostgresConfig.ConnectionString), &gorm.Config{})
	require.NoError(t, err)

	router := apiRouter.NewRouter(db, *cfg, log)

	// 1) Create workspace
	wsBody := map[string]any{"name": "Test Workspace"}
	wsBytes, err := json.Marshal(wsBody)
	require.NoError(t, err)

	wsReq := httptest.NewRequest(http.MethodPost, "/api/v1/notion/workspaces", bytes.NewReader(wsBytes))
	wsReq.Header.Set("Content-Type", "application/json")
	wsRec := httptest.NewRecorder()
	router.ServeHTTP(wsRec, wsReq)
	require.Equal(t, http.StatusCreated, wsRec.Code)

	var wsCreated struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(wsRec.Body.Bytes(), &wsCreated))

	var workspace struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(wsCreated.Data, &workspace))

	// 2) Create page
	pageBody := map[string]any{
		"workspaceId": workspace.ID,
		"title":       "Test Page",
	}
	pageBytes, err := json.Marshal(pageBody)
	require.NoError(t, err)

	pageReq := httptest.NewRequest(http.MethodPost, "/api/v1/notion/pages", bytes.NewReader(pageBytes))
	pageReq.Header.Set("Content-Type", "application/json")
	pageRec := httptest.NewRecorder()
	router.ServeHTTP(pageRec, pageReq)
	require.Equal(t, http.StatusCreated, pageRec.Code)

	var pageCreated struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(pageRec.Body.Bytes(), &pageCreated))

	var page struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(pageCreated.Data, &page))

	// 3) Create first block
	blockBody := map[string]any{
		"pageId":   page.ID,
		"type":     "paragraph",
		"content":  "Hello world",
		"position": 0,
	}
	blockBytes, err := json.Marshal(blockBody)
	require.NoError(t, err)

	blockReq := httptest.NewRequest(http.MethodPost, "/api/v1/notion/blocks", bytes.NewReader(blockBytes))
	blockReq.Header.Set("Content-Type", "application/json")
	blockRec := httptest.NewRecorder()
	router.ServeHTTP(blockRec, blockReq)
	require.Equal(t, http.StatusCreated, blockRec.Code, "expected block to be created")

	var blockCreated struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(blockRec.Body.Bytes(), &blockCreated))

	var block struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Content  string `json:"content"`
		Position int    `json:"position"`
	}
	require.NoError(t, json.Unmarshal(blockCreated.Data, &block))
	require.Equal(t, "paragraph", block.Type)
	require.Equal(t, "Hello world", block.Content)
	require.Equal(t, 0, block.Position)

	// 4) Create second block
	block2Body := map[string]any{
		"pageId":   page.ID,
		"type":     "heading",
		"content":  "Section Title",
		"position": 1,
	}
	block2Bytes, err := json.Marshal(block2Body)
	require.NoError(t, err)

	block2Req := httptest.NewRequest(http.MethodPost, "/api/v1/notion/blocks", bytes.NewReader(block2Bytes))
	block2Req.Header.Set("Content-Type", "application/json")
	block2Rec := httptest.NewRecorder()
	router.ServeHTTP(block2Rec, block2Req)
	require.Equal(t, http.StatusCreated, block2Rec.Code)

	var block2Created struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(block2Rec.Body.Bytes(), &block2Created))

	var block2 struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(block2Created.Data, &block2))

	// 5) List blocks for page
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/notion/blocks?pageId="+page.ID, nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var listResp struct {
		Success bool              `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &listResp))
	require.GreaterOrEqual(t, len(listResp.Data), 2, "should have at least 2 blocks")

	// 6) Update block content (last-write-wins test)
	updateBody := map[string]any{
		"content": "Updated content",
	}
	updateBytes, err := json.Marshal(updateBody)
	require.NoError(t, err)

	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/notion/blocks/"+block.ID, bytes.NewReader(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	require.Equal(t, http.StatusOK, updateRec.Code)

	var updateResp struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &updateResp))

	var updatedBlock struct {
		Content string `json:"content"`
	}
	require.NoError(t, json.Unmarshal(updateResp.Data, &updatedBlock))
	require.Equal(t, "Updated content", updatedBlock.Content)

	// 7) Delete block
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/notion/blocks/"+block2.ID, nil)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	require.Equal(t, http.StatusNoContent, deleteRec.Code)

	// 8) Verify block is deleted
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/notion/blocks/"+block2.ID, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	// Should return 404 or error indicating block not found
	require.NotEqual(t, http.StatusOK, getRec.Code, "deleted block should not be retrievable")
}
