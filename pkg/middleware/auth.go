package middleware

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/patrickmn/go-cache"
	"net/http"
)

const (
	AuthHeaderPrefix = "Bearer "
	AuthKeyUserID    = "userId"
	AuthKeyUsername  = "username"
	AuthKeyRoles     = "roles"
)

// 定义错误信息常量
const (
	ErrMsgMissingAuthHeader = "缺少Authorization Header"
	ErrMsgInvalidAuthFormat = "无效的Authorization Header格式"
	ErrMsgInvalidToken      = "无效的token"
	ErrMsgTokenExpired      = "token已过期"
	ErrMsgTokenInBlacklist  = "token已被禁用"
)

// TokenClaims 定义JWT中的自定义声明
type TokenClaims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthConfig 认证中间件配置
type AuthConfig struct {
	JwtSecret           string        // JWT密钥
	TokenExpiration     time.Duration // Token过期时间
	CacheExpiration     time.Duration // 缓存过期时间
	CacheCleanupTime    time.Duration // 缓存清理时间间隔
	EnableTokenCaching  bool          // 是否启用token缓存
	IgnorePaths         []string      // 忽略认证的路径
	EnableRoleChecking  bool          // 是否启用角色检查
	SkipExpirationCheck bool          // 是否跳过过期检查(开发环境可用)
}

// DefaultAuthConfig 返回默认配置
func DefaultAuthConfig(secret string, debug bool) AuthConfig {
	return AuthConfig{
		JwtSecret:           secret,
		TokenExpiration:     time.Hour * 24,
		CacheExpiration:     time.Minute * 10,
		CacheCleanupTime:    time.Minute * 30,
		EnableTokenCaching:  true,
		IgnorePaths:         []string{"/health", "/metrics"},
		EnableRoleChecking:  true,
		SkipExpirationCheck: debug,
	}
}

// AuthManager 认证管理器
type AuthManager struct {
	config     AuthConfig
	tokenCache *cache.Cache
	blacklist  sync.Map
}

// NewAuthManager 创建认证管理器实例
func NewAuthManager(config AuthConfig) *AuthManager {
	return &AuthManager{
		config:     config,
		tokenCache: cache.New(config.CacheExpiration, config.CacheCleanupTime),
		blacklist:  sync.Map{},
	}
}

// Auth 认证中间件
func (am *AuthManager) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为忽略路径
		path := c.Request.URL.Path
		for _, ignorePath := range am.config.IgnorePaths {
			if strings.HasPrefix(path, ignorePath) {
				c.Next()
				return
			}
		}

		auth := c.GetHeader("Authorization")
		// 检查Authorization头是否存在
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": ErrMsgMissingAuthHeader})
			c.Abort()
			return
		}

		// 检查Authorization格式是否正确
		if !strings.HasPrefix(auth, AuthHeaderPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": ErrMsgInvalidAuthFormat})
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(auth, AuthHeaderPrefix)

		// 检查黑名单
		if _, exists := am.blacklist.Load(tokenString); exists {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": ErrMsgTokenInBlacklist})
			c.Abort()
			return
		}

		// 验证并解析token
		claims, err := am.validateAndParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": err.Error()})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set(AuthKeyUserID, claims.UserID)
		c.Set(AuthKeyUsername, claims.Username)
		c.Set(AuthKeyRoles, claims.Roles)

		c.Next()
	}
}

// validateAndParseToken 验证并解析token
func (am *AuthManager) validateAndParseToken(tokenString string) (*TokenClaims, error) {
	// 检查缓存
	if am.config.EnableTokenCaching {
		if claims, found := am.tokenCache.Get(tokenString); found {
			return claims.(*TokenClaims), nil
		}
	}

	// 解析Token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		// 确保token的签名方法是我们期望的
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(am.config.JwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrMsgInvalidToken, err)
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		// 检查是否过期
		if !am.config.SkipExpirationCheck {
			if time.Now().After(claims.ExpiresAt.Time) {
				return nil, fmt.Errorf(ErrMsgTokenExpired)
			}
		}

		// 存入缓存
		if am.config.EnableTokenCaching {
			am.tokenCache.Set(tokenString, claims, am.config.CacheExpiration)
		}

		return claims, nil
	}
	return nil, fmt.Errorf(ErrMsgInvalidToken)
}

// GenerateToken 生成JWT token
func (am *AuthManager) GenerateToken(userID, username string, roles []string) (string, error) {
	now := time.Now()
	claims := TokenClaims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(am.config.TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "api-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.config.JwtSecret))
}

// InvalidateToken 使token失效(加入黑名单)
func (am *AuthManager) InvalidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	token = strings.TrimPrefix(token, AuthHeaderPrefix)
	// 加入黑名单
	am.blacklist.Store(token, time.Now())
	// 从缓存中移除
	am.tokenCache.Delete(token)
}

// RoleCheck 检查用户是否具有特定角色的中间件
func (am *AuthManager) RoleCheck(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !am.config.EnableRoleChecking {
			c.Next()
			return
		}

		// 获取上下文中的角色信息
		roles, exists := c.Get(AuthKeyRoles)
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"msg": "未找到角色信息"})
			c.Abort()
			return
		}

		userRoles := roles.([]string)
		hasRole := false

		for _, role := range userRoles {
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"msg": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}
