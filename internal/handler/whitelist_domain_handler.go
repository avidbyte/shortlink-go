package handler

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"shortlink-platform/backend/internal/apperrors"
	"shortlink-platform/backend/internal/dto"
	"shortlink-platform/backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"shortlink-platform/backend/response"
)

// CreateWhitelistDomainHandler 创建白名单域名（POST /handler/whitelist）
func CreateWhitelistDomainHandler(c *gin.Context) {

	var req dto.CreateWhiteListDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 检查错误是否为 ValidationErrors 类型
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			// 遍历所有校验错误
			for _, e := range validationErrs {
				// 通过反射获取字段的 msg 标签值
				field, ok := reflect.TypeOf(req).FieldByName(e.Field())
				if !ok {
					_ = c.Error(apperrors.InvalidRequestErrorDefault())
					return
				}

				customMsg := field.Tag.Get("msg")
				if customMsg != "" {
					_ = c.Error(apperrors.InvalidRequestError(customMsg))
					return
				}
			}
		}
		// 如果没有找到自定义错误提示，返回默认错误
		_ = c.Error(apperrors.InvalidRequestErrorDefault())
		return
	}

	if err := service.CreateWhitelistDomain(req.Domain); err != nil {
		// 记录关键业务参数和错误上下文
		zap.L().Warn("whitelist domain creation failed",
			zap.Error(err),
			zap.String("domain", req.Domain),
		)
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.OK("", "域名已添加至白名单"))
}

// ListWhitelistDomainsHandler 分页查询白名单（GET /handler/whitelist?domain=xxx&page=1&size=10）
func ListWhitelistDomainsHandler(c *gin.Context) {
	// 1. 解析查询参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	domain := c.DefaultQuery("domain", "")

	// 参数转换
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		_ = c.Error(apperrors.InvalidRequestError("页码必须为正整数"))
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		_ = c.Error(apperrors.InvalidRequestError("每页数量必须为1-100之间的整数"))
		return
	}

	// 3. 调用服务层查询
	pageResp, err := service.ListWhitelistDomains(page, size, domain)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// 构造响应
	c.JSON(http.StatusOK, response.OK(pageResp, "success"))
}

// DeleteWhitelistDomainHandler 删除白名单域名（DELETE /handler/whitelist/:id）
func DeleteWhitelistDomainHandler(c *gin.Context) {
	// 1. 解析路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Error("无效的ID"))
		return
	}

	// 2. 调用服务层删除
	if err := service.DeleteWhitelistDomain(uint(id)); err != nil {
		if err != nil {
			_ = c.Error(err)
			return
		}
	}
	// 3. 返回成功响应
	c.JSON(http.StatusOK, response.OK("", "白名单域名已删除"))
}
