package model

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType 令牌类型
type TokenType string

const (
	TokenTypeBasic  TokenType = "Basic"  // Basic令牌类型
	TokenTypeBearer TokenType = "Bearer" // Bearer令牌类型
)

// TokenOwn 令牌拥有者类型
type TokenOwn string

const (
	TokenOwnOrg TokenOwn = "org" // 组织类型
	TokenOwnApp TokenOwn = "app" // 应用类型
)

// Token JWT令牌模型
type Token struct {
	Code         string    `json:"code,omitempty"`         // 授权码(OAuth2流程使用)
	TokenType    TokenType `json:"tokenType"`              // 令牌类型
	IssuedAt     int64     `json:"issuedAt"`               // 签发时间
	AccessToken  string    `json:"accessToken"`            // 访问令牌
	RefreshToken string    `json:"refreshToken,omitempty"` // 刷新令牌
	ExpireSec    int64     `json:"expireSec"`              // 过期时间(秒)
	RefExpireHou int64     `json:"refExpireHou"`           // 刷新过期时间(时)

	Claims *TokenClaims `json:"-"` // 令牌声明(不序列化)
}

// TokenClaims JWT的payload结构
type TokenClaims struct {
	AccountID uint64   `json:"accountId,omitempty"` // 账号ID
	OwnType   TokenOwn `json:"ownType,omitempty"`   // 令牌拥有者类型
	OwnID     uint64   `json:"ownId,omitempty"`     // 令牌拥有者ID
	jwt.RegisteredClaims
}

// NewToken 创建一个新的Token实例
func NewToken(
	accountID uint64, ownType TokenOwn, ownID uint64,
	expireSec, refExpireHou int64, issuer string,
) *Token {
	token := &Token{
		TokenType:    TokenTypeBearer,
		IssuedAt:     time.Now().Unix(),
		AccessToken:  "", // 将由GenerateJWT方法填充
		RefreshToken: "", // 将由GenerateRefreshToken方法填充
		ExpireSec:    expireSec,
		RefExpireHou: refExpireHou,
	}
	token.Claims = NewTokenClaims(accountID, ownType, ownID, expireSec, issuer)
	return token
}

func NewTokenClaims(
	accountID uint64, ownType TokenOwn, ownID uint64,
	expireSec int64, issuer string,
) *TokenClaims {
	now := time.Now()
	// 计算过期时间
	var expiresAt *jwt.NumericDate
	if expireSec > 0 {
		expiresAt = jwt.NewNumericDate(now.Add(time.Duration(expireSec) * time.Second))
	}
	return &TokenClaims{
		AccountID: accountID,
		OwnType:   ownType,
		OwnID:     ownID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   fmt.Sprintf("%d", accountID),
			ExpiresAt: expiresAt,
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
}

// GenerateAccessJWTToken 生成访问令牌
func (t *Token) GenerateAccessJWTToken(secret string) error {
	accessClaims := NewTokenClaims(t.Claims.AccountID, t.Claims.OwnType, t.Claims.OwnID, t.ExpireSec, t.Claims.Issuer)
	var err error
	t.AccessToken, err = accessClaims.GenerateJWTToken(secret)
	return err
}

// GenerateRefreshJWTToken 生成刷新令牌
func (t *Token) GenerateRefreshJWTToken(secret string) error {
	// 刷新令牌通常比访问令牌有更长的有效期
	expireSec := t.RefExpireHou * 3600
	refreshClaims := NewTokenClaims(t.Claims.AccountID, t.Claims.OwnType, t.Claims.OwnID, expireSec, t.Claims.Issuer)
	var err error
	t.RefreshToken, err = refreshClaims.GenerateJWTToken(secret)
	return err
}

// GenerateJWTToken 使用提供的密钥生成JWT令牌
func (tc *TokenClaims) GenerateJWTToken(secret string) (string, error) {
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
		if checkExpire {
			if time.Now().After(claims.ExpiresAt.Time) {
				return nil, fmt.Errorf("token已过期")
			}
		}
		return claims, nil
	}
	return nil, fmt.Errorf("无效的token结构")
}
