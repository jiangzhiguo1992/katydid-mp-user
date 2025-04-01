package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	// 使用LRU缓存减少重复计算
	langCache *lru.Cache[string, string]
	langOnce  sync.Once
)

// languagePreference 表示带权重的语言偏好
type languagePreference struct {
	lang   string  // 语言代码
	weight float64 // 权重值
}

// Language 创建处理请求语言的中间件
func Language(keyAccept string, maxSize int) gin.HandlerFunc {
	// 初始化LRU缓存
	langOnce.Do(func() {
		var err error
		langCache, err = lru.New[string, string](maxSize)
		if err != nil {
			log.Error("■ ■ Language(中间件) ■ ■ 初始化缓存失败", log.FError(err))
			// 创建一个容量较小的缓存作为后备
			langCache, _ = lru.New[string, string](maxSize / 10)
		}
	})

	return func(c *gin.Context) {
		acceptLang := c.GetHeader("Accept-Language")
		if acceptLang == "" {
			// 如果头部为空，直接使用默认语言
			c.Set(keyAccept, i18n.DefLang())
			c.Next()
			return
		}

		// 优先检查缓存
		if resultLang, found := langCache.Get(acceptLang); found {
			c.Set(keyAccept, resultLang)
			c.Next()
			return
		}

		// 解析所有语言偏好并按权重排序（从高到低）
		preferences := parseLanguagePreferences(acceptLang)

		// 查找最佳匹配语言
		resultLang := findBestLanguageMatch(preferences)

		// 更新缓存
		langCache.Add(acceptLang, resultLang)
		log.Debugf("■ ■ Language ■ ■(中间件) 更新缓存: %s -> %s", acceptLang, resultLang)

		c.Set(keyAccept, resultLang)
		c.Next()
	}
}

// parseLanguagePreferences 解析Accept-Language头部，返回按权重排序的语言偏好列表
func parseLanguagePreferences(header string) []languagePreference {
	parts := strings.Split(header, ",")
	preferences := make([]languagePreference, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue // 跳过空项
		}

		pref := languagePreference{weight: 1.0} // 默认权重为1.0

		// 分离语言代码和权重值
		subParts := strings.Split(part, ";")
		pref.lang = strings.TrimSpace(subParts[0])
		if pref.lang == "" {
			continue // 跳过无效语言代码
		}

		// 如果有权重值，解析它
		if len(subParts) > 1 {
			qParts := strings.Split(subParts[1], "=")
			if len(qParts) == 2 && strings.TrimSpace(qParts[0]) == "q" {
				if weight, err := strconv.ParseFloat(strings.TrimSpace(qParts[1]), 64); err == nil && weight >= 0 && weight <= 1 {
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
	if len(preferences) == 0 {
		return i18n.DefLang() // 如果没有偏好，返回默认语言
	}

	for _, pref := range preferences {
		cleanLang := pref.lang
		if cleanLang == "" {
			continue
		}

		// 检查完整语言代码
		if i18n.HasLang(cleanLang) {
			return cleanLang
		}

		// 分解语言标签（如 zh-CN -> zh, CN）
		subParts := strings.Split(cleanLang, "-")
		if len(subParts) > 1 {
			// 先检查主语言部分
			mainLang := strings.TrimSpace(subParts[0])
			if mainLang != "" && i18n.HasLang(mainLang) {
				return mainLang
			}

			// 再检查区域部分
			region := strings.TrimSpace(subParts[1])
			if region != "" && i18n.HasLang(region) {
				return region
			}
		}
	}

	// 如果没有找到匹配的语言，使用默认语言
	return i18n.DefLang()
}
