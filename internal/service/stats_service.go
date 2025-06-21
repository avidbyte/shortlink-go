package service

import (
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"shortlink-go/constant"
	"shortlink-go/pkg/logging"
)

// RecordDailyPV 记录每日 PV
func RecordDailyPV(conn redis.Conn, shortCode string) {
	dailyPvKey := constant.GetDailyPVKey(constant.GetDateKey())

	_, err := conn.Do("HINCRBY", dailyPvKey, shortCode, 1)
	if err != nil {
		logging.Logger.Error("Failed to record daily PV",
			zap.String("key", dailyPvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
	}

	_, err = conn.Do("EXPIRE", dailyPvKey, 3*24*3600) // 3天过期
	if err != nil {
		logging.Logger.Error("Failed to record daily PV Expire",
			zap.String("key", dailyPvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
	}
}

// RecordDailyUV 记录每日 UV
func RecordDailyUV(conn redis.Conn, shortCode string, ip string) {

	dailyUvKey := constant.GetDailyUVKey(shortCode, constant.GetDateKey())

	_, err := conn.Do("PFADD", dailyUvKey, ip)
	if err != nil {
		logging.Logger.Error("Failed to record daily UV",
			zap.String("key", dailyUvKey),
			zap.String("ip", ip),
			zap.Error(err))
	}

	_, err = conn.Do("EXPIRE", dailyUvKey, 3*24*3600) // 3天过期
	if err != nil {
		logging.Logger.Error("Failed to record daily UV Expire",
			zap.String("key", dailyUvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
	}
}

// RecordTotalPV 记录总 PV
func RecordTotalPV(conn redis.Conn, shortCode string) {
	totalPvKey := constant.GetTotalPVKey(shortCode)
	_, err := conn.Do("INCR", totalPvKey)
	if err != nil {
		logging.Logger.Error("Failed to record total PV",
			zap.String("key", totalPvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
	}
}

// RecordTotalUV 记录总UV
func RecordTotalUV(conn redis.Conn, shortCode string, ip string) {
	totalUvKey := constant.GetTotalUVKey(shortCode)
	_, err := conn.Do("PFADD", totalUvKey, ip)
	if err != nil {
		logging.Logger.Error("Failed to record total UV",
			zap.String("key", totalUvKey),
			zap.String("ip", ip),
			zap.Error(err))
	}
}

// GetDailyPv 获取某日期的短链接访问量（PV）
func GetDailyPv(conn redis.Conn, shortCode string, date string) (int64, error) {
	dailyPvKey := constant.GetDailyPVKey(date)

	// 使用 HGET 查询指定 shortCode 的访问量
	reply, err := conn.Do("HGET", dailyPvKey, shortCode)
	if err != nil {
		logging.Logger.Error("Failed to get daily PV",
			zap.String("key", dailyPvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
		return 0, err
	}

	// 将 Redis 回复转换为 int64
	result, err := redis.Int64(reply, err)
	if err != nil {
		logging.Logger.Error("Failed to parse daily PV",
			zap.String("key", dailyPvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
		return 0, err
	}

	return result, nil
}

// GetDailyUv 获取某日期的短链接独立访客数（UV）
func GetDailyUv(conn redis.Conn, shortCode string, date string) (int64, error) {
	dailyUvKey := constant.GetDailyUVKey(shortCode, date)

	// 使用 PFCount 查询 HyperLogLog 的基数（UV 数量）
	reply, err := conn.Do("PFCount", dailyUvKey)
	if err != nil {
		logging.Logger.Error("Failed to get daily UV",
			zap.String("key", dailyUvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
		return 0, err
	}

	// 将 Redis 回复转换为 int64
	result, err := redis.Int64(reply, err)
	if err != nil {
		logging.Logger.Error("Failed to parse daily UV",
			zap.String("key", dailyUvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
		return 0, err
	}

	return result, nil
}

// GetTotalPv 获取短链接的总访问量（PV）
func GetTotalPv(conn redis.Conn, shortCode string) (int64, error) {
	totalPvKey := constant.GetTotalPVKey(shortCode)

	// 使用 GET 查询字符串类型的计数器
	reply, err := conn.Do("GET", totalPvKey)
	if err != nil {
		logging.Logger.Error("Failed to get total PV",
			zap.String("key", totalPvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
		return 0, err
	}

	// 将 Redis 回复转换为 int64
	result, err := redis.Int64(reply, err)
	if err != nil {
		logging.Logger.Error("Failed to parse total PV",
			zap.String("key", totalPvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
		return 0, err
	}

	return result, nil
}

// GetTotalUv 获取短链接的总独立访客数（UV）
func GetTotalUv(conn redis.Conn, shortCode string) (int64, error) {
	totalUvKey := constant.GetTotalUVKey(shortCode)

	// 使用 PFCount 查询 HyperLogLog 的基数（UV 数量）
	reply, err := conn.Do("PFCount", totalUvKey)
	if err != nil {
		logging.Logger.Error("Failed to get total UV",
			zap.String("key", totalUvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
		return 0, err
	}

	// 将 Redis 回复转换为 int64
	result, err := redis.Int64(reply, err)
	if err != nil {
		logging.Logger.Error("Failed to parse total UV",
			zap.String("key", totalUvKey),
			zap.String("short_code", shortCode),
			zap.Error(err))
		return 0, err
	}

	return result, nil
}
