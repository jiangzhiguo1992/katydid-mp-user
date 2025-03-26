package middleware

import (
	"github.com/gin-gonic/gin"
)

// XSS 跨站脚本攻击
// 1.所有返回HTML的页面都需要检验是否安全，有内嵌脚本
// 2.就不检验所有的入库/展示文本了，不是HTML不用做XSS防护
func XSS() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		c.Next()
	}
}
