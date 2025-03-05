package middleware

import (
	"github.com/gin-gonic/gin"
	"katydid-mp-user/pkg/i18n"
	"strings"
)

const LanguageKey = "Use-Language"

func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")

		// 检查完整语言代码是否支持
		if i18n.HasLang(lang) {
			c.Set(LanguageKey, lang)
		} else if parts := strings.Split(lang, "-"); len(parts) > 1 {
			// 检查语言基础部分是否支持
			if i18n.HasLang(parts[0]) {
				c.Set(LanguageKey, parts[0])
			} else {
				// 使用默认语言
				c.Set(LanguageKey, i18n.DefLang())
			}
		} else {
			// 使用默认语言
			c.Set(LanguageKey, i18n.DefLang())
		}

		c.Next()
	}
}
