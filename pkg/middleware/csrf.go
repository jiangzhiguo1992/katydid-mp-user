package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CSRF 跨站请求伪造
// 1.服务端生成随机Token，嵌入表单或请求头中，验证请求是否携带合法Token
// 2.不要在cookie里面放敏感数据，防止被外站拿到并携带攻击(head和form不会有这个顾虑)
// 3.只要不在cookie里放敏感数据，就不用做CSRF防护
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 限制跨站请求携带Cookie，白名单可以携带
		c.SetSameSite(http.SameSiteStrictMode)

		c.Next()
	}
}
