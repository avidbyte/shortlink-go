package response

import (
	"shortlink-go/internal/apperrors"
	"time"
)

// Response 是一个通用的 API 响应结构
type Response[T any] struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// PageResponse 分页响应结构体
type PageResponse[T any] struct {
	Page      int `json:"page"`
	Size      int `json:"size"`
	TotalPage int `json:"totalPage"`
	Total     int `json:"total"`
	List      []T `json:"list"`
}

// OK 构造一个成功的响应
func OK[T any](data T, message string) *Response[T] {
	return &Response[T]{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// Error 构造一个失败的响应
func Error(message string) *Response[any] {
	return &Response[any]{
		Success:   false,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().UnixMilli(),
	}
}

// ErrorFromAppError 新增：基于 AppError 构造错误响应
func ErrorFromAppError(err *apperrors.AppError) *Response[any] {
	return &Response[any]{
		Success:   false,
		Message:   err.Message,
		Data:      nil,
		Timestamp: time.Now().UnixMilli(),
	}
}
