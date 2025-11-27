package errcode

import "fmt"

// Err 错误接口
type Err interface {
	Code() int
	Msg() string
	Details() string
	WithDetails(details string) Err
	Error() string
}

// errCode 错误码结构体
type errCode struct {
	code    int
	message string
	details string
}

// Code 返回错误码
func (e *errCode) Code() int {
	return e.code
}

// Msg 返回错误消息
func (e *errCode) Msg() string {
	return e.message
}

// Details 返回错误详情
func (e *errCode) Details() string {
	return e.details
}

// WithDetails 添加错误详情
func (e *errCode) WithDetails(details string) Err {
	return &errCode{
		code:    e.code,
		message: e.message,
		details: details,
	}
}

// Error 实现error接口
func (e *errCode) Error() string {
	if e.details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.code, e.message, e.details)
	}
	return fmt.Sprintf("[%d] %s", e.code, e.message)
}

// NewErr 创建新的错误码
func NewErr(code int, message string) Err {
	return &errCode{
		code:    code,
		message: message,
		details: "",
	}
}

// 预定义错误码
var (
	// 通用错误码 1000-1999
	ErrSuccess = NewErr(0, "success")
	ErrServer  = NewErr(1000, "internal server error")
	ErrUnknown = NewErr(1001, "unknown error")

	// 参数错误 2000-2999
	ErrParamsNotValid = NewErr(2000, "invalid parameters")
	ErrParamsMissing  = NewErr(2001, "missing required parameters")
	ErrParamsFormat   = NewErr(2002, "parameter format error")

	// 业务错误 3000-3999
	ErrResourceNotFound = NewErr(3000, "resource not found")
	ErrResourceExists   = NewErr(3001, "resource already exists")
	ErrOperationFailed  = NewErr(3002, "operation failed")

	// 认证授权错误 4000-4999
	ErrUnauthorized = NewErr(4000, "unauthorized")
	ErrForbidden    = NewErr(4001, "forbidden")
	ErrTokenInvalid = NewErr(4002, "invalid token")
	ErrTokenExpired = NewErr(4003, "token expired")

	// 数据库错误 5000-5999
	ErrDatabase = NewErr(5000, "database error")
	ErrCache    = NewErr(5001, "cache error")
)
