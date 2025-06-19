package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"shortlink-platform/backend/internal/apperrors"
	"shortlink-platform/backend/internal/dto"
	"shortlink-platform/backend/internal/repository"
	"shortlink-platform/backend/internal/service"
	"shortlink-platform/backend/pkg/logging"
	"shortlink-platform/backend/response"
	"strconv"
)

func CreateShortLinkHandler(c *gin.Context) {

	var req dto.CreateShortLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		// 记录请求上下文（方法、路径、原始请求体）
		zap.L().Warn("Request body binding failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		//显式忽略返回值
		_ = c.Error(apperrors.InvalidRequestErrorDefault())
		return
	}

	zap.L().Info("Request Headers", zap.Any("headers", c.Request.Header))

	if err := service.CreateShortLink(req); err != nil {
		// 记录关键业务参数和错误上下文
		zap.L().Warn("Short chain creation failed",
			zap.Error(err),
			zap.String("short_code", req.ShortCode),
		)
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.OK("", "Short chain creation successful"))
}

// ListShortLinksHandler 分页查询短链列表
func ListShortLinksHandler(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	shortCode := c.Query("shortCode")

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

	// 调用服务层
	pageResp, err := service.ListShortLinks(page, size, shortCode)
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
		_ = c.Error(apperrors.InvalidRequestError("无效的 ID"))
		return
	}

	// 解析请求体
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.InvalidRequestError("请求体格式错误"))
		return
	}

	// 校验 status 值
	if req.Status != 0 && req.Status != 1 {
		_ = c.Error(apperrors.InvalidRequestError("status 必须为 0 或 1"))
		return
	}

	// 调用服务层
	if err := service.UpdateShortLinkStatus(uint(id), req.Status == 1); err != nil {
		_ = c.Error(err)
		return
	}

	// 构造响应
	c.JSON(http.StatusOK, response.OK(struct{}{}, "短链状态已更新"))
}

// UpdateShortLinkHandler

func UpdateShortLinkHandler(c *gin.Context) {
	// 1. 从 URL 路径中提取短链 ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		zap.L().Warn("Invalid short link ID",
			zap.String("id", idStr),
			zap.Error(err))
		_ = c.Error(apperrors.BusinessError(http.StatusBadRequest, "无效的短链 ID"))
		return
	}

	// 2. 绑定请求体到 DTO
	var req dto.UpdateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 记录请求上下文
		zap.L().Warn("Request body binding failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		_ = c.Error(apperrors.InvalidRequestErrorDefault())
		return
	}

	// 4. 调用服务层更新逻辑
	if err := service.UpdateShortLink(uint(id), req.TargetURL); err != nil {
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
