package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
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
	// Debug 是否启用调试日志
	Debug bool

	// 内部使用，预编译的正则表达式
	regexPatterns    []*regexp.Regexp
	wildcardPatterns []*regexp.Regexp
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
		Debug:            false,
	}
}

// Cors 返回 CORS 中间件处理函数
func Cors(options ...*CorsOptions) gin.HandlerFunc {
	// 使用默认配置或用户提供的配置
	var opts *CorsOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = DefaultCorsOptions()
	}

	// 初始化并预编译所有正则表达式
	opts.compilePatterns()

	return opts.handleRequest
}

// 编译所有正则表达式模式
func (opts *CorsOptions) compilePatterns() {
	opts.regexPatterns = make([]*regexp.Regexp, 0)
	opts.wildcardPatterns = make([]*regexp.Regexp, 0)

	for _, origin := range opts.AllowOrigins {
		// 处理正则模式
		if strings.HasPrefix(origin, "regex:") {
			pattern := strings.TrimPrefix(origin, "regex:")
			if re, err := regexp.Compile(pattern); err == nil {
				opts.regexPatterns = append(opts.regexPatterns, re)
				if opts.Debug {
					slog.Info("■ ■ cors ■ ■ 编译正则表达式: %s", pattern)
				}
			} else if opts.Debug {
				slog.Error("■ ■ cors ■ ■ 无法编译正则表达式 %s: %v", pattern, err)
			}
		} else if strings.Contains(origin, "*") {
			// 处理通配符模式
			pattern := strings.Replace(origin, ".", "\\.", -1)
			pattern = strings.Replace(pattern, "*", ".*", -1)
			pattern = "^" + pattern + "$"
			if re, err := regexp.Compile(pattern); err == nil {
				opts.wildcardPatterns = append(opts.wildcardPatterns, re)
				if opts.Debug {
					slog.Info("■ ■ cors ■ ■ 编译通配符模式: %s -> %s", origin, pattern)
				}
			}
		}
	}
}

// 处理 CORS 请求
func (opts *CorsOptions) handleRequest(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")

	// 如果没有 Origin 头，可能不是跨域请求
	if origin == "" {
		c.Next()
		return
	}

	// 检查是否允许该源
	if !opts.isOriginAllowed(origin) {
		if opts.Debug {
			slog.Warn("■ ■ cors ■ ■ 拒绝源: %s", origin)
		}
		c.Next()
		return
	}

	if opts.Debug {
		slog.Info("■ ■ cors ■ ■ 接受源: %s", origin)
	}

	// 设置 CORS 头
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", strings.Join(opts.AllowMethods, ", "))
	c.Header("Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ", "))
	c.Header("Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ", "))
	c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", opts.MaxAge))
	c.Header("Vary", "Origin")

	// 处理认证信息
	if opts.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// 处理预检请求
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

// 检查源是否被允许
func (opts *CorsOptions) isOriginAllowed(origin string) bool {
	// 允许所有源
	if contains(opts.AllowOrigins, "*") {
		return true
	}

	// 直接匹配
	if contains(opts.AllowOrigins, origin) {
		return true
	}

	// 正则表达式匹配
	for _, re := range opts.regexPatterns {
		if re.MatchString(origin) {
			return true
		}
	}

	// 通配符匹配
	for _, re := range opts.wildcardPatterns {
		if re.MatchString(origin) {
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
