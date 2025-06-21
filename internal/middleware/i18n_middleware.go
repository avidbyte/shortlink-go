package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"

	thirdPartyI18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"shortlink-go/internal/i18n"
)

func I18nMiddleware(bundle *thirdPartyI18n.Bundle) gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptLanguage := c.GetHeader("Accept-Language")
		tags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
		lang := "en" // 默认语言
		for _, tag := range tags {
			if contains(i18n.SupportedLanguages, tag.String()) {
				lang = tag.String()
				break
			}
		}

		// ✅ 使用第三方库的 NewLocalizer
		localizer := thirdPartyI18n.NewLocalizer(bundle, lang)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "i18n.Localizer", localizer))
		c.Next()
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
