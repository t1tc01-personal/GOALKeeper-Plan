package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP request metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestSize       *prometheus.HistogramVec
	HTTPResponseSize      *prometheus.HistogramVec

	// Error metrics
	HTTPErrorsTotal       *prometheus.CounterVec
	ErrorRateByType       *prometheus.CounterVec

	// Page load metrics
	PageLoadLatency       *prometheus.HistogramVec
	PageLoadTotal         *prometheus.CounterVec

	// Permission metrics
	PermissionChecksTotal *prometheus.CounterVec
	PermissionFailures    *prometheus.CounterVec
	PermissionCheckLatency *prometheus.HistogramVec

	// Database metrics
	DatabaseQueryDuration *prometheus.HistogramVec
	DatabaseErrorsTotal   *prometheus.CounterVec
}

// NewMetrics creates a new Metrics instance with all Prometheus metrics registered
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP request metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status_code"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to 100MB
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to 100MB
			},
			[]string{"method", "path", "status_code"},
		),

		// Error metrics
		HTTPErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP errors",
			},
			[]string{"method", "path", "status_code", "error_type"},
		),
		ErrorRateByType: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "error_rate_by_type_total",
				Help: "Error rate by error type",
			},
			[]string{"error_type", "error_code"},
		),

		// Page load metrics
		PageLoadLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "page_load_latency_seconds",
				Help:    "Page load latency in seconds",
				Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.0, 5.0, 10.0}, // Optimized for page loads
			},
			[]string{"page_id", "user_id"},
		),
		PageLoadTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "page_load_total",
				Help: "Total number of page loads",
			},
			[]string{"page_id", "status"},
		),

		// Permission metrics
		PermissionChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "permission_checks_total",
				Help: "Total number of permission checks",
			},
			[]string{"resource_type", "action", "result"},
		),
		PermissionFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "permission_failures_total",
				Help: "Total number of permission failures",
			},
			[]string{"resource_type", "action", "reason"},
		),
		PermissionCheckLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "permission_check_latency_seconds",
				Help:    "Permission check latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5}, // Optimized for permission checks
			},
			[]string{"resource_type", "action"},
		),

		// Database metrics
		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
			},
			[]string{"operation", "table"},
		),
		DatabaseErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_errors_total",
				Help: "Total number of database errors",
			},
			[]string{"operation", "table", "error_type"},
		),
	}
}

// RecordHTTPRequest records metrics for an HTTP request
func (m *Metrics) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	statusCodeStr := strconv.Itoa(statusCode)

	labels := prometheus.Labels{
		"method":      method,
		"path":        path,
		"status_code": statusCodeStr,
	}

	m.HTTPRequestsTotal.With(labels).Inc()
	m.HTTPRequestDuration.With(labels).Observe(duration.Seconds())
	m.HTTPResponseSize.With(labels).Observe(float64(responseSize))

	requestLabels := prometheus.Labels{
		"method": method,
		"path":   path,
	}
	m.HTTPRequestSize.With(requestLabels).Observe(float64(requestSize))

	// Record error if status code >= 400
	if statusCode >= 400 {
		errorLabels := prometheus.Labels{
			"method":      method,
			"path":        path,
			"status_code": statusCodeStr,
			"error_type":  getErrorType(statusCode),
		}
		m.HTTPErrorsTotal.With(errorLabels).Inc()
	}
}

// RecordPageLoad records metrics for a page load
func (m *Metrics) RecordPageLoad(pageID, userID string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	labels := prometheus.Labels{
		"page_id": pageID,
		"user_id": userID,
	}

	m.PageLoadLatency.With(labels).Observe(duration.Seconds())

	loadLabels := prometheus.Labels{
		"page_id": pageID,
		"status":  status,
	}
	m.PageLoadTotal.With(loadLabels).Inc()
}

// RecordPermissionCheck records metrics for a permission check
func (m *Metrics) RecordPermissionCheck(resourceType, action string, hasPermission bool, duration time.Duration, reason string) {
	result := "allowed"
	if !hasPermission {
		result = "denied"
	}

	labels := prometheus.Labels{
		"resource_type": resourceType,
		"action":        action,
		"result":        result,
	}

	m.PermissionChecksTotal.With(labels).Inc()
	m.PermissionCheckLatency.With(prometheus.Labels{
		"resource_type": resourceType,
		"action":        action,
	}).Observe(duration.Seconds())

	if !hasPermission {
		failureLabels := prometheus.Labels{
			"resource_type": resourceType,
			"action":        action,
			"reason":        reason,
		}
		m.PermissionFailures.With(failureLabels).Inc()
	}
}

// RecordError records an error by type
func (m *Metrics) RecordError(errorType, errorCode string) {
	labels := prometheus.Labels{
		"error_type": errorType,
		"error_code": errorCode,
	}
	m.ErrorRateByType.With(labels).Inc()
}

// RecordDatabaseQuery records metrics for a database query
func (m *Metrics) RecordDatabaseQuery(operation, table string, duration time.Duration, err error) {
	labels := prometheus.Labels{
		"operation": operation,
		"table":     table,
	}
	m.DatabaseQueryDuration.With(labels).Observe(duration.Seconds())

	if err != nil {
		errorLabels := prometheus.Labels{
			"operation":  operation,
			"table":      table,
			"error_type": getDatabaseErrorType(err),
		}
		m.DatabaseErrorsTotal.With(errorLabels).Inc()
	}
}

// getErrorType categorizes HTTP status codes into error types
func getErrorType(statusCode int) string {
	switch {
	case statusCode >= 400 && statusCode < 500:
		return "client_error"
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}

// getDatabaseErrorType categorizes database errors
func getDatabaseErrorType(err error) string {
	if err == nil {
		return "none"
	}
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "connection"):
		return "connection"
	case contains(errStr, "constraint"):
		return "constraint"
	case contains(errStr, "not found"):
		return "not_found"
	default:
		return "unknown"
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

