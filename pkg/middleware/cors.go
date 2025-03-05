package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsOptions CORS 中间件配置选项
type CorsOptions struct {
	// AllowOrigins 允许的源域名列表，使用 * 表示允许所有源
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
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
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

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 处理 Origin
		if origin != "" {
			// 如果允许所有源或源在白名单中
			if contains(opts.AllowOrigins, "*") || contains(opts.AllowOrigins, origin) {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		// 设置其他 CORS 头
		c.Header("Access-Control-Allow-Methods", strings.Join(opts.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ", "))
		c.Header("Access-Control-Max-Age", string(rune(opts.MaxAge)))

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

// contains 检查数组中是否包含指定值
func contains(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}
