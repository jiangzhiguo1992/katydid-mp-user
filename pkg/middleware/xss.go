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
	// XSS攻击模式检测正则表达式
	xssPatterns = []*regexp.Regexp{
		// 脚本标签及其变种
		regexp.MustCompile(`(?i)<\s*script[\s\S]*?>`),
		regexp.MustCompile(`(?i)</\s*script\s*>`),

		// JavaScript 协议变种
		regexp.MustCompile(`(?i)javascript\s*:`),
		regexp.MustCompile(`(?i)vbscript\s*:`),
		regexp.MustCompile(`(?i)livescript\s*:`),
		regexp.MustCompile(`(?i)data\s*:.*script`),

		// JavaScript 事件处理器
		regexp.MustCompile(`(?i)\s+on\w+\s*=`),

		// 危险函数
		regexp.MustCompile(`(?i)\b(eval|setTimeout|setInterval|Function|execScript)\s*\(`),
		regexp.MustCompile(`(?i)(document|window|location|cookie|localStorage)\.(cookie|domain|write|location)`),

		// 内联框架及对象标签
		regexp.MustCompile(`(?i)<\s*(iframe|embed|object|base|applet)\b`),

		// 数据URI
		regexp.MustCompile(`(?i)data:(?:text|image|application)/(?:html|xml|xhtml|svg)`),
		regexp.MustCompile(`(?i)data:.*?;base64`),

		// 表达式和绕过
		regexp.MustCompile(`(?i)expression\s*\(`),
		regexp.MustCompile(`(?i)@import\s+`),
		regexp.MustCompile(`(?i)url\s*\(`),

		// HTML5 特性
		regexp.MustCompile(`(?i)formaction\s*=`),
		regexp.MustCompile(`(?i)srcdoc\s*=`),

		// 元素属性
		regexp.MustCompile(`(?i)\bhref\s*=\s*["']?(?:javascript:|data:text|vbscript:)`),
		regexp.MustCompile(`(?i)\bsrc\s*=\s*["']?(?:javascript:|data:text|vbscript:)`),

		// 常见的HTML注入向量
		regexp.MustCompile(`(?i)<\s*style[^>]*>.*?(expression|behavior|javascript|vbscript).*?</style>`),
		regexp.MustCompile(`(?i)<\s*link[^>]*(?:href|xlink:href)\s*=\s*["']?(?:javascript:|data:text|vbscript:)`),

		// SVG嵌入式脚本
		regexp.MustCompile(`(?i)<\s*svg[^>]*>.*?<\s*script`),
	}

	// 预定义常见危险内容类型
	dangerousContentTypes = map[string]bool{
		"application/json":                  true,
		"application/x-www-form-urlencoded": true,
		"text/html":                         true,
		"text/plain":                        true,
		"multipart/form-data":               true,
	}
)

// XSSConfig XSS防护中间件配置
type XSSConfig struct {
	// 检查严格模式
	StrictMode bool
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
}

// DefaultXSSConfig 默认XSS配置
func DefaultXSSConfig() XSSConfig {
	return XSSConfig{
		StrictMode:           true,
		FilterParams:         true,
		FilterBody:           true,
		FilterResponse:       true,
		EnableCache:          true,
		CacheExpiration:      5 * time.Minute,
		CacheCleanupInterval: 10 * time.Minute,
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
			if ok := checkURLParams(c, policy, xssCache); !ok {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "URL参数中检测到可能的XSS攻击内容",
				})
			}
		}

		// 检查并过滤请求体
		if cfg.FilterBody {
			if ok := checkRequestBody(c, policy, xssCache); !ok {
				ResponseData(c, http.StatusBadRequest, gin.H{
					"error": "无法处理请求内容",
				})
				return
			}
		}

		// 如果需要过滤响应，包装响应写入器
		if cfg.FilterResponse {
			c.Writer = &xssResponseWriter{ResponseWriter: c.Writer, policy: policy, cache: xssCache}
		}

		//if gin.Mode() == gin.DebugMode {
		//	log.InfoFmt("■ ■ Cors ■ ■ 编译正则表达式: %s", pattern)
		//} else {
		//	log.InfoFmtOutput("■ ■ Cors ■ ■ 编译正则表达式: %s", true, pattern)
		//}

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
func checkURLParams(c *gin.Context, policy *bluemonday.Policy, xssCache *cache.Cache) bool {
	query := c.Request.URL.Query()
	changed := false

	for _, values := range query {
		for i, value := range values {
			// 检查是否包含XSS攻击模式
			if containsXSSPattern(value) {
				return false
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
	return true
}

// 过滤请求体
func checkRequestBody(c *gin.Context, policy *bluemonday.Policy, xssCache *cache.Cache) bool {
	// 读取请求体
	body, err := c.GetRawData()
	if err != nil {
		return false
	}

	// 检查是否为空
	if len(body) == 0 {
		// 重置空请求体
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		return true
	}

	// 将请求体转换为字符串
	content := string(body)

	// 检查是否包含XSS攻击模式
	if containsXSSPattern(content) {
		return false
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
		contentType := c.GetHeader("Content-Type")
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
	return false
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

	// 遍历内容类型检查最常见类型
	for dangerousType := range dangerousContentTypes {
		if strings.Contains(contentType, dangerousType) {
			return true
		}
	}
	return false
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
