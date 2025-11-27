package logger

import (
	"go.uber.org/zap"
)

// 提供便捷的日志方法，封装常用的日志操作

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 致命错误日志（会退出程序）
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// 常用的 Field 辅助函数

// String 字符串字段
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int 整数字段
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Int32 32位整数字段
func Int32(key string, val int32) zap.Field {
	return zap.Int32(key, val)
}

// Int64 64位整数字段
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Float64 浮点数字段
func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

// Bool 布尔字段
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Any 任意类型字段
func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

// ErrorField 错误字段（常用）
func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

// Duration 时间间隔字段
func Duration(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}
