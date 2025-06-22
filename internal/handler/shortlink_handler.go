package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"shortlink-go/internal/apperrors"
	"shortlink-go/internal/dto"
	"shortlink-go/internal/i18n"
	"shortlink-go/internal/repository"
	"shortlink-go/internal/service"
	"shortlink-go/pkg/logging"
	"shortlink-go/response"
	"strconv"
)

func CreateShortLinkHandler(c *gin.Context) {

	var req dto.CreateShortLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		// 记录请求上下文（方法、路径、原始请求体）
		logging.Logger.Warn("Request body binding failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		//显式忽略返回值
		//message := i18n.T(c.Request.Context(), "error.invalid_request", nil)
		_ = c.Error(apperrors.InvalidRequestErrorDefault())
		return
	}

	zap.L().Info("Request Headers", zap.Any("headers", c.Request.Header))

	if err := service.CreateShortLink(c.Request.Context(), req); err != nil {
		// 记录关键业务参数和错误上下文
		logging.Logger.Warn("Short chain creation failed",
			zap.Error(err),
			zap.String("short_code", req.ShortCode),
		)
		_ = c.Error(err)
		return
	}
	message := i18n.T(c.Request.Context(), "success.short_link_created", nil)
	c.JSON(http.StatusOK, response.OK("", message))
}

// ListShortLinksHandler 分页查询短链列表
func ListShortLinksHandler(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	shortCode := c.Query("shortCode")
	targetUrl := c.Query("targetUrl")

	// 获取 redirectCode，并转换为 int
	redirectCodeStr := c.Query("redirectCode")
	var redirectCode int
	if redirectCodeStr != "" {
		var err error
		redirectCode, err = strconv.Atoi(redirectCodeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirectCode"})
			return
		}
	}

	// 获取 disabled，并转换为 bool
	disabledStr := c.Query("disabled")
	var disabled *bool // 用指针以区分“未传”和“传了 false”
	if disabledStr != "" {
		value, err := strconv.ParseBool(disabledStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid disabled"})
			return
		}
		disabled = &value
	}

	// 参数转换
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		//页码必须为正整数
		_ = c.Error(apperrors.InvalidRequestError("Page number must be a positive integer"))
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		//每页数量必须为1-100之间的整数
		_ = c.Error(apperrors.InvalidRequestError("The number of pages must be an integer between 1 and 100."))
		return
	}

	// 调用服务层
	pageResp, err := service.ListShortLinks(c.Request.Context(), page, size, shortCode, targetUrl, redirectCode, disabled)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// 构造响应
	c.JSON(http.StatusOK, response.OK(pageResp, "success"))
}

// UpdateShortLinkStatusHandler 更新短链状态（启用/禁用）
func UpdateShortLinkStatusHandler(c *gin.Context) {
	// 从 URL 获取 ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		message := i18n.T(c.Request.Context(), "error.invalid_id", nil)
		_ = c.Error(apperrors.InvalidRequestError(message))
		return
	}

	// 解析请求体
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.T(c.Request.Context(), "error.request_body_invalid", nil)
		_ = c.Error(apperrors.InvalidRequestError(message))
		return
	}

	// 校验 status 值
	if req.Status != 0 && req.Status != 1 {
		message := i18n.T(c.Request.Context(), "error.status_invalid", nil)
		_ = c.Error(apperrors.InvalidRequestError(message))
		return
	}

	// 调用服务层
	if err := service.UpdateShortLinkStatus(c.Request.Context(), uint(id), req.Status == 1); err != nil {
		_ = c.Error(err)
		return
	}
	message := i18n.T(c.Request.Context(), "success.short_link_status_updated", nil)
	c.JSON(http.StatusOK, response.OK(struct{}{}, message))
}

// UpdateShortLinkHandler

func UpdateShortLinkHandler(c *gin.Context) {
	// 1. 从 URL 路径中提取短链 ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		message := i18n.T(c.Request.Context(), "error.invalid_id", nil)
		_ = c.Error(apperrors.BusinessError(http.StatusBadRequest, message))
		return
	}

	// 2. 绑定请求体到 DTO
	var req dto.UpdateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.T(c.Request.Context(), "error.request_body_invalid", nil)
		_ = c.Error(apperrors.InvalidRequestError(message))
		return
	}

	// 4. 调用服务层更新逻辑
	if err := service.UpdateShortLink(c.Request.Context(), uint(id), req.TargetURL); err != nil {
		// 记录关键业务参数和错误上下文
		zap.L().Warn("Short chain update failed",
			zap.Error(err),
			zap.Uint("id", uint(id)),
			zap.String("target_url", req.TargetURL),
		)
		_ = c.Error(err)
		return
	}

	// 5. 返回成功响应
	c.JSON(http.StatusOK, response.OK("", "Short chain update successful"))
}

func RedirectToTargetURLHandler(c *gin.Context) {
	// 提取路径作为完整的 short_code（自动去掉前导 '/'）
	path := c.Request.URL.Path[1:] // 例如 /f/test3 → f/test3
	ip := c.ClientIP()

	// 查询缓存或数据库
	shortLink, ok := service.RedirectToTargetURL(path, ip)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	conn := repository.RedisPool.Get()

	defer func() {
		if err := conn.Close(); err != nil {
			logging.Logger.Error("Failed to close Redis connection",
				zap.Error(err),
				zap.String("operation", "close"),
				zap.String("connection_type", "redis"),
			)
		}
	}()

	// 记录访问统计
	service.RecordDailyPV(conn, path)
	service.RecordDailyUV(conn, path, ip)
	service.RecordTotalPV(conn, path)
	service.RecordTotalUV(conn, path, ip)

	// 获取目标 URL 和状态码
	redirectCode := shortLink.RedirectCode
	targetURL := shortLink.TargetURL

	// 设置响应头（仅在 302 时）
	if redirectCode == http.StatusFound {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	c.Redirect(redirectCode, targetURL)
}

func GetStats(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	stats, err := service.GetStatsByShortLinkID(uint(id))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
