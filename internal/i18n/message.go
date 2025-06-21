package i18n

//
//import (
//	"context"
//	"github.com/nicksnyder/go-i18n/v2/i18n"
//	"shortlink-go/constant"
//)
//
//type MessageSource interface {
//	GetMessage(ctx context.Context, messageID string) (string, error)
//}
//
//type DefaultMessageSource struct{}
//
//func (m *DefaultMessageSource) GetMessage(ctx context.Context, messageID string) (string, error) {
//	// 1. 优先从 Context 中获取已缓存的 Localizer
//	if localizer, ok := ctx.Value("i18n.Localizer").(*i18n.Localizer); ok {
//		config := &i18n.LocalizeConfig{
//			MessageID: messageID,
//		}
//		return localizer.Localize(config)
//	}
//
//	// 2. 如果没有缓存，则回退到从语言标签创建 Localizer
//	lang, ok := ctx.Value(constant.LanguageContextKey).(string)
//	if !ok {
//		lang = "en"
//	}
//	localizer := i18n.NewLocalizer(Bundle, lang)
//	config := &i18n.LocalizeConfig{
//		MessageID: messageID,
//	}
//	return localizer.Localize(config)
//}
//
//
