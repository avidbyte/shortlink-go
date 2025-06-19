package service

import (
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"shortlink-platform/backend/internal/apperrors"
	"shortlink-platform/backend/internal/model"
	"shortlink-platform/backend/internal/repository"
	"shortlink-platform/backend/response"
)

// CreateWhitelistDomain 创建白名单域名
func CreateWhitelistDomain(domain string) error {
	if domain == "" {
		return apperrors.BusinessError(http.StatusBadRequest, "域名不能为空")
	}

	// 2. URL 格式校验
	if _, err := url.ParseRequestURI(domain); err != nil {
		return apperrors.BusinessError(http.StatusBadRequest, "domain 格式不合法")
	}

	var existing model.WhitelistDomain
	if err := repository.DB.Where("domain = ?", domain).First(&existing).Error; err == nil {
		return apperrors.BusinessError(http.StatusBadRequest, "该域名已存在")
	}

	whitelist := &model.WhitelistDomain{
		Domain: domain,
	}
	if err := repository.DB.Create(whitelist).Error; err != nil {
		zap.L().Info("创建白名单域名失败", zap.Error(err))
		return err
	}
	return nil
}

// ListWhitelistDomains 支持分页查询白名单列表
func ListWhitelistDomains(page, size int, domain string) (*response.PageResponse[model.WhitelistDomain], error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10 // 默认每页10条，最大100条
	}

	// 构建查询条件
	db := repository.DB.Model(&model.WhitelistDomain{})
	if domain != "" {
		db = db.Where("domain LIKE ?", "%"+domain+"%")
	}

	// 查询总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, apperrors.SystemError("统计白名单记录数失败: " + err.Error())
	}

	// 如果总数为0，直接返回空结果，不执行分页查询
	if total == 0 {
		return &response.PageResponse[model.WhitelistDomain]{
			Page:      page,
			Size:      size,
			Total:     0,
			TotalPage: 0,
			List:      []model.WhitelistDomain{},
		}, nil
	}

	// 查询分页数据
	var links []model.WhitelistDomain
	if err := db.
		Limit(size).
		Offset((page - 1) * size).
		Order("id DESC").
		Find(&links).Error; err != nil {
		return nil, apperrors.SystemError("查询域名白名单失败: " + err.Error())
	}

	// 计算总页数
	totalPage := (int(total) + size - 1) / size

	return &response.PageResponse[model.WhitelistDomain]{
		Page:      page,
		Size:      size,
		Total:     int(total),
		TotalPage: totalPage,
		List:      links,
	}, nil
}

// DeleteWhitelistDomain 删除白名单域名
func DeleteWhitelistDomain(id uint) error {
	if err := repository.DB.Delete(&model.WhitelistDomain{}, id).Error; err != nil {
		return apperrors.SystemError("删除域名白名单失败: " + err.Error())
	}
	return nil
}
