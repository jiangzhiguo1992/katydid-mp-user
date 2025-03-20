package middleware

import (
	"katydid-mp-user/pkg/auth"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"net/http"
)

const (
	AuthHeaderToken  = "Authorization"
	AuthHeaderPrefix = "Bearer "

	AuthKeyToken     = "token"
	AuthKeyOwnKind   = "ownKind"
	AuthKeyOwnID     = "ownId"
	AuthKeyUserID    = "userId"
	AuthKeyAccountID = "accountId"
)

// AuthConfig 认证中间件配置
type AuthConfig struct {
	JwtSecret          string        // JWT密钥
	EnableTokenCaching bool          // 是否启用token缓存
	CacheExpiration    time.Duration // 缓存过期时间
	CacheCleanupTime   time.Duration // 缓存清理时间间隔
	IgnorePaths        []string      // 忽略认证的路径
	SkipExpireCheck    bool          // 是否跳过过期检查(开发环境可用)
}

// DefaultAuthConfig 返回默认配置
func DefaultAuthConfig(secret string) AuthConfig {
	return AuthConfig{
		JwtSecret:          secret,
		EnableTokenCaching: true,
		CacheExpiration:    time.Minute * 10,
		CacheCleanupTime:   time.Minute * 30,
		IgnorePaths:        []string{"/health", "/metrics"},
		SkipExpireCheck:    false,
	}
}

var (
	config     AuthConfig
	tokenCache *cache.Cache // token缓存
	blacklist  = sync.Map{} // 黑名单
)

// BlacklistToken 使token失效(加入黑名单)
func BlacklistToken(c *gin.Context) {
	token := c.GetHeader(AuthHeaderToken)
	token = strings.TrimPrefix(token, AuthHeaderPrefix)
	// 加入黑名单
	blacklist.Store(token, time.Now())
	// 从缓存中移除
	tokenCache.Delete(token)
}

// Auth 认证中间件
func Auth(config AuthConfig) gin.HandlerFunc {
	tokenCache = cache.New(config.CacheExpiration, config.CacheCleanupTime)

	return func(c *gin.Context) {
		// 检查是否为忽略路径
		path := c.Request.URL.Path
		for _, ignorePath := range config.IgnorePaths {
			if strings.HasPrefix(path, ignorePath) {
				c.Next()
				return
			}
		}

		authStr := c.GetHeader(AuthHeaderToken)
		// 检查Authorization头是否存在
		if authStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "no_login"})
			c.Abort()
			return
		}

		// 检查Authorization格式是否正确
		if !strings.HasPrefix(authStr, AuthHeaderPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid_token_header"})
			c.Abort()
			return
		}

		// 提取token
		tokenStr := strings.TrimPrefix(authStr, AuthHeaderPrefix)
		if !auth.IsTokenFormat(tokenStr) {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid_token_struct"})
			c.Abort()
			return
		}

		// 检查黑名单
		if _, exists := blacklist.Load(tokenStr); exists {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "token_is_black_list"})
			c.Abort()
			return
		}

		// 验证并解析token
		claims, err := validateAndParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": err.Error()})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set(AuthKeyToken, tokenStr)
		c.Set(AuthKeyOwnKind, claims.OwnKind)
		c.Set(AuthKeyOwnID, claims.OwnID)
		c.Set(AuthKeyAccountID, claims.AccountID)
		c.Set(AuthKeyUserID, claims.UserID)

		c.Next()
	}
}

// validateAndParseToken 验证并解析token
func validateAndParseToken(tokenStr string) (*auth.TokenClaims, error) {
	// 检查缓存
	if config.EnableTokenCaching {
		if claims, found := tokenCache.Get(tokenStr); found {
			return claims.(*auth.TokenClaims), nil
		}
	}

	claims, err := auth.ParseJWT(tokenStr, config.JwtSecret, config.SkipExpireCheck)

	// 存入缓存
	if err != nil {
		if config.EnableTokenCaching {
			tokenCache.Set(tokenStr, claims, config.CacheExpiration)
		}
	}
	return claims, err
}
