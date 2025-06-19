package dto

import (
	"github.com/gin-gonic/gin"
	"shortlink-platform/backend/pkg/utils"
)

// CreateShortLinkRequest 用于创建短链的请求参数
type CreateShortLinkRequest struct {
	TargetURL    string `json:"targetUrl" binding:"required,url"` // Gin 内置 URL 校验
	ShortCode    string `json:"shortCode" binding:"required,max=32"`
	RedirectCode int    `json:"redirectCode" binding:"required,oneof=301 302"` // 仅允许301/302
	// 最大长度 32（Gin 限制）
	Disabled bool `json:"disabled" `
}

// UpdateShortLinkRequest 用于更新短链的请求参数
type UpdateShortLinkRequest struct {
	TargetURL string `json:"targetUrl" binding:"required,url" msg:"targetUrl must be a valid URL"` // 必填字段，Gin 内置 URL 校验
}

// Validate 自定义验证逻辑
func (r *CreateShortLinkRequest) Validate() error {
	// 1. 复用公共的 TargetURL 校验逻辑
	if err := utils.ValidateTargetURL(r.TargetURL); err != nil {
		return gin.Error{
			Err:  err,
			Type: gin.ErrorTypeBind,
		}
	}

	// 2. 调用独立的 ShortCode 校验方法
	if err := utils.ValidateShortCode(r.ShortCode); err != nil {
		return gin.Error{
			Err:  err,
			Type: gin.ErrorTypeBind,
		}
	}

	return nil
}
