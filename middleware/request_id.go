package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDKey 请求ID在上下文中的key
	RequestIDKey = "request_id"
	// TraceIDKey 追踪ID在上下文中的key
	TraceIDKey = "trace_id"
)

// RequestIDMiddleware 请求ID中间件
// 为每个请求生成唯一的请求ID，用于日志追踪和问题排查
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID（支持分布式追踪）
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 如果没有，生成新的UUID
			requestID = uuid.New().String()
		}

		// 尝试从请求头获取追踪ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			// 如果没有，使用请求ID作为追踪ID
			traceID = requestID
		}

		// 将请求ID和追踪ID存储到上下文中
		c.Set(RequestIDKey, requestID)
		c.Set(TraceIDKey, traceID)

		// 将请求ID添加到响应头，方便客户端追踪
		c.Header("X-Request-ID", requestID)
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}
