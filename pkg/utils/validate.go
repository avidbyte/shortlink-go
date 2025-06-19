package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"unicode"
)

// ValidateShortCode 校验 ShortCode 是否合法
func ValidateShortCode(shortCode string) error {
	if shortCode == "" {
		return fmt.Errorf("short_code 不能为空")
	}

	if ContainsWhitespace(shortCode) {
		return fmt.Errorf("short_code 不能包含空格")
	}

	shortCodePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)*$`)
	if !shortCodePattern.MatchString(shortCode) {
		return fmt.Errorf("short_code 格式不合法")
	}

	return nil
}

// ValidateTargetURL 校验目标 URL 的合法性
func ValidateTargetURL(targetURL string) error {
	// 1. 检查目标 URL 是否为空
	if targetURL == "" {
		return fmt.Errorf("target_url 不能为空")
	}

	// 2. URL 格式校验
	if _, err := url.ParseRequestURI(targetURL); err != nil {
		return fmt.Errorf("target_url 格式不合法")
	}

	// 3. URL 长度限制
	if len(targetURL) > 2048 {
		return fmt.Errorf("target_url 超过最大长度 2048 字符")
	}
	return nil
}

func ContainsWhitespace(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}
