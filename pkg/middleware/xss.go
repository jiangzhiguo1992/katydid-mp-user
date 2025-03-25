package middleware

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/patrickmn/go-cache"
)

var (
	// 扩展XSS攻击模式检测
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script.*?>.*?</script.*?>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`), // 匹配所有事件处理程序
		regexp.MustCompile(`(?i)data:text/html`),
		regexp.MustCompile(`(?i)<iframe.*?>`),
		regexp.MustCompile(`(?i)document\.cookie`),
		regexp.MustCompile(`(?i)document\.domain`),
		regexp.MustCompile(`(?i)eval\(`),
		regexp.MustCompile(`(?i)setTimeout\(`),
		regexp.MustCompile(`(?i)setInterval\(`),
		regexp.MustCompile(`(?i)new\s+Function\(`),
	}
)

// XSSConfig XSS防护中间件配置
type XSSConfig struct {
	// 是否过滤请求参数
	FilterParams bool
	// 是否过滤请求体
	FilterBody bool
	// 是否过滤响应
	FilterResponse bool
	// 是否启用缓存（提高性能）
	EnableCache bool
	// 缓存过期时间
	CacheExpiration time.Duration
	// 缓存清理时间间隔
	CacheCleanupInterval time.Duration
	// 最大缓存项数量 (go-cache不直接支持项数量限制，通过过期时间间接控制)
	MaxCacheItems int
	// 检查严格模式
	StrictMode bool
}

// DefaultXSSConfig 默认XSS配置
func DefaultXSSConfig() XSSConfig {
	return XSSConfig{
		FilterParams:         true,
		FilterBody:           true,
		FilterResponse:       true,
		EnableCache:          true,
		CacheExpiration:      5 * time.Minute,
		CacheCleanupInterval: 10 * time.Minute,
		MaxCacheItems:        1000, // 通过过期时间间接控制
		StrictMode:           true,
	}
}

// XSS 跨站脚本攻击
func XSS(config ...XSSConfig) gin.HandlerFunc {
	// 使用默认配置或者用户提供的配置
	var cfg XSSConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultXSSConfig()
	}

	// 创建HTML过滤策略
	policy := createPolicy(cfg.StrictMode)

	// 如果启用缓存，初始化go-cache
	var xssCache *cache.Cache
	if cfg.EnableCache {
		xssCache = cache.New(cfg.CacheExpiration, cfg.CacheCleanupInterval)
	}

	return func(c *gin.Context) {
		// 设置安全响应头
		setSecurityHeaders(c)

		// 检查并过滤URL参数
		if cfg.FilterParams {
			filterURLParams(c, policy, xssCache)
		}

		// 检查并过滤请求体
		if cfg.FilterBody {
			if err := filterRequestBody(c, policy, xssCache); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "无法处理请求内容",
				})
				return
			}
		}

		// 如果需要过滤响应，包装响应写入器
		if cfg.FilterResponse {
			c.Writer = &xssResponseWriter{ResponseWriter: c.Writer, policy: policy, cache: xssCache}
		}

		c.Next()
	}
}

// 创建HTML过滤策略
func createPolicy(strictMode bool) *bluemonday.Policy {
	// 不论是否严格模式，都使用严格策略，然后根据需要添加灵活性
	policy := bluemonday.StrictPolicy()

	if !strictMode {
		// 非严格模式下允许一些基本格式
		policy.AllowStandardAttributes()
		policy.AllowStandardURLs()
		policy.AllowElements("p", "br", "b", "i", "strong", "em")
	}
	return policy
}

// 设置安全响应头
func setSecurityHeaders(c *gin.Context) {
	// 启用XSS过滤器，如果检测到XSS攻击，浏览器将阻止页面渲染
	c.Header("X-XSS-Protection", "1; mode=block")

	// 设置内容安全策略(CSP)，限制资源的加载来源:
	// - default-src 'self': 默认只允许从同源加载所有类型资源
	// - script-src 'self': 仅允许从同源加载脚本
	// - object-src 'none': 禁止所有插件资源(如Flash)
	// - base-uri 'self': 限制<base>标签的URL只能是同源
	// - frame-ancestors 'none': 禁止任何网站在框架中嵌入此页面(防止点击劫持)
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")

	// 防止浏览器对响应内容进行MIME类型嗅探，避免XSS风险
	// 例如，防止将text/plain文件当作JavaScript执行
	c.Header("X-Content-Type-Options", "nosniff")
}

// 过滤URL参数
func filterURLParams(c *gin.Context, policy *bluemonday.Policy, xssCache *cache.Cache) {
	query := c.Request.URL.Query()
	changed := false

	for _, values := range query {
		for i, value := range values {
			// 检查是否包含XSS攻击模式
			if containsXSSPattern(value) {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "URL参数中检测到可能的XSS攻击内容",
				})
				return
			}

			// 如果启用缓存，先尝试从缓存获取
			var sanitized string
			var found bool

			if xssCache != nil {
				if cachedValue, exists := xssCache.Get(value); exists {
					sanitized = cachedValue.(string)
					found = true
				}
			}

			if !found {
				// 未命中缓存，进行过滤
				sanitized = policy.Sanitize(value)
				if xssCache != nil {
					xssCache.Set(value, sanitized, cache.DefaultExpiration)
				}
			}

			if sanitized != value {
				values[i] = sanitized
				changed = true
			}
		}
	}

	// 如果有参数被修改，更新请求URL
	if changed {
		c.Request.URL.RawQuery = query.Encode()
	}
}

// 过滤请求体
func filterRequestBody(c *gin.Context, policy *bluemonday.Policy, xssCache *cache.Cache) error {
	contentType := c.GetHeader("Content-Type")
	// 只处理相关内容类型
	if !shouldProcessContentType(contentType) {
		return nil
	}

	// 读取请求体
	body, err := c.GetRawData()
	if err != nil {
		return err
	}

	// 检查是否为空
	if len(body) == 0 {
		// 重置空请求体
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		return nil
	}

	// 将请求体转换为字符串
	content := string(body)

	// 检查是否包含XSS攻击模式
	if containsXSSPattern(content) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "请求体中检测到可能的XSS攻击内容",
		})
		return nil
	}

	// 尝试从缓存获取过滤后的内容
	var sanitized string
	var found bool

	if xssCache != nil && len(content) < 1024 { // 只缓存小于1KB的内容
		if cachedValue, exists := xssCache.Get(content); exists {
			sanitized = cachedValue.(string)
			found = true
		}
	}

	if !found {
		// 根据内容类型选择过滤方法
		if strings.Contains(contentType, "json") {
			sanitized = sanitizeJSON(content, policy)
		} else {
			sanitized = policy.Sanitize(content)
		}

		// 更新缓存
		if xssCache != nil && len(content) < 1024 {
			xssCache.Set(content, sanitized, cache.DefaultExpiration)
		}
	}

	// 重新设置请求体
	c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(sanitized)))
	return nil
}

// 检查文本是否包含XSS攻击模式
func containsXSSPattern(text string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// 判断是否应该处理该内容类型
func shouldProcessContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "multipart/form-data")
}

// 优化的响应写入器
type xssResponseWriter struct {
	gin.ResponseWriter
	policy *bluemonday.Policy
	cache  *cache.Cache
}

// 重写Write方法，在写入前过滤内容
func (w *xssResponseWriter) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	contentType := w.Header().Get("Content-Type")
	if !shouldProcessContentType(contentType) {
		return w.ResponseWriter.Write(b)
	}

	// 尝试从缓存获取
	var sanitized []byte
	var found bool

	if w.cache != nil && len(b) < 4096 { // 只缓存小于4KB的响应
		content := string(b)
		if cachedValue, exists := w.cache.Get(content); exists {
			sanitized = []byte(cachedValue.(string))
			found = true
		}
	}

	if !found {
		// 根据内容类型选择过滤方法
		if strings.Contains(contentType, "json") {
			sanitized = []byte(sanitizeJSON(string(b), w.policy))
		} else {
			sanitized = w.policy.SanitizeBytes(b)
		}

		// 更新缓存
		if w.cache != nil && len(b) < 4096 {
			w.cache.Set(string(b), string(sanitized), cache.DefaultExpiration)
		}
	}

	return w.ResponseWriter.Write(sanitized)
}

// JSON特定的净化处理
func sanitizeJSON(jsonStr string, policy *bluemonday.Policy) string {
	// 这里可以实现更复杂的JSON解析和净化
	// 简单实现：对JSON字符串中的特殊字符进行转义
	result := strings.Replace(jsonStr, "<", "\\u003c", -1)
	result = strings.Replace(result, ">", "\\u003e", -1)
	return result
}
