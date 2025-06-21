package i18n

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"strings"
)

// SupportedLanguages 是手动维护的支持语言列表
var SupportedLanguages []string

// InitI18n 初始化 i18n 包
func InitI18n(filePaths []string, defaultLang string) (*i18n.Bundle, error) {
	bundle := i18n.NewBundle(language.MustParse(defaultLang))
	// ⚠️ 注册 TOML 解析器
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	SupportedLanguages = make([]string, 0) // 清空旧列表

	for _, filePath := range filePaths {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		// 解析文件名中的语言标签（如 en.toml -> "en"）
		lang := extractLanguageFromPath(filePath)
		SupportedLanguages = append(SupportedLanguages, lang)

		_, err = bundle.ParseMessageFileBytes(file, filePath)
		if err != nil {
			return nil, err
		}
	}
	return bundle, nil
}

// 从文件路径中提取语言标签（假设文件名格式为 <lang>.toml）
func extractLanguageFromPath(filePath string) string {
	// 1. 获取文件名（不带路径）
	baseName := filepath.Base(filePath) // 例如: en.toml

	// 2. 去除文件扩展名
	langWithExt := strings.TrimSuffix(baseName, filepath.Ext(baseName)) // 例如: en

	return langWithExt
}

func T(ctx context.Context, key string, data map[string]interface{}) string {
	localizer := ctx.Value("i18n.Localizer").(*i18n.Localizer)
	config := &i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	}
	return localizer.MustLocalize(config)
}
