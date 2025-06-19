package apperrors

import (
	"net/http"
)

// AppError 自定义错误类型
type AppError struct {
	Code    int
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithCode 创建通用业务错误
func WithCode(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// BusinessError 封装业务逻辑错误（通用）
func BusinessError(code int, message string) *AppError {
	return WithCode(code, message)
}

// InvalidRequestError 封装参数校验错误
func InvalidRequestError(message string) *AppError {
	return WithCode(http.StatusBadRequest, message)
}

// InvalidRequestErrorDefault 默认参数校验错误
func InvalidRequestErrorDefault() *AppError {
	return WithCode(http.StatusBadRequest, "Parameter verification failed")
}

// SystemError 封装系统内部错误
func SystemError(message string) *AppError {
	return WithCode(http.StatusInternalServerError, message)
}

// SystemErrorDefault 默认系统内部错误
func SystemErrorDefault() *AppError {
	return WithCode(http.StatusInternalServerError, "System error")
}
