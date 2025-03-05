package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")

		// 检查Authorization头是否存在
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "缺少Authorization Header",
			})
			c.Abort()
			return
		}

		// 检查Authorization格式是否正确
		if !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "无效的Authorization Header",
			})
			c.Abort()
			return
		}

		// 提取并验证token
		token := strings.TrimPrefix(auth, "Bearer ")
		if !validateToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "无效的token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateToken 验证token是否有效
// 实际应用中，应该使用JWT库进行proper的token验证
func validateToken(token string) bool {
	// 在实际应用中，这里应该使用JWT库来验证token
	// 例如: return jwt.ValidateToken(token)

	// 临时使用硬编码的token进行验证
	return token == "jwt_token"
}
