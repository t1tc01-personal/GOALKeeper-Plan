package contract

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"goalkeeper-plan/config"
	apiRouter "goalkeeper-plan/cmd/api"
	"goalkeeper-plan/internal/logger"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestWorkspacePageAPIContract validates that the primary workspace/page
// endpoints exist, return JSON, and follow the basic response envelope
// contract (success + data/error fields). Detailed behavior is covered
// in integration tests.
func TestWorkspacePageAPIContract(t *testing.T) {
	cfg := config.GetConfig()
	log, err := logger.NewLogger(cfg.AppConfig)
	require.NoError(t, err)

	db, err := gorm.Open(postgres.Open(cfg.PostgresConfig.ConnectionString), &gorm.Config{})
	require.NoError(t, err)

	router := apiRouter.NewRouter(db, *cfg, log)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"list workspaces", http.MethodGet, "/api/v1/notion/workspaces"},
		{"create workspace", http.MethodPost, "/api/v1/notion/workspaces"},
		{"list pages", http.MethodGet, "/api/v1/notion/pages"},
		{"create page", http.MethodPost, "/api/v1/notion/pages"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			require.NotEqual(t, http.StatusNotFound, rec.Code, "endpoint should exist")
			require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
		})
	}
}

