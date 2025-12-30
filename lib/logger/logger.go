package logger

import (
	"context"

	"go.uber.org/zap"
)

type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

//go:generate mockgen -destination=mocks/mock_logger.go -package=mocks care-cordination/lib/logger Logger
type Logger interface {
	Debug(ctx context.Context, operation string, message string, fields ...zap.Field)
	Info(ctx context.Context, operation string, message string, fields ...zap.Field)
	Warn(ctx context.Context, operation string, message string, fields ...zap.Field)
	Error(ctx context.Context, operation string, message string, fields ...zap.Field)
	ZapLogger() *zap.Logger
}

type logger struct {
	Logger *zap.Logger
}

func NewLogger(env string) Logger {
	var config zap.Config
	if env == "production" {
		config = zap.NewProductionConfig()
		config.DisableStacktrace = true
		config.OutputPaths = []string{"stdout"}
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		config = zap.NewDevelopmentConfig()
		config.OutputPaths = []string{"stdout"}
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapLogger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return &logger{
		Logger: zapLogger,
	}
}

func (l *logger) getCommonFields(
	ctx context.Context,
	event string,
	operation string,
	fields ...zap.Field,
) []zap.Field {
	requestID := "unknown"
	if v, ok := ctx.Value("X-Request-Id").(string); ok {
		requestID = v
	}
	commonFields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("event", event),
		zap.String("operation", operation),
	}
	commonFields = append(commonFields, fields...)

	return commonFields
}

// Convenience methods for simpler logging
func (l *logger) Debug(ctx context.Context, operation string, message string, fields ...zap.Field) {
	fields = l.getCommonFields(ctx, "app", operation, fields...)
	l.Logger.Debug(message, fields...)
}

func (l *logger) Info(ctx context.Context, operation string, message string, fields ...zap.Field) {
	fields = l.getCommonFields(ctx, "app", operation, fields...)
	l.Logger.Info(message, fields...)
}

func (l *logger) Warn(ctx context.Context, operation string, message string, fields ...zap.Field) {
	fields = l.getCommonFields(ctx, "app", operation, fields...)
	l.Logger.Warn(message, fields...)
}

func (l *logger) Error(ctx context.Context, operation string, message string, fields ...zap.Field) {
	fields = l.getCommonFields(ctx, "app", operation, fields...)
	l.Logger.Error(message, fields...)
}

func (l *logger) ZapLogger() *zap.Logger {
	return l.Logger
}
