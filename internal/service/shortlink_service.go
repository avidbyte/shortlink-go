package service

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"shortlink-go/constant"
	"shortlink-go/internal/apperrors"
	"shortlink-go/internal/dto"
	"shortlink-go/internal/i18n"
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
		message := i18n.T(ctx, err.Error(), nil)
		return apperrors.InvalidRequestError(message)
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
		TargetURL:    req.TargetURL,
		ShortCode:    req.ShortCode,
		RedirectCode: req.RedirectCode,
		Disabled:     req.Disabled, // 默认 false
	}

	// 数据库持久化
	if err := repository.DB.Create(shortLink).Error; err != nil {
		logging.Logger.Info("数据库操作失败", zap.Error(err))
		return apperrors.SystemErrorDefault()
	}
	return nil
}

// ListShortLinks 支持分页查询短链列表
func ListShortLinks(
	ctx context.Context,
	page, size int,
	shortCode string,
	targetUrl string,
	redirectCode int,
	disabled *bool,
) (*response.PageResponse[model.ShortLink], error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// 构建查询条件
	db := repository.DB.Model(&model.ShortLink{})

	if shortCode != "" {
		db = db.Where("short_code LIKE ?", "%"+shortCode+"%")
	}
	if targetUrl != "" {
		db = db.Where("target_url LIKE ?", "%"+targetUrl+"%")
	}
	if redirectCode != 0 {
		db = db.Where("redirect_code = ?", redirectCode)
	}
	if disabled != nil {
		db = db.Where("disabled = ?", *disabled)
	}

	// 查询总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		logging.Logger.Error("统计短链记录数失败", zap.Error(err))
		message := i18n.T(ctx, "error.system_error", nil)
		return nil, apperrors.SystemError(message)
	}

	if total == 0 {
		return &response.PageResponse[model.ShortLink]{
			Page:      page,
			Size:      size,
			Total:     0,
			TotalPage: 0,
			List:      []model.ShortLink{},
		}, nil
	}

	// 分页查询
	var links []model.ShortLink
	if err := db.
		Limit(size).
		Offset((page - 1) * size).
		Order("id DESC").
		Find(&links).Error; err != nil {
		logging.Logger.Error("分页查询短链失败", zap.Error(err))
		message := i18n.T(ctx, "error.system_error", nil)
		return nil, apperrors.SystemError(message)
	}

	totalPage := (int(total) + size - 1) / size

	return &response.PageResponse[model.ShortLink]{
		Page:      page,
		Size:      size,
		Total:     int(total),
		TotalPage: totalPage,
		List:      links,
	}, nil
}

// UpdateShortLink 更新短链配置（包含状态可选修改）
func UpdateShortLink(ctx context.Context, id uint, targetUrl string, redirectCode int, newDisabled *bool) error {

	// 校验目标 URL
	if err := utils.ValidateTargetURL(targetUrl); err != nil {
		message := i18n.T(ctx, err.Error(), nil)
		return apperrors.InvalidRequestError(message)
	}

	// 查询现有短链记录
	var existing model.ShortLink
	if err := repository.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logging.Logger.Info("短链不存在",
				zap.Uint("id", id),
				zap.String("target_url", targetUrl))
			message := i18n.T(ctx, "error.shortcode_not_found", nil)
			return apperrors.BusinessError(http.StatusNotFound, message)
		}
		logging.Logger.Error("查询短链失败",
			zap.Uint("id", id),
			zap.String("target_url", targetUrl),
			zap.Error(err))
		message := i18n.T(ctx, "error.system_error", nil)
		return apperrors.SystemError(message)
	}

	// 判断状态是否需要变更
	if newDisabled != nil && *newDisabled != existing.Disabled {
		if *newDisabled {
			// 禁用：同步统计、备份、清理 Redis
			if err := DoStatisticalData(&existing, constant.GetDateKey()); err != nil {
				logging.Logger.Error("禁用时同步统计数据失败",
					zap.Uint("id", existing.ID),
					zap.Error(err))
				return apperrors.SystemError(i18n.T(ctx, "error.system_error", nil))
			}

			if err := HandleShortLinkRedisHllBackup(&existing); err != nil {
				logging.Logger.Error("禁用时清理 Redis 缓存失败",
					zap.Uint("id", existing.ID),
					zap.Error(err))
				return apperrors.SystemError(i18n.T(ctx, "error.system_error", nil))
			}

			if err := HandleShortLinkRedisCleanup(&existing); err != nil {
				logging.Logger.Error("禁用时清理 Redis 缓存失败",
					zap.Uint("id", existing.ID),
					zap.Error(err))
				return apperrors.SystemError(i18n.T(ctx, "error.system_error", nil))
			}
		} else {
			// 由禁用变为启用，恢复 Redis 缓存
			if err := RestoreShortLinkCacheFromDB(&existing); err != nil {
				logging.Logger.Error("启用时恢复 Redis 缓存失败",
					zap.Uint("id", existing.ID),
					zap.Error(err))
				return apperrors.SystemError(i18n.T(ctx, "error.redis_restore_failed", nil))
			}
		}
		// 更新状态字段
		existing.Disabled = *newDisabled
	}

	// 更新 targetUrl（如果有变更）
	if existing.TargetURL != targetUrl {
		existing.TargetURL = targetUrl
	}

	if existing.RedirectCode != redirectCode {
		existing.RedirectCode = redirectCode
	}

	existing.UpdatedAt = time.Now()

	// 保存更新
	if err := repository.DB.Save(&existing).Error; err != nil {
		logging.Logger.Error("更新短链失败",
			zap.Uint("id", id),
			zap.String("target_url", targetUrl),
			zap.Bool("disabled", existing.Disabled),
			zap.Error(err))
		message := i18n.T(ctx, "error.system_error", nil)
		return apperrors.SystemError(message)
	}

	return nil
}

func RedirectToTargetURL(shortCode string, ip string) (*model.ShortLink, bool) {
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
	logging.Logger.Info("#StatisticalData | start")
	var shortLinks []model.ShortLink
	if err := repository.DB.Find(&shortLinks).Error; err != nil {
		logging.Logger.Error("获取短链列表失败", zap.Error(err))
		return err
	}

	for _, shortLink := range shortLinks {
		// 禁用：同步统计、备份、清理 Redis
		if err := DoStatisticalData(&shortLink, constant.GetDateKey()); err != nil {
			logging.Logger.Error("禁用时同步统计数据失败",
				zap.Uint("id", shortLink.ID),
				zap.Error(err))
		}

	}

	logging.Logger.Info("#StatisticalData | end")
	return nil
}

func DeleteShortLink(ctx context.Context, id uint) error {
	return repository.DB.Transaction(func(tx *gorm.DB) error {
		// 查询现有短链记录
		var existing model.ShortLink
		if err := tx.First(&existing, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			logging.Logger.Error("查询短链失败",
				zap.Uint("id", id),
				zap.Error(err))
			message := i18n.T(ctx, "error.system_error", nil)
			return apperrors.SystemError(message)
		}

		// 删除 daily_stats 统计记录
		if err := tx.Where("short_link_id = ?", existing.ID).Delete(&model.DailyStat{}).Error; err != nil {
			logging.Logger.Error("删除 daily_stats 统计记录失败",
				zap.Uint("id", existing.ID),
				zap.Error(err))
			return apperrors.SystemError(i18n.T(ctx, "error.daily_stats_delete_failed", nil))
		}

		// 删除 short_link 本身
		if err := tx.Delete(&model.ShortLink{}, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				message := i18n.T(ctx, "error.shortcode_not_found", nil)
				return apperrors.BusinessError(http.StatusNotFound, message)
			}
			logging.Logger.Error("删除短链失败",
				zap.Uint("id", id),
				zap.Error(err))
			message := i18n.T(ctx, "error.system_error", nil)
			return apperrors.SystemError(message)
		}

		// 清理 Redis 缓存
		if err := HandleShortLinkRedisCleanup(&existing); err != nil {
			logging.Logger.Error("禁用时清理 Redis 缓存失败",
				zap.Uint("id", existing.ID),
				zap.String("shortcode", existing.ShortCode),
				zap.Error(err))
			return apperrors.SystemError(i18n.T(ctx, "error.redis_cleanup_failed", nil))
		}

		return nil
	})
}

func DoStatisticalData(shortLink *model.ShortLink, today string) error {
	updatedAt := shortLink.UpdatedAt
	if shortLink.Disabled && !updatedAt.IsZero() {
		yesterday := time.Now().AddDate(0, 0, -1)
		if updatedAt.Before(yesterday) {
			logging.Logger.Warn("Skipping sync for disabled shortcode",
				zap.String("shortcode", shortLink.ShortCode))
			return nil
		}
	}

	dailyPv, dailyUv, totalPv, totalUv, err := GetStatisticalData(*shortLink, today)
	if err != nil {
		return err
	}

	shortLink.TotalPV = totalPv
	shortLink.TotalUV = totalUv

	return SaveStatisticalData(shortLink, today, dailyPv, dailyUv, totalPv, totalUv)
}

func GetStatisticalData(shortLink model.ShortLink, today string) (dailyPv, dailyUv, totalPv, totalUv uint64, err error) {
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

	dailyPv, err = GetDailyPv(conn, shortLink.ShortCode, today)
	if err != nil {
		return
	}

	dailyUv, err = GetDailyUv(conn, shortLink.ShortCode, today)
	if err != nil {
		return
	}

	totalPv, err = GetTotalPv(conn, shortLink.ShortCode)
	if err != nil {
		return
	}

	totalUv, err = GetTotalUv(conn, shortLink.ShortCode)
	if err != nil {
		return
	}

	return
}

// SaveStatisticalData 保存统计数据
func SaveStatisticalData(shortLink *model.ShortLink, today string, dailyPv, dailyUv, totalPv, totalUv uint64) error {

	if err := repository.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "short_link_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"pv", "uv"}),
	}).Create(&model.DailyStat{
		ShortLinkID: shortLink.ID,
		Date:        today,
		PV:          dailyPv,
		UV:          dailyUv,
	}).Error; err != nil {
		return err
	}

	if err := repository.DB.Model(&shortLink).
		Where("id = ?", shortLink.ID).
		Updates(map[string]interface{}{
			"total_pv": totalPv,
			"total_uv": totalUv,
		}).Error; err != nil {
		logging.Logger.Error("Failed to update total PV/UV", zap.Error(err))
		return err
	}

	return nil
}

func HandleShortLinkRedisHllBackup(shortLink *model.ShortLink) error {
	conn := repository.RedisPool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logging.Logger.Error("关闭 Redis 连接失败",
				zap.Error(err),
				zap.String("operation", "close"),
				zap.String("connection_type", "redis"),
			)
		}
	}()

	shortcode := shortLink.ShortCode
	totalUvKey := constant.GetTotalUVKey(shortcode)

	hllData, err := redis.Bytes(conn.Do("DUMP", totalUvKey))
	if err != nil {
		if err == redis.ErrNil {
			logging.Logger.Info("HyperLogLog 不存在，无需备份", zap.String("key", totalUvKey))
			hllData = nil
		} else {
			logging.Logger.Warn("备份 HyperLogLog 失败", zap.Error(err))
			return err
		}
	}
	shortLink.UvHLLBackup = hllData

	if err := repository.DB.Save(shortLink).Error; err != nil {
		logging.Logger.Error("保存 UV HyperLogLog 备份失败", zap.Error(err))
		return err
	}

	return nil
}

func HandleShortLinkRedisCleanup(shortLink *model.ShortLink) error {
	conn := repository.RedisPool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logging.Logger.Error("关闭 Redis 连接失败",
				zap.Error(err),
				zap.String("operation", "close"),
				zap.String("connection_type", "redis"),
			)
		}
	}()

	shortcode := shortLink.ShortCode
	totalUvKey := constant.GetTotalUVKey(shortcode)
	cacheKey := constant.GetShortCodeKey(shortcode)
	totalPvKey := constant.GetTotalPVKey(shortcode)

	for _, key := range []string{cacheKey, totalPvKey, totalUvKey} {
		if _, err := conn.Do("DEL", key); err != nil {
			logging.Logger.Warn("删除 Redis 缓存失败", zap.String("key", key), zap.Error(err))
			// return err
		}
	}

	return nil
}

// RestoreShortLinkCacheFromDB 从数据库恢复 Redis 缓存（PV + UV）
func RestoreShortLinkCacheFromDB(shortLink *model.ShortLink) error {
	conn := repository.RedisPool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logging.Logger.Error("关闭 Redis 连接失败",
				zap.Error(err),
				zap.String("operation", "close"),
				zap.String("connection_type", "redis"),
			)
		}
	}()

	shortcode := shortLink.ShortCode
	totalPvKey := constant.GetTotalPVKey(shortcode)
	totalUvKey := constant.GetTotalUVKey(shortcode)

	// 恢复 PV
	if shortLink.TotalPV > 0 {
		if _, err := conn.Do("SET", totalPvKey, shortLink.TotalPV); err != nil {
			logging.Logger.Warn("恢复 Redis 总 PV 失败",
				zap.String("key", totalPvKey),
				zap.Uint64("value", shortLink.TotalPV),
				zap.Error(err))
		}
	}

	// 恢复 UV HyperLogLog
	if len(shortLink.UvHLLBackup) > 0 {
		_, _ = conn.Do("DEL", totalUvKey)
		if _, err := conn.Do("RESTORE", totalUvKey, 0, shortLink.UvHLLBackup); err != nil {
			logging.Logger.Warn("恢复 Redis HyperLogLog 失败",
				zap.String("key", totalUvKey),
				zap.Error(err))
		}
	}
	return nil
}
