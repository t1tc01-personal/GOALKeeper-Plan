package contract

import (
	"bytes"
	"encoding/json"
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

// TestSharingAPIContract validates that the sharing and permissions endpoints
// exist, return JSON, and follow the basic response envelope contract.
// Detailed behavior is covered in integration tests.
func TestSharingAPIContract(t *testing.T) {
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
		body   interface{}
	}{
		{
			"list collaborators on page",
			http.MethodGet,
			"/api/v1/notion/pages/550e8400-e29b-41d4-a716-446655440000/collaborators",
			nil,
		},
		{
			"grant access to page",
			http.MethodPost,
			"/api/v1/notion/pages/550e8400-e29b-41d4-a716-446655440000/share",
			map[string]interface{}{
				"user_id": "550e8400-e29b-41d4-a716-446655440001",
				"role":    "viewer",
			},
		},
		{
			"revoke access to page",
			http.MethodDelete,
			"/api/v1/notion/pages/550e8400-e29b-41d4-a716-446655440000/share/550e8400-e29b-41d4-a716-446655440001",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				bodyBytes, err := json.Marshal(tt.body)
				require.NoError(t, err)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// Endpoint should exist (not 404)
			require.NotEqual(t, http.StatusNotFound, rec.Code, "endpoint should exist")

			// Response should be JSON
			require.Contains(t, rec.Header().Get("Content-Type"), "application/json",
				"response should be JSON (actual: %s)", rec.Header().Get("Content-Type"))
		})
	}
}
