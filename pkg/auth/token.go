package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenKind 令牌类型
type TokenKind string

const (
	TokenKindBasic  TokenKind = "Basic"  // Basic令牌类型
	TokenKindBearer TokenKind = "Bearer" // Bearer令牌类型
)

var SigningMethod = jwt.SigningMethodHS256 // 签名方法

type (
	// Token JWT令牌模型
	Token struct {
		Code     string    `json:"code,omitempty"` // 授权码(OAuth2流程使用)
		Kind     TokenKind `json:"kind"`           // 令牌类型
		IssuedAt int64     `json:"issuedAt"`       // 签发时间

		Token     string `json:"token"`     // 访问令牌
		ExpireSec int64  `json:"expireSec"` // 过期时间(秒) (-1就没有不过期间)

		Claims *TokenClaims `json:"-"` // 令牌声明(不序列化)
	}

	// TokenClaims JWT的payload结构
	TokenClaims struct {
		TokenID string `json:"jti,omitempty"` // JWT唯一标识符

		OwnKind   int16   `json:"ownKind,omitempty"`   // 令牌拥有者类型
		OwnID     uint64  `json:"ownId,omitempty"`     // 令牌拥有者ID
		AccountID uint64  `json:"accountId,omitempty"` // 账号ID
		UserID    *uint64 `json:"userId,omitempty"`    // 用户ID
		// TODO:GG roles (记得加到middleware里)

		jwt.RegisteredClaims `json:"-"` // 注册声明(不序列化)
	}
)

// NewToken 创建一个新的Token实例
func NewToken(
	ownKind int16, ownID uint64, accountID uint64, userID *uint64, issuer string,
	expireSec int64,
) *Token {
	token := &Token{
		Kind:      TokenKindBearer,
		IssuedAt:  time.Now().Unix(),
		Token:     "", // 将由GenerateJWT方法填充
		ExpireSec: expireSec,
	}
	token.Claims = NewTokenClaims(ownKind, ownID, accountID, userID, issuer, expireSec)
	return token
}

func NewTokenClaims(
	ownKind int16, ownID uint64, accountID uint64, userID *uint64,
	issuer string, expireSec int64,
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
		TokenID: tokenID,
		OwnKind: ownKind, OwnID: ownID, AccountID: accountID, UserID: userID,
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
	return t.generateJWTToken(secret, tokenID)
}

// generateJWTToken 生成访问令牌
func (t *Token) generateJWTToken(secret string, tokenID *string) error {
	claims := NewTokenClaims(t.Claims.OwnKind, t.Claims.OwnID, t.Claims.AccountID, t.Claims.UserID, t.Claims.Issuer, t.ExpireSec)

	// 设置令牌ID
	if tokenID != nil && *tokenID != "" {
		claims.TokenID = *tokenID
	} else {
		claims.TokenID = t.Claims.TokenID // 使用相同的TokenID
	}
	claims.RegisteredClaims.ID = claims.TokenID // 确保两处ID一致

	var err error
	t.Token, err = claims.generateJWTToken(secret)
	return err
}

// GenerateJWTToken 使用提供的密钥生成JWT令牌
func (tc *TokenClaims) generateJWTToken(secret string) (string, error) {
	// 创建token
	token := jwt.NewWithClaims(SigningMethod, tc)
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

// generateSecureRandomString 生成安全的随机字符串
func generateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(bytes)
	if len(encoded) < length {
		return "", fmt.Errorf("token_too_short")
	}
	return encoded[:length], nil
}

// ParseJWT 解析JWT令牌
func ParseJWT(tokenStr string, secret string, checkExpire bool) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		// 确保token的签名方法是我们期望的
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid_token_sign_method") // token.Header["alg"]
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err // "无效的token: %w"
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		// 检查是否过期 (只是token自带的过期，外部应该还会检查，灵活机制)
		if checkExpire && (claims.ExpiresAt != nil) {
			if time.Now().After(claims.ExpiresAt.Time) {
				return nil, fmt.Errorf("token_is_expire")
			}
		}
		return claims, nil
	}
	return nil, fmt.Errorf("invalid_token_struct")
}

// IsTokenFormat 检查格式 (header.payload.signature)
func IsTokenFormat(tokenStr string) bool {
	// 检查是否为空
	if tokenStr == "" {
		return false
	}

	// 检查是否包含两个点分隔符
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return false
	}

	// 验证头部和载荷部分是有效的 base64url 编码
	// 注意：签名部分可能包含填充，验证方式稍有不同
	for _, part := range parts[:2] {
		if part == "" {
			return false
		}

		// 添加可能缺少的填充
		padded := part
		switch len(part) % 4 {
		case 2:
			padded += "=="
		case 3:
			padded += "="
		}

		// 检查每个部分是否是有效的base64URL编码
		if _, err := base64.RawURLEncoding.DecodeString(padded); err != nil {
			return false
		}
	}

	// 简单检查签名部分非空
	if parts[2] == "" {
		return false
	}
	return true
}
