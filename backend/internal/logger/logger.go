package logger

import (
	"goalkeeper-plan/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Flush(timeoutSeconds int) error
}

type logger struct {
	zap *zap.Logger
}

func NewLogger(cfg config.AppConfig) (Logger, error) {
	var zapConfig zap.Config

	if cfg.Debug {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &logger{zap: zapLogger}, nil
}

func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

func (l *logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

func (l *logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

func (l *logger) Flush(timeoutSeconds int) error {
	return l.zap.Sync()
}

