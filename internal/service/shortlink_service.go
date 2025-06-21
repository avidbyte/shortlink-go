package service

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"net/http"
	"shortlink-go/constant"
	"shortlink-go/internal/apperrors"
	"shortlink-go/internal/dto"
	"shortlink-go/internal/model"
	"shortlink-go/internal/repository"
	"shortlink-go/pkg/logging"
	"shortlink-go/pkg/utils"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"shortlink-go/response"
)

// CreateShortLink 创建短链
func CreateShortLink(ctx context.Context, req dto.CreateShortLinkRequest) error {
	// Gin 标准验证
	if err := req.Validate(); err != nil {
		return apperrors.InvalidRequestError(err.Error())
	}

	// 检查短链是否已存在
	var existing model.ShortLink
	if err := repository.DB.Where("short_code = ?", req.ShortCode).First(&existing).Error; err == nil {
		logging.Logger.Info("短链已存在", zap.Error(err))
		return apperrors.BusinessError(http.StatusConflict, "短链已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logging.Logger.Info("查询短链失败", zap.Error(err))
		return apperrors.SystemErrorDefault()
	}

	// 构建模型
	shortLink := &model.ShortLink{
		TargetURL: req.TargetURL,
		ShortCode: req.ShortCode,
		Disabled:  req.Disabled, // 默认 false
	}

	// 数据库持久化
	if err := repository.DB.Create(shortLink).Error; err != nil {
		logging.Logger.Info("数据库操作失败", zap.Error(err))
		return apperrors.SystemErrorDefault()
	}
	return nil
}

// ListShortLinks 支持分页查询短链列表
func ListShortLinks(ctx context.Context, page, size int, shortCode string) (*response.PageResponse[model.ShortLink], error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10 // 默认每页10条，最大100条
	}

	// 构建查询条件
	db := repository.DB.Model(&model.ShortLink{})
	if shortCode != "" {
		db = db.Where("short_code LIKE ?", "%"+shortCode+"%")
	}

	// 查询总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, apperrors.SystemError("统计短链记录数失败: " + err.Error())
	}

	// 如果总数为0，直接返回空结果，不执行分页查询
	if total == 0 {
		return &response.PageResponse[model.ShortLink]{
			Page:      page,
			Size:      size,
			Total:     0,
			TotalPage: 0,
			List:      []model.ShortLink{},
		}, nil
	}

	// 查询分页数据
	var links []model.ShortLink
	if err := db.
		Limit(size).
		Offset((page - 1) * size).
		Order("id DESC").
		Find(&links).Error; err != nil {
		logging.Logger.Info("数据库操作失败", zap.Error(err))
		return nil, apperrors.SystemErrorDefault()
	}

	// 计算总页数
	totalPage := (int(total) + size - 1) / size

	return &response.PageResponse[model.ShortLink]{
		Page:      page,
		Size:      size,
		Total:     int(total),
		TotalPage: totalPage,
		List:      links,
	}, nil
}

// UpdateShortLinkStatus 更新短链状态（启用/禁用）
func UpdateShortLinkStatus(ctx context.Context, id uint, disabled bool) error {
	var link model.ShortLink
	if err := repository.DB.First(&link, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.BusinessError(http.StatusConflict, "短链不存在")
		}
		logging.Logger.Error("查询短链失败",
			zap.Uint("id", id),
			zap.Bool("disabled", disabled),
			zap.Error(err))
		return apperrors.SystemError("查询短链失败: " + err.Error())
	}

	// 更新状态
	link.Disabled = disabled
	if err := repository.DB.Save(&link).Error; err != nil {
		return apperrors.SystemError("更新短链状态失败: " + err.Error())
	}

	if disabled {

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

		// 禁用短链，删除 Redis 中的缓存
		cacheKey := constant.GetShortCodeKey(link.ShortCode)
		if _, err := conn.Do("DEL", cacheKey); err != nil {
			logging.Logger.Warn("Redis 删除缓存失败",
				zap.String("cache_key", cacheKey),
				zap.Error(err))
		}
	}

	return nil
}

// UpdateShortLink 仅更新短链的 target_url 字段
func UpdateShortLink(ctx context.Context, id uint, targetUrl string) error {
	// 查询现有短链记录
	var existing model.ShortLink
	if err := repository.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.BusinessError(http.StatusNotFound, "短链不存在")
		}
		logging.Logger.Warn("查询短链失败",
			zap.Uint("id", id),
			zap.String("target_url", targetUrl),
			zap.Error(err))
		return apperrors.SystemErrorDefault()
	}

	// 校验目标 URL（复用公共逻辑）
	if err := utils.ValidateTargetURL(targetUrl); err != nil {
		return apperrors.InvalidRequestError(err.Error())
	}

	if existing.TargetURL == targetUrl {
		return nil // 无需更新
	}

	// 更新 targetUrl 字段
	existing.TargetURL = targetUrl
	existing.UpdatedAt = time.Now() // 可选：更新时间戳

	// 保存更新
	if err := repository.DB.Save(&existing).Error; err != nil {
		logging.Logger.Warn("更新短链失败",
			zap.Uint("id", id),
			zap.String("target_url", targetUrl),
			zap.Error(err))
		return apperrors.SystemErrorDefault()
	}

	return nil
}

func RedirectToTargetURL(ctx context.Context, shortCode string, ip string) (*model.ShortLink, bool) {
	if err := utils.ValidateShortCode(shortCode); err != nil {
		logging.Logger.Error("无效的 short_code",
			zap.String("short_code", shortCode),         // 出错的 short_code
			zap.String("action", "validate_short_code"), // 当前操作
		)
		return nil, false
	}

	cacheKey := constant.GetShortCodeKey(shortCode)

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

	// 从 Redis 中查询缓存
	var cachedValue []byte
	var err error
	cachedValue, err = redis.Bytes(conn.Do("GET", cacheKey))
	if err == nil {
		var shortLink model.ShortLink
		if err := json.Unmarshal(cachedValue, &shortLink); err == nil {
			return &shortLink, true
		} else if string(cachedValue) == "" {
			return nil, false
		} else {
			logging.Logger.Warn("Failed to unmarshal cached value",
				zap.String("cache_key", cacheKey),
				zap.Error(err))
		}
	} else if err != redis.ErrNil {
		logging.Logger.Warn("Error getting from Redis",
			zap.String("cache_key", cacheKey),
			zap.Error(err))
	}

	// 缓存未命中，从数据库查询
	var shortLink model.ShortLink
	result := repository.DB.Where("short_code = ? AND disabled = ?", shortCode, false).First(&shortLink)
	if result.Error != nil {
		// 缓存空值，防止缓存穿透
		_, err := conn.Do("SET", cacheKey, "", "EX", 300)
		if err != nil {
			logging.Logger.Error("设置缓存失败",
				zap.String("cache_key", cacheKey),
				zap.Error(err),
			)
		}
		return nil, false
	}

	// 缓存结果（1小时）
	cachedValue, _ = json.Marshal(shortLink)

	_, err = conn.Do("SET", cacheKey, cachedValue, "EX", 3600)
	if err != nil {
		// 记录日志或者做其他错误处理
		logging.Logger.Error("设置缓存失败",
			zap.String("cache_key", cacheKey),
			zap.Any("value", cachedValue),
			zap.Error(err),
		)
	}

	RecordDailyPV(conn, shortCode)
	RecordDailyUV(conn, shortCode, ip)
	RecordTotalPV(conn, shortCode)
	RecordTotalUV(conn, shortCode, ip)

	return &shortLink, true

}

func StatisticalData() error {
	logging.Logger.Info("StatisticalData start")
	var links []model.ShortLink
	if err := repository.DB.Find(&links).Error; err != nil {
		logging.Logger.Error("获取短链列表失败", zap.Error(err))
		return err
	}
	// 获取当前日期，并格式化为 "2006-01-02" 格式的字符串
	today := time.Now().Format("2006-01-02")
	for _, link := range links {
		DoStatisticalData(link, today)
	}

	logging.Logger.Info("StatisticalData end")
	return nil
}

func DoStatisticalData(shortLink model.ShortLink, today string) {
	updatedAt := shortLink.UpdatedAt // time.Time 类型
	// 判断逻辑
	if shortLink.Disabled && !updatedAt.IsZero() { // updatedAt 不为零值（即有更新时间）
		yesterday := time.Now().AddDate(0, 0, -1)
		if updatedAt.Before(yesterday) {
			logging.Logger.Warn("#doStatisticalData | Skipping sync for shortcode",

				zap.String("shortcode", shortLink.ShortCode),
				zap.Bool("disabled", shortLink.Disabled),
				zap.Time("updatedTime", updatedAt),
			)
			return
		}
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

	dailyPv, _ := GetDailyPv(conn, shortLink.ShortCode, today)

	dailyUv, _ := GetDailyUv(conn, shortLink.ShortCode, today)

	totalPv, _ := GetTotalPv(conn, shortLink.ShortCode)

	totalUv, _ := GetTotalUv(conn, shortLink.ShortCode)

	// 更新数据库中的每日统计（DailyStat）
	dailyStat := &model.DailyStat{
		ShortLinkID: shortLink.ID,
		Date:        today,
		PV:          dailyPv,
		UV:          dailyUv,
	}

	// 获取数据库连接
	db := repository.DB.Where("short_link_id = ? AND date = ?", shortLink.ID, today).
		Assign("pv", dailyPv, "uv", dailyUv).
		FirstOrCreate(dailyStat)

	// 检查错误
	if db.Error != nil {
		logging.Logger.Error("Failed to insert or update daily stat",
			zap.Uint("short_link_id", shortLink.ID),
			zap.String("date", today),
			zap.Int64("pv", dailyPv),
			zap.Int64("uv", dailyUv),
			zap.Error(db.Error), // ✅ 正确：使用 db.Error
		)
	}

	// 更新数据库中的短链接总 PV/UV
	if err := repository.DB.Model(&shortLink).
		Where("id = ?", shortLink.ID).
		Updates(map[string]interface{}{
			"total_pv": totalPv,
			"total_uv": totalUv,
		}).Error; err != nil {
		logging.Logger.Error("Failed to update total PV/UV",
			zap.Uint("id", shortLink.ID),
			zap.Int64("total_pv", totalPv),
			zap.Int64("total_uv", totalUv),
			zap.Error(err))
	}

}

// GetStatsByShortLinkID 获取统计信息
func GetStatsByShortLinkID(id uint) ([]model.DailyStat, error) {
	var stats []model.DailyStat
	err := repository.DB.Where("short_link_id = ?", id).Order("date DESC").Find(&stats).Error
	return stats, err
}
