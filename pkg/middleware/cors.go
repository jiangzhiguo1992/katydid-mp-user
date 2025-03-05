package middleware

import (
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsOptions CORS 中间件配置选项
type CorsOptions struct {
	// AllowOrigins 允许的源域名列表，使用 * 表示允许所有源
	// 支持通配符: example.* 或正则: regex:^https://.*example\.com$
	AllowOrigins []string
	// AllowMethods 允许的 HTTP 方法
	AllowMethods []string
	// AllowHeaders 允许的 HTTP 头
	AllowHeaders []string
	// ExposeHeaders 允许客户端访问的响应头
	ExposeHeaders []string
	// AllowCredentials 是否允许携带认证信息(cookies)
	AllowCredentials bool
	// MaxAge 预检请求结果缓存时间(秒)
	MaxAge int
}

// DefaultCorsOptions 默认 CORS 配置
func DefaultCorsOptions() *CorsOptions {
	return &CorsOptions{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// Cors 返回 CORS 中间件处理函数
func Cors(options ...*CorsOptions) gin.HandlerFunc {
	// 使用默认配置或用户提供的配置
	opts := DefaultCorsOptions()
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	}

	// 预编译正则表达式以提高性能
	regexps := compileRegexPatterns(opts.AllowOrigins)

	// 检查是否允许所有源
	allowAll := contains(opts.AllowOrigins, "*")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 处理 Origin
		if origin != "" {
			allowed := false

			if allowAll {
				allowed = true
			} else {
				allowed = isOriginAllowed(origin, opts.AllowOrigins, regexps)
			}

			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		// 设置其他 CORS 头
		c.Header("Access-Control-Allow-Methods", strings.Join(opts.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ", "))
		c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", opts.MaxAge))

		// 设置安全相关头部
		c.Header("Vary", "Origin")

		if opts.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 继续处理其他请求
		c.Next()
	}
}

// 编译正则表达式模式
func compileRegexPatterns(origins []string) map[string]*regexp.Regexp {
	regexps := make(map[string]*regexp.Regexp)

	for _, origin := range origins {
		if strings.HasPrefix(origin, "regex:") {
			pattern := strings.TrimPrefix(origin, "regex:")
			if re, err := regexp.Compile(pattern); err == nil {
				regexps[origin] = re
			}
		}
	}

	return regexps
}

// 检查源是否被允许
func isOriginAllowed(origin string, allowedOrigins []string, regexps map[string]*regexp.Regexp) bool {
	// 直接匹配
	if contains(allowedOrigins, origin) {
		return true
	}

	// 通配符匹配
	for _, allowed := range allowedOrigins {
		// 处理通配符情况
		if strings.Contains(allowed, "*") {
			pattern := strings.Replace(allowed, ".", "\\.", -1)
			pattern = strings.Replace(pattern, "*", ".*", -1)
			pattern = "^" + pattern + "$"
			if matched, _ := regexp.MatchString(pattern, origin); matched {
				return true
			}
			continue
		}

		// 处理正则情况
		if strings.HasPrefix(allowed, "regex:") {
			if re, exists := regexps[allowed]; exists && re.MatchString(origin) {
				return true
			}
			continue
		}

		// 处理子域名情况
		if strings.HasPrefix(origin, allowed) || path.Dir(origin) == allowed {
			return true
		}
	}

	return false
}

// contains 检查数组中是否包含指定值
func contains(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}
