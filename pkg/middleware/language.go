package middleware

import (
	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru/v2"
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
	// 使用默认合理的缓存大小，避免过度参数化
	if maxSize <= 0 {
		maxSize = 1000 // 使用合理的默认值
	}

	// 初始化LRU缓存
	langOnce.Do(func() {
		var err error
		langCache, err = lru.New[string, string](maxSize)
		if err != nil {
			log.Error("■ ■ Language(中间件) ■ ■ 初始化缓存失败", log.FError(err))
		}
	})
	if langCache == nil {
		return func(context *gin.Context) {}
	}

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
	if header == "" {
		return []languagePreference{}
	}

	parts := strings.Split(header, ",")
	preferences := make([]languagePreference, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue // 跳过空项
		}

		weight := 1.0 // 默认权重
		langCode := part

		// 查找分号位置，避免不必要的字符串分割
		if qIndex := strings.IndexByte(part, ';'); qIndex >= 0 && qIndex < len(part)-1 {
			langCode = strings.TrimSpace(part[:qIndex])
			if langCode == "" {
				continue // 跳过空语言代码
			}

			// 只在有q=参数时解析权重
			qPart := strings.TrimSpace(part[qIndex+1:])
			if strings.HasPrefix(qPart, "q=") {
				qValue := strings.TrimSpace(strings.TrimPrefix(qPart, "q="))
				if qValue != "" {
					if parsedWeight, err := strconv.ParseFloat(qValue, 64); err == nil {
						// 确保权重在有效范围内
						if parsedWeight < 0 {
							weight = 0
						} else if parsedWeight > 1 {
							weight = 1
						} else {
							weight = parsedWeight
						}
					}
				}
			}
		}

		// 对语言代码进行额外的有效性验证
		if isValidLanguageCode(langCode) {
			preferences = append(preferences, languagePreference{lang: langCode, weight: weight})
		}
	}

	// 按权重排序语言偏好（从高到低）
	sort.Slice(preferences, func(i, j int) bool {
		return preferences[i].weight > preferences[j].weight
	})
	return preferences
}

// 辅助函数：检查语言代码是否有效
func isValidLanguageCode(code string) bool {
	// 简单验证：至少包含一个有效字符，且不含特殊字符
	if code == "" {
		return false
	}
	// 语言代码通常是2-3个字母，或带有区域代码的格式如zh-CN
	// 这里用简单验证，实际项目可能需要更复杂的验证
	for _, c := range code {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// findBestLanguageMatch 找到最佳匹配的语言
func findBestLanguageMatch(preferences []languagePreference) string {
	if len(preferences) == 0 {
		return i18n.DefLang()
	}

	for _, pref := range preferences {
		lang := pref.lang
		if lang == "" {
			continue
		}

		// 先检查完整语言代码
		if i18n.HasLang(lang) {
			return lang
		}

		// 仅当需要时分解语言标签
		if dashIndex := strings.IndexByte(lang, '-'); dashIndex > 0 && dashIndex < len(lang)-1 {
			// 主语言部分
			mainLang := lang[:dashIndex]
			if i18n.HasLang(mainLang) {
				return mainLang
			}

			// 区域部分（确保是有效的索引范围）
			if dashIndex+1 < len(lang) {
				region := lang[dashIndex+1:]
				if region != "" && i18n.HasLang(region) {
					return region
				}
			}
		}
	}

	// 如果没有找到匹配的语言，使用默认语言
	return i18n.DefLang()
}
