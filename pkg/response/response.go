package response

import (
	"blog/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构体
type Response struct {
	Code      int         `json:"code"`                 // 业务状态码，0表示成功
	Message   string      `json:"message"`              // 响应消息
	Data      interface{} `json:"data,omitempty"`       // 响应数据，可选
	Timestamp int64       `json:"timestamp"`            // 时间戳（Unix时间戳，秒）
	RequestID string      `json:"request_id,omitempty"` // 请求ID，用于追踪
	TraceID   string      `json:"trace_id,omitempty"`   // 追踪ID，用于分布式追踪
}

// ResponseWriter 响应写入器
type ResponseWriter struct {
	ctx *gin.Context
}

// NewResponse 创建新的响应写入器
func NewResponse(ctx *gin.Context) *ResponseWriter {
	return &ResponseWriter{ctx: ctx}
}

// getRequestID 从上下文中获取请求ID
func (r *ResponseWriter) getRequestID() string {
	if requestID, exists := r.ctx.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// getTraceID 从上下文中获取追踪ID
func (r *ResponseWriter) getTraceID() string {
	if traceID, exists := r.ctx.Get("trace_id"); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// Success 成功响应
func (r *ResponseWriter) Success(data interface{}) {
	r.JSON(http.StatusOK, 0, "success", data)
}

// SuccessWithMessage 带自定义消息的成功响应
func (r *ResponseWriter) SuccessWithMessage(message string, data interface{}) {
	r.JSON(http.StatusOK, 0, message, data)
}

// Error 错误响应
func (r *ResponseWriter) Error(code int, message string) {
	r.JSON(http.StatusOK, code, message, nil)
}

// ErrorWithData 带数据的错误响应
func (r *ResponseWriter) ErrorWithData(code int, message string, data interface{}) {
	r.JSON(http.StatusOK, code, message, data)
}

// BadRequest 400错误响应
func (r *ResponseWriter) BadRequest(message string) {
	r.JSON(http.StatusBadRequest, 400, message, nil)
}

// Unauthorized 401错误响应
func (r *ResponseWriter) Unauthorized(message string) {
	r.JSON(http.StatusUnauthorized, 401, message, nil)
}

// Forbidden 403错误响应
func (r *ResponseWriter) Forbidden(message string) {
	r.JSON(http.StatusForbidden, 403, message, nil)
}

// NotFound 404错误响应
func (r *ResponseWriter) NotFound(message string) {
	r.JSON(http.StatusNotFound, 404, message, nil)
}

// InternalServerError 500错误响应
func (r *ResponseWriter) InternalServerError(message string) {
	r.JSON(http.StatusInternalServerError, 500, message, nil)
}

// JSON 统一的JSON响应方法
func (r *ResponseWriter) JSON(httpStatus, code int, message string, data interface{}) {
	response := Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: r.getRequestID(),
		TraceID:   r.getTraceID(),
	}

	// 记录响应日志
	if code != 0 {
		logger.Warn("API response error",
			logger.String("path", r.ctx.Request.URL.Path),
			logger.String("method", r.ctx.Request.Method),
			logger.Int("code", code),
			logger.String("message", message),
			logger.String("request_id", response.RequestID),
		)
	} else {
		logger.Debug("API response success",
			logger.String("path", r.ctx.Request.URL.Path),
			logger.String("method", r.ctx.Request.Method),
			logger.String("request_id", response.RequestID),
		)
	}

	r.ctx.JSON(httpStatus, response)
}

// Reply 兼容原有接口的响应方法
// 支持 errcode.Err 类型的错误
func (r *ResponseWriter) Reply(err interface{}, data ...interface{}) {
	// 如果第一个参数是错误
	if err != nil {
		// 尝试转换为 errcode.Err 类型
		if errCode, ok := err.(interface {
			Code() int
			Msg() string
			Details() string
		}); ok {
			message := errCode.Msg()
			if details := errCode.Details(); details != "" {
				message = message + ": " + details
			}
			r.Error(errCode.Code(), message)
			return
		}

		// 尝试转换为标准error类型
		if errStd, ok := err.(error); ok {
			r.InternalServerError(errStd.Error())
			return
		}
	}

	// 没有错误，返回成功
	var responseData interface{}
	if len(data) > 0 {
		responseData = data[0]
	}
	r.Success(responseData)
}
