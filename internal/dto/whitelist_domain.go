package dto

type CreateWhiteListDomainRequest struct {
	Domain string `json:"domain" binding:"required,url" msg:"Domain must be a valid URL"` // Gin 内置 URL 校验
}
