package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/auth"
	"katydid-mp-user/pkg/valid"
	"reflect"
	"time"
)

type (
	// Token 令牌
	Token struct {
		*model.Base
		AccessToken  string  `json:"accessToken" validate:"required,format-access"` // 访问token
		RefreshToken *string `json:"refreshToken" validate:"format-refresh"`        // 刷新token

		OwnKind   OwnKind `json:"ownKind" validate:"required,range-own"` // 账号类型
		OwnID     uint64  `json:"ownId" validate:"required"`             // 账号ID
		DeviceID  string  `json:"deviceId" validate:"required"`          // 设备ID
		AccountID uint64  `json:"accountId" validate:"required"`         // 账号ID

		UserID *uint64 `json:"userId"` // 用户ID auths传过来的
		RoleID *uint64 `json:"roleId"` // 角色ID TODO:GG user传过来的? 还是account传过来的? 可以传到token里面
		// TODO:GG 很多ID都要绑定token，方便获取，记得更新也要关联

		AccessExpireAt  int64  `json:"accessExpireAt"`  // 访问token过期时间
		RefreshExpireAt *int64 `json:"refreshExpireAt"` // 刷新token过期时间

		Account *Account `json:"account"` // 账号信息
	}
)

func NewTokenEmpty() *Token {
	return &Token{
		Base: model.NewBase(make(map[string]any)),
	}
}

func NewToken(
	accessToken string, refreshToken *string,
	ownKind OwnKind, ownID uint64, deviceID string, accountID uint64,
	accessExpireAt int64, refreshExpireAt *int64,
) *Token {
	base := model.NewBase(make(map[string]any))
	base.Status = model.StatusInit
	return &Token{
		Base:        base,
		AccessToken: accessToken, RefreshToken: refreshToken,
		OwnKind: ownKind, OwnID: ownID, DeviceID: deviceID, AccountID: accountID,
		AccessExpireAt: accessExpireAt, RefreshExpireAt: refreshExpireAt,
		UserID: nil, RoleID: nil,
	}
}

func (t *Token) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 访问token
			"format-access": func(value reflect.Value, param string) bool {
				val := value.Interface().(string)
				return auth.IsTokenFormat(val)
			},
			// 刷新token
			"format-refresh": func(value reflect.Value, param string) bool {
				val := value.Interface().(*string)
				if val == nil {
					return true
				}
				return auth.IsTokenFormat(*val)
			},
			// 所属类型
			"range-own": func(value reflect.Value, param string) bool {
				val := value.Interface().(OwnKind)
				switch val {
				case OwnKindOrg,
					OwnKindRole,
					OwnKindApp,
					OwnKindClient,
					OwnKindUser:
					return true
				default:
					return false
				}
			},
		},
	}
}

func (t *Token) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"AccessToken": {"required_token_access_err", false, nil},
					"OwnKind":     {"required_token_own_kind_err", false, nil},
					"OwnID":       {"required_token_own_id_err", false, nil},
					"DeviceID":    {"required_token_device_id_err", false, nil},
					"AccountID":   {"required_token_account_id_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"format-access":  {"format_token_access_err", false, nil},
				"format-refresh": {"format_token_refresh_err", false, nil},
				"range-own":      {"range_token_own_err", false, nil},
			},
		},
	}
}

// Generate 生成token
func (t *Token) Generate(
	issuer string, jwtSecret string,
	accessExpireSec int64, refreshExpireHou *int64,
) (*auth.Token, *auth.Token, bool) {
	// 创建新的Access，生成JWT令牌 (传旧的token进去)
	accessToken := auth.NewToken(int16(t.OwnKind), t.OwnID, t.AccountID, t.UserID, issuer, accessExpireSec)
	if err := accessToken.GenerateJWTTokens(jwtSecret, &t.AccessToken); err != nil {
		return nil, nil, false
	}
	t.AccessToken = accessToken.Token
	t.AccessExpireAt = time.Now().Add(time.Duration(accessToken.ExpireSec) * time.Second).Unix()

	// 创建新的Refresh，生成JWT令牌 (传旧的token进去)
	var refreshToken *auth.Token
	if refreshExpireHou != nil {
		// 刷新令牌通常比访问令牌有更长的有效期
		refreshExpireSec := *refreshExpireHou * 3600
		refreshToken = auth.NewToken(int16(t.OwnKind), t.OwnID, t.AccountID, t.UserID, issuer, refreshExpireSec)
		if err := refreshToken.GenerateJWTTokens(jwtSecret, t.RefreshToken); err != nil {
			return nil, nil, false
		}
		t.RefreshToken = &refreshToken.Token
		timeAt := time.Now().Add(time.Duration(refreshToken.ExpireSec) * time.Second).Unix()
		t.RefreshExpireAt = &timeAt
	}
	return accessToken, refreshToken, true
}

// ValidateAccess 验证访问令牌
func (t *Token) ValidateAccess(
	jwtSecret string, checkExpire bool,
) (*auth.TokenClaims, bool) {
	// 检查过期时间
	if checkExpire && t.IsAccessExpired() {
		return nil, false
	}
	// 解析和验证JWT
	checkExpire = checkExpire && (t.AccessExpireAt < 0)
	claims, err := auth.ParseJWT(t.AccessToken, jwtSecret, checkExpire)
	return claims, err != nil
}

// ValidateRefresh 验证刷新令牌
func (t *Token) ValidateRefresh(
	jwtSecret string, checkExpire bool,
) (*auth.TokenClaims, bool) {
	if t.RefreshToken == nil {
		return nil, true // 没有就不检查
	}
	// 检查过期时间
	if checkExpire && t.IsRefreshExpired() {
		return nil, false
	}
	// 解析和验证JWT
	checkExpire = checkExpire && (t.RefreshExpireAt != nil) && (*t.RefreshExpireAt < 0)
	claims, err := auth.ParseJWT(*t.RefreshToken, jwtSecret, checkExpire)
	return claims, err != nil
}

// IsAccessExpired 检查访问token是否过期
func (t *Token) IsAccessExpired() bool {
	if t.AccessExpireAt < 0 { // -1表示永不过期
		return false
	} else if t.AccessExpireAt == 0 {
		return true // 不能使用
	}
	return time.Now().Unix() > t.AccessExpireAt
}

// IsRefreshExpired 检查刷新token是否过期
func (t *Token) IsRefreshExpired() bool {
	if (t.RefreshExpireAt == nil) || (*t.RefreshExpireAt == 0) {
		return true // 不能使用
	} else if *t.RefreshExpireAt < 0 {
		return false // -1表示永不过期
	}
	return time.Now().Unix() > *t.RefreshExpireAt
}
