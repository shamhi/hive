package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	loggerRequestIdKey contextKey = "x-request-id"
	loggerTraceIdKey   contextKey = "x-trace-id"
)

// Logger interface
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}

type L struct {
	z *zap.Logger
}

func NewLogger(mode string) (Logger, error) {
	var zCfg zap.Config

	if mode == "local" || mode == "dev" {
		zCfg = zap.NewDevelopmentConfig()
		zCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Красивые цвета в консоли
	} else {
		zCfg = zap.NewProductionConfig()
		zCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	zCfg.EncoderConfig.TimeKey = "timestamp"
	zCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	z, err := zCfg.Build()
	if err != nil {
		return nil, err
	}

	return &L{z: z}, nil
}

func (l *L) Sync() error {
	return l.z.Sync()
}

func (l *L) commonFields(ctx context.Context, fields []zap.Field) []zap.Field {
	newFields := make([]zap.Field, 0, len(fields)+2)
	newFields = append(newFields, fields...)

	if reqID := getStringFromKey(ctx, loggerRequestIdKey); reqID != "" {
		newFields = append(newFields, zap.String(string(loggerRequestIdKey), reqID))
	}

	if traceID := getStringFromKey(ctx, loggerTraceIdKey); traceID != "" {
		newFields = append(newFields, zap.String(string(loggerTraceIdKey), traceID))
	}

	return newFields
}

func (l *L) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Debug(msg, l.commonFields(ctx, fields)...)
}

func (l *L) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Info(msg, l.commonFields(ctx, fields)...)
}

func (l *L) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Warn(msg, l.commonFields(ctx, fields)...)
}

func (l *L) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Error(msg, l.commonFields(ctx, fields)...)
}

func (l *L) With(fields ...zap.Field) Logger {
	return &L{z: l.z.With(fields...)}
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, loggerRequestIdKey, requestID)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, loggerTraceIdKey, traceID)
}

func getStringFromKey(ctx context.Context, key contextKey) string {
	if value := ctx.Value(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}
