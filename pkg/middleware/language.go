package middleware

import (
	"github.com/gin-gonic/gin"
	"katydid-mp-user/pkg/i18n"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// LanguageKey 是上下文中存储语言信息的键
const LanguageKey = "Use-Language"

// 使用缓存减少重复计算
var (
	langCache     = make(map[string]string)
	langCacheLock sync.RWMutex
)

// languagePreference 表示带权重的语言偏好
type languagePreference struct {
	lang   string  // 语言代码
	weight float64 // 权重值
}

// Language 创建处理请求语言的中间件
func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			// 如果头部为空，直接使用默认语言
			c.Set(LanguageKey, i18n.DefLang())
			c.Next()
			return
		}

		// 优先检查缓存
		langCacheLock.RLock()
		if cachedLang, exists := langCache[lang]; exists {
			langCacheLock.RUnlock()
			c.Set(LanguageKey, cachedLang)
			c.Next()
			return
		}
		langCacheLock.RUnlock()

		// 解析所有语言偏好及其权重
		preferences := parseLanguagePreferences(lang)

		// 按权重排序语言偏好（从高到低）
		sort.Slice(preferences, func(i, j int) bool {
			return preferences[i].weight > preferences[j].weight
		})

		var resultLang string

		// 按权重顺序检查每个语言偏好
		for _, pref := range preferences {
			cleanLang := pref.lang
			// 检查完整语言代码
			if i18n.HasLang(cleanLang) {
				resultLang = cleanLang
				break
			}

			// 分解语言标签（如 zh-CN -> zh, CN）
			subParts := strings.Split(cleanLang, "-")
			if len(subParts) > 1 {
				// 检查主语言部分
				if i18n.HasLang(subParts[0]) {
					resultLang = subParts[0]
					break
				}

				// 检查区域部分
				if i18n.HasLang(subParts[1]) {
					resultLang = subParts[1]
					break
				}
			}
		}

		// 如果没有找到匹配的语言，使用默认语言
		if resultLang == "" {
			resultLang = i18n.DefLang()
		}

		// 更新缓存
		langCacheLock.Lock()
		if len(langCache) > 1000 { // 避免缓存无限增长
			langCache = make(map[string]string)
		}
		langCache[lang] = resultLang
		langCacheLock.Unlock()

		c.Set(LanguageKey, resultLang)
		c.Next()
	}
}

// parseLanguagePreferences 解析Accept-Language头部，返回按权重排序的语言偏好列表
func parseLanguagePreferences(header string) []languagePreference {
	parts := strings.Split(header, ",")
	preferences := make([]languagePreference, 0, len(parts))

	for _, part := range parts {
		pref := languagePreference{weight: 1.0} // 默认权重为1.0

		// 分离语言代码和权重值
		subParts := strings.Split(part, ";")
		pref.lang = strings.TrimSpace(subParts[0])

		// 如果有权重值，解析它
		if len(subParts) > 1 {
			qParts := strings.Split(subParts[1], "=")
			if len(qParts) == 2 && strings.TrimSpace(qParts[0]) == "q" {
				if weight, err := strconv.ParseFloat(strings.TrimSpace(qParts[1]), 64); err == nil {
					pref.weight = weight
				}
			}
		}
		preferences = append(preferences, pref)
	}
	return preferences
}
