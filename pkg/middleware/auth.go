package middleware

import (
	"katydid-mp-user/pkg/auth"
	"katydid-mp-user/pkg/log"
	"regexp"
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
	JwtSecret          string                     // JWT密钥
	EnableTokenCaching bool                       // 是否启用token缓存
	CacheExpiration    time.Duration              // 缓存过期时间
	CacheCleanupTime   time.Duration              // 缓存清理时间间隔
	BlacklistTTL       time.Duration              // 黑名单项过期时间
	BlacklistCleanup   time.Duration              // 黑名单清理间隔
	IgnorePaths        []string                   // 忽略认证的路径
	SkipExpireCheck    bool                       // 是否跳过过期检查(开发环境可用)
	ErrorResponse      func(*gin.Context, string) // 自定义错误响应
}

// DefaultAuthConfig 返回默认配置
func DefaultAuthConfig(secret string) AuthConfig {
	return AuthConfig{
		JwtSecret:          secret,
		EnableTokenCaching: true,
		CacheExpiration:    time.Minute * 10,
		CacheCleanupTime:   time.Minute * 30,
		BlacklistTTL:       time.Hour * 24,
		BlacklistCleanup:   time.Hour,
		IgnorePaths:        []string{"/health", "/metrics"},
		SkipExpireCheck:    false,
		ErrorResponse:      defAuthErrorResponse,
	}
}

// 默认错误响应处理
func defAuthErrorResponse(c *gin.Context, msg string) {
	ResponseData(c, http.StatusUnauthorized, gin.H{"msg": msg})
}

var (
	authConfig AuthConfig

	authTokenCache *cache.Cache // token缓存
	authBlacklist  *cache.Cache // 黑名单

	authRegexps    = make(map[string]*regexp.Regexp) // 正则表达式缓存
	authRegexMutex sync.RWMutex                      // 正则表达式缓存的锁
)

// BlacklistToken 使token失效(加入黑名单)，被storage同步
func BlacklistToken(c *gin.Context) {
	token := c.GetHeader(AuthHeaderToken)
	token = strings.TrimPrefix(token, AuthHeaderPrefix)
	// 加入黑名单，设置过期时间
	authBlacklist.Set(token, time.Now(), authConfig.BlacklistTTL)
	// 从缓存中移除
	authTokenCache.Delete(token)

	if gin.Mode() == gin.DebugMode {
		log.InfoFmt("■ ■ Auth ■ ■ 加入黑名单:%s", maskToken(token))
	} else {
		log.InfoFmtOutput("■ ■ Auth ■ ■ 加入黑名单:%s", true, maskToken(token))
	}
}

// ��护敏感信息 - 掩码处理token值
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// Auth 认证中间件
func Auth(config AuthConfig) gin.HandlerFunc {
	authTokenCache = cache.New(config.CacheExpiration, config.CacheCleanupTime)
	authBlacklist = cache.New(config.BlacklistTTL, config.BlacklistCleanup)

	return func(c *gin.Context) {
		// 检查是否为忽略路径
		path := c.Request.URL.Path
		for _, ignorePath := range config.IgnorePaths {
			if authMatchRegex(path, ignorePath) {
				c.Next()
				return
			}
		}

		authStr := c.GetHeader(AuthHeaderToken)
		// 检查Authorization头是否存在
		if authStr == "" {
			config.ErrorResponse(c, "no_login")
			c.Abort()
			return
		}

		// 检查Authorization格式是否正确
		if !strings.HasPrefix(authStr, AuthHeaderPrefix) {
			config.ErrorResponse(c, "invalid_token_header")
			c.Abort()
			return
		}

		// 提取token
		tokenStr := strings.TrimPrefix(authStr, AuthHeaderPrefix)
		if !auth.IsTokenFormat(tokenStr) {
			config.ErrorResponse(c, "invalid_token_struct")
			c.Abort()
			return
		}

		// 检查黑名单 (只检查内存里的)
		if _, exists := authBlacklist.Get(tokenStr); exists {
			config.ErrorResponse(c, "token_is_black_list")
			c.Abort()
			return
		}

		// 验证并解析token
		claims, err := validateAndParseToken(tokenStr)
		if err != nil {
			config.ErrorResponse(c, err.Error())
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中 (claims里有的)
		c.Set(AuthKeyToken, tokenStr)
		c.Set(AuthKeyOwnKind, claims.OwnKind)
		c.Set(AuthKeyOwnID, claims.OwnID)
		c.Set(AuthKeyAccountID, claims.AccountID)
		c.Set(AuthKeyUserID, claims.UserID)

		log.DebugFmt("■ ■ Auth ■ ■ 设置进Header: %v", claims)

		c.Next()
	}
}

// validateAndParseToken 验证并解析token
func validateAndParseToken(tokenStr string) (*auth.TokenClaims, error) {
	// 检查缓存
	if authConfig.EnableTokenCaching {
		if claims, found := authTokenCache.Get(tokenStr); found {
			log.DebugFmt("■ ■ Auth ■ ■ 缓存命中: %v", claims)
			return claims.(*auth.TokenClaims), nil
		}
	}
	log.DebugFmt("■ ■ Auth ■ ■ 缓存未命中: %s", tokenStr)

	// 解析并验证token
	claims, err := auth.ParseJWT(tokenStr, authConfig.JwtSecret, authConfig.SkipExpireCheck)
	if err != nil {
		return nil, err
	}

	// 只有成功，才存入缓存
	if authConfig.EnableTokenCaching {
		log.DebugFmt("■ ■ Auth ■ ■ 更新缓存: %v", claims)
		authTokenCache.Set(tokenStr, claims, authConfig.CacheExpiration)
	}
	return claims, err
}

// authMatchRegex 判断路径是否匹配正则表达式
func authMatchRegex(path string, pattern string) bool {
	// 如果不是正则表达式，直接前缀匹配
	if !strings.HasPrefix(pattern, "^") && !strings.HasSuffix(pattern, "$") {
		return strings.HasPrefix(path, pattern)
	}

	// 使用正则表达式匹配
	authRegexMutex.RLock()
	re, exists := authRegexps[pattern]
	authRegexMutex.RUnlock()

	// 如果不存在，则编译正则表达式并缓存
	if !exists {
		authRegexMutex.Lock()
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			authRegexMutex.Unlock()
			return false
		}
		authRegexps[pattern] = re
		authRegexMutex.Unlock()
	}
	return re.MatchString(path)
}
