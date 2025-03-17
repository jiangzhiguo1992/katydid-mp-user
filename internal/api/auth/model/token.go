package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/auth"
	"time"
)

type (
	// Token 令牌
	Token struct {
		*model.Base
		Token string `json:"token"` // token

		OwnKind   OwnKind `json:"ownKind"`   // 账号类型
		OwnID     uint64  `json:"ownId"`     // 账号ID
		DeviceID  string  `json:"deviceId"`  // 设备ID
		AccountID uint64  `json:"accountId"` // 账号ID

		UserID *uint64 `json:"userId"` // 用户ID auths传过来的
		RoleID *uint64 `json:"roleId"` // 角色ID TODO:GG user传过来的? 还是account传过来的? 可以传到token里面
		// TODO:GG 很多ID都要绑定token，方便获取，记得更新也要关联

		ExpireAt int64 `json:"expireAt"` // 过期时间

		Account *Account `json:"account,omitempty"` // 账号信息
	}
)

func NewTokenEmpty() *Token {
	return &Token{
		Base: model.NewBase(make(map[string]any)),
	}
}

func NewToken(
	token string,
	ownKind OwnKind, ownID uint64, deviceID string, accountID uint64,
	expireAt int64,
) *Token {
	return &Token{
		Base:      model.NewBase(make(map[string]any)),
		AccountID: accountID, OwnKind: ownKind, OwnID: ownID,
		DeviceID: deviceID, Token: token, ExpireAt: expireAt,
	}
}

// IsExpired 检查token是否已过期
func (t *Token) IsExpired() bool {
	if t.ExpireAt < 0 { // -1表示永不过期
		return false
	} else if t.ExpireAt == 0 {
		return true // 不能使用
	}
	return time.Now().Unix() > t.ExpireAt
}

// Generate 生成token
func (t *Token) Generate(
	issuer string, jwtSecret string,
	expireSec int64,
) (*auth.Token, bool) {
	// 创建新的Token
	token := auth.NewToken(int16(t.OwnKind), t.OwnID, t.AccountID, t.UserID, issuer, expireSec)
	// 生成JWT令牌 (传旧的token进去)
	if err := token.GenerateJWTTokens(jwtSecret, &t.Token); err != nil {
		return nil, false
	}
	// 记录token
	t.Token = token.Token
	t.ExpireAt = time.Now().Add(time.Duration(token.ExpireSec) * time.Second).Unix()
	return token, true
}

// ValidateToken 验证token
func (t *Token) ValidateToken(
	token string, jwtSecret string, checkExpire bool,
) (*auth.TokenClaims, bool) {
	// 验证是否匹配
	if t.Token != token {
		return nil, false
	}
	// 检查过期时间
	if checkExpire && t.IsExpired() {
		return nil, false
	}
	// 解析和验证JWT
	claims, err := auth.ParseJWT(token, jwtSecret, checkExpire)
	return claims, err != nil
}
