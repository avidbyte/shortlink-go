package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"shortlink-go/internal/apperrors"
	"shortlink-go/response"
)

// GlobalErrorMiddleware 全局错误中间件
func GlobalErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果有错误发生
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				var appErr *apperrors.AppError
				if errors.As(err.Err, &appErr) {
					c.AbortWithStatusJSON(appErr.Code, response.ErrorFromAppError(appErr))
					return
				}
			}

			// 默认处理未定义的错误
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Error("系统内部错误"))
			return
		}
	}
}
