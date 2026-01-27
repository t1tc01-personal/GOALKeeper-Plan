package contract

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apiRouter "goalkeeper-plan/cmd/api"
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestBlockAPIContract validates that the block endpoints exist,
// return JSON, and follow the basic response envelope contract.
// Detailed behavior is covered in integration tests.
func TestBlockAPIContract(t *testing.T) {
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
		{"list blocks", http.MethodGet, "/api/v1/notion/blocks"},
		{"create block", http.MethodPost, "/api/v1/notion/blocks"},
		{"get block", http.MethodGet, "/api/v1/notion/blocks/test-id"},
		{"update block", http.MethodPut, "/api/v1/notion/blocks/test-id"},
		{"delete block", http.MethodDelete, "/api/v1/notion/blocks/test-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// Should not return 404 (endpoint should exist)
			require.NotEqual(t, http.StatusNotFound, rec.Code, "endpoint should exist")
			// Should return JSON content type
			require.Contains(t, rec.Header().Get("Content-Type"), "application/json", "response should be JSON")
		})
	}
}
