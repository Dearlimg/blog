package logger

import (
	"blog/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetRequestID 从 gin.Context 中获取 request_id
// 如果不存在，返回空字符串
func GetRequestID(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, exists := ctx.Get(middleware.RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceID 从 gin.Context 中获取 trace_id
// 如果不存在，返回空字符串
func GetTraceID(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, exists := ctx.Get(middleware.TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// WithRequestID 为日志添加 request_id 字段
// 如果 request_id 存在，会自动添加到日志字段中
func WithRequestID(ctx *gin.Context, fields ...zap.Field) []zap.Field {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		fields = append(fields, String("request_id", requestID))
	}
	traceID := GetTraceID(ctx)
	if traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}
	return fields
}

// DebugWithCtx 带上下文的调试日志（自动包含 request_id）
func DebugWithCtx(ctx *gin.Context, msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, WithRequestID(ctx, fields...)...)
}

// InfoWithCtx 带上下文的信息日志（自动包含 request_id）
func InfoWithCtx(ctx *gin.Context, msg string, fields ...zap.Field) {
	GetLogger().Info(msg, WithRequestID(ctx, fields...)...)
}

// WarnWithCtx 带上下文的警告日志（自动包含 request_id）
func WarnWithCtx(ctx *gin.Context, msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, WithRequestID(ctx, fields...)...)
}

// ErrorWithCtx 带上下文的错误日志（自动包含 request_id）
func ErrorWithCtx(ctx *gin.Context, msg string, fields ...zap.Field) {
	GetLogger().Error(msg, WithRequestID(ctx, fields...)...)
}
