package unit

import "go.uber.org/zap"

// NoopLogger is a logger that discards all messages for testing
type NoopLogger struct{}

func (n *NoopLogger) Debug(msg string, fields ...zap.Field) {}
func (n *NoopLogger) Info(msg string, fields ...zap.Field)  {}
func (n *NoopLogger) Warn(msg string, fields ...zap.Field)  {}
func (n *NoopLogger) Error(msg string, fields ...zap.Field) {}
func (n *NoopLogger) Fatal(msg string, fields ...zap.Field) {}
func (n *NoopLogger) Flush(timeoutSeconds int) error        { return nil }

// NewNoopLogger creates a new noop logger for testing
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}
