package middleware

import (
	"github.com/gin-gonic/gin"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	HeaderKeyAcceptLanguage = "Accept-Language"
)

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
		lang := c.GetHeader(HeaderKeyAcceptLanguage)
		if lang == "" {
			log.DebugFmt("■ ■ Language ■ ■ 空就默认语言: %s", i18n.DefLang())
			// 如果头部为空，直接使用默认语言
			c.Set(LanguageKey, i18n.DefLang())
			c.Next()
			return
		}

		// 优先检查缓存
		resultLang := getCachedLanguage(lang)
		if resultLang != "" {
			log.DebugFmt("■ ■ Language ■ ■ 缓存命中: %s -> %s", lang, resultLang)
			c.Set(LanguageKey, resultLang)
			c.Next()
			return
		}
		log.DebugFmt("■ ■ Language ■ ■ 缓存未命中: %s", lang)

		// 解析所有语言偏好并按权重排序（从高到低）
		preferences := parseLanguagePreferences(lang)

		// 查找最佳匹配语言
		resultLang = findBestLanguageMatch(preferences)

		// 更新缓存
		updateLanguageCache(lang, resultLang)
		log.DebugFmt("■ ■ Language ■ ■ 更新缓存: %s -> %s", lang, resultLang)

		c.Set(LanguageKey, resultLang)
		c.Next()
	}
}

// getCachedLanguage 从缓存中获取语言
func getCachedLanguage(lang string) string {
	langCacheLock.RLock()
	defer langCacheLock.RUnlock()

	if cachedLang, exists := langCache[lang]; exists {
		return cachedLang
	}
	return ""
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

	// 按权重排序语言偏好（从高到低）
	sort.Slice(preferences, func(i, j int) bool {
		return preferences[i].weight > preferences[j].weight
	})
	return preferences
}

// findBestLanguageMatch 找到最佳匹配的语言
func findBestLanguageMatch(preferences []languagePreference) string {
	for _, pref := range preferences {
		cleanLang := pref.lang
		// 检查完整语言代码
		if i18n.HasLang(cleanLang) {
			return cleanLang
		}

		// 分解语言标签（如 zh-CN -> zh, CN）
		subParts := strings.Split(cleanLang, "-")
		if len(subParts) > 1 {
			// 先检查主语言部分
			if i18n.HasLang(subParts[0]) {
				return subParts[0]
			}
			// 再检查区域部分
			if i18n.HasLang(subParts[1]) {
				return subParts[1]
			}
		}
	}

	// 如果没有找到匹配的语言，使用默认语言
	return i18n.DefLang()
}

// updateLanguageCache 更新语言缓存
func updateLanguageCache(key, lang string) {
	langCacheLock.Lock()
	defer langCacheLock.Unlock()

	// 缓存最大条目数，避免缓存无限增长
	//if len(langCache) >= 10000 {
	//	langCache = make(map[string]string)
	//}
	langCache[key] = lang
}
