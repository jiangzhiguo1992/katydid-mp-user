package model

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenKindBasic  TokenKind = "Basic"  // Basic令牌类型
	TokenKindBearer TokenKind = "Bearer" // Bearer令牌类型
	TokenKindJWT    TokenKind = "JWT"    // JWT令牌类型
)

const (
	TokenOwnOrg    TokenOwn = "org"    // 组织类型
	TokenOwnApp    TokenOwn = "app"    // 应用类型
	TokenOwnClient TokenOwn = "client" // 客户端类型
)

type (
	// Token JWT令牌模型
	Token struct {
		Code         string    `json:"code,omitempty"`         // 授权码(OAuth2流程使用)
		Kind         TokenKind `json:"kind"`                   // 令牌类型
		IssuedAt     int64     `json:"issuedAt"`               // 签发时间
		AccessToken  string    `json:"accessToken"`            // 访问令牌
		RefreshToken string    `json:"refreshToken,omitempty"` // 刷新令牌
		ExpireSec    int64     `json:"expireSec"`              // 过期时间(秒)
		RefExpireHou int64     `json:"refExpireHou"`           // 刷新过期时间(时)

		Claims *TokenClaims `json:"-"` // 令牌声明(不序列化)
	}

	// TokenKind 令牌类型
	TokenKind string

	// TokenClaims JWT的payload结构
	TokenClaims struct {
		TokenID   string   `json:"jti,omitempty"`       // JWT唯一标识符
		AccountID uint64   `json:"accountId,omitempty"` // 账号ID
		UserID    *uint64  `json:"userId,omitempty"`    // 用户ID
		OwnKind   TokenOwn `json:"ownKind,omitempty"`   // 令牌拥有者类型
		OwnID     uint64   `json:"ownId,omitempty"`     // 令牌拥有者ID
		// TODO:GG roles
		jwt.RegisteredClaims
	}
	// TokenOwn 令牌拥有者类型
	TokenOwn string
)

// NewToken 创建一个新的Token实例
func NewToken(
	accountID uint64, userID *uint64, ownKind TokenOwn, ownID uint64,
	expireSec, refExpireHou int64, issuer string,
) *Token {
	token := &Token{
		Kind:         TokenKindBearer,
		IssuedAt:     time.Now().Unix(),
		AccessToken:  "", // 将由GenerateJWT方法填充
		RefreshToken: "", // 将由GenerateRefreshToken方法填充
		ExpireSec:    expireSec,
		RefExpireHou: refExpireHou,
	}
	token.Claims = NewTokenClaims(accountID, userID, ownKind, ownID, expireSec, issuer)
	return token
}

func NewTokenClaims(
	accountID uint64, userID *uint64, ownKind TokenOwn, ownID uint64,
	expireSec int64, issuer string,
) *TokenClaims {
	// 生成唯一令牌ID
	tokenID, _ := generateSecureRandomString(16)
	// 计算过期时间
	now := time.Now()
	var expiresAt *jwt.NumericDate
	if expireSec > 0 {
		expiresAt = jwt.NewNumericDate(now.Add(time.Duration(expireSec) * time.Second))
	}
	return &TokenClaims{
		TokenID:   tokenID,
		AccountID: accountID,
		UserID:    userID,
		OwnKind:   ownKind,
		OwnID:     ownID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   fmt.Sprintf("%d", accountID),
			ExpiresAt: expiresAt,
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}
}

// GenerateJWTTokens 生成所有访问令牌
func (t *Token) GenerateJWTTokens(secret string, oldToken *string) error {
	var tokenID *string
	if oldToken != nil && *oldToken != "" {
		// 解析刷新令牌但不检查过期
		claims, err := ParseJWT(*oldToken, secret, false)
		if err != nil {
			return err
		}
		tokenID = &claims.RegisteredClaims.ID
	}
	err := t.generateAccessJWTToken(secret, tokenID)
	if err != nil {
		return err
	}
	return t.generateRefreshJWTToken(secret, tokenID)
}

// generateAccessJWTToken 生成访问令牌
func (t *Token) generateAccessJWTToken(secret string, tokenID *string) error {
	claims := NewTokenClaims(t.Claims.AccountID, t.Claims.UserID, t.Claims.OwnKind, t.Claims.OwnID, t.ExpireSec, t.Claims.Issuer)
	if tokenID != nil && *tokenID != "" {
		claims.TokenID = *tokenID             // 使用相同的TokenID
		claims.RegisteredClaims.ID = *tokenID // 使用相同的TokenID
	}
	var err error
	t.AccessToken, err = claims.generateJWTToken(secret)
	return err
}

// generateRefreshJWTToken 生成刷新令牌
func (t *Token) generateRefreshJWTToken(secret string, tokenID *string) error {
	// 刷新令牌通常比访问令牌有更长的有效期
	expireSec := t.RefExpireHou * 3600
	claims := NewTokenClaims(t.Claims.AccountID, t.Claims.UserID, t.Claims.OwnKind, t.Claims.OwnID, expireSec, t.Claims.Issuer)
	if tokenID != nil && *tokenID != "" {
		claims.TokenID = *tokenID             // 使用相同的TokenID
		claims.RegisteredClaims.ID = *tokenID // 使用相同的TokenID
	} else {
		claims.TokenID = t.Claims.TokenID             // 使用相同的TokenID
		claims.RegisteredClaims.ID = t.Claims.TokenID // 使用相同的TokenID
	}
	var err error
	t.RefreshToken, err = claims.generateJWTToken(secret)
	return err
}

// GenerateJWTToken 使用提供的密钥生成JWT令牌
func (tc *TokenClaims) generateJWTToken(secret string) (string, error) {
	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tc)
	// 签名token
	return token.SignedString([]byte(secret))
}

// IsExpired 检查令牌是否过期
func (t *Token) IsExpired() bool {
	if t.Claims == nil || t.Claims.ExpiresAt == nil {
		return true
	}
	return time.Now().After(t.Claims.ExpiresAt.Time)
}

// IsRefreshExpired 检查刷新令牌是否过期
func (t *Token) IsRefreshExpired() bool {
	if t.Claims == nil || t.Claims.ExpiresAt == nil {
		return true
	}
	return time.Now().After(t.Claims.ExpiresAt.Time.Add(time.Duration(t.RefExpireHou) * time.Hour))
}

// generateSecureRandomString 生成安全的随机字符串
func generateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// ParseJWT 解析JWT令牌
func ParseJWT(tokenStr string, secret string, checkExpire bool) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 确保token的签名方法是我们期望的
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("无效的token: %w", err)
	}
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		// 检查是否过期
		if checkExpire && (claims.ExpiresAt != nil) {
			if time.Now().After(claims.ExpiresAt.Time) {
				return nil, fmt.Errorf("token已过期")
			}
		}
		return claims, nil
	}
	return nil, fmt.Errorf("无效的token结构")
}
