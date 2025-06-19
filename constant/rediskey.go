package constant

import (
	"fmt"
	"time"
)

// 常量定义
const (
	BasePrefix = "redirect:"
	Separator  = ":"
)

// Redis 键模板
const (
	ShortCode = BasePrefix + "shortcode:%s"
	DailyPV   = BasePrefix + "pv" + Separator + "%s"                    // redirect:pv:yyyyMMdd
	DailyUV   = BasePrefix + "uv" + Separator + "%s" + Separator + "%s" // redirect:uv:yyyyMMdd:shortcode
	TotalPV   = BasePrefix + "total_pv" + Separator + "%s"              // redirect:total_pv:shortcode
	TotalUV   = BasePrefix + "total_uv" + Separator + "%s"              // redirect:total_uv:shortcode
)

// GetShortCodeKey 生成 shortCode key
func GetShortCodeKey(shortcode string) string {
	return fmt.Sprintf(ShortCode, shortcode)
}

// GetDateKey  生成当前日期的键（格式：yyyyMMdd）
func GetDateKey() string {
	return time.Now().Format("20060102") // Go 中日期格式规则：2006-01-02
}

// GetDailyPVKey 生成每日 PV 键（格式：redirect:pv:yyyyMMdd）
func GetDailyPVKey(date string) string {
	return fmt.Sprintf(DailyPV, date)
}

// GetDailyUVKey 生成每日 UV 键（格式：redirect:uv:yyyyMMdd:shortcode）
func GetDailyUVKey(shortcode, date string) string {
	return fmt.Sprintf(DailyUV, date, shortcode)
}

// GetTotalUVKey 生成总 UV 键（格式：redirect:total_uv:shortcode）
func GetTotalUVKey(shortcode string) string {
	return fmt.Sprintf(TotalUV, shortcode)
}

// GetTotalPVKey 生成总 PV 键（格式：redirect:total_pv:shortcode）
func GetTotalPVKey(shortcode string) string {
	return fmt.Sprintf(TotalPV, shortcode)
}
