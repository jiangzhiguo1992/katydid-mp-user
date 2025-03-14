package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/data"
	"time"
)

type (
	// Account 账号结构
	Account struct {
		*model.Base
		OwnKind  OwnKind `json:"ownKind"`  // 账号拥有者类型(注册的)
		OwnID    uint64  `json:"ownId"`    // 账号拥有者ID(注册的)
		UserID   *uint64 `json:"userId"`   // 账号拥有者Id (有些org/app不填user)
		Nickname string  `json:"nickname"` // 昵称 (没有user的app/org会用这个，放外面是方便搜索)

		Tokens                map[OwnKind]map[uint64]string `json:"tokens"`        // [OwnKind][OwnID] -> token列表(jwt)
		TokenExpireAts        map[OwnKind]map[uint64]int64  `json:"expireAts"`     // [OwnKind][OwnID] -> token过期时间列表 -1为永久
		RefreshTokens         map[OwnKind]map[uint64]string `json:"refreshTokens"` // [OwnKind][OwnID] -> refreshToken列表(jwt)
		RefreshTokenExpireAts map[OwnKind]map[uint64]int64  `json:"refExpireAts"`  // [OwnKind][OwnID] -> refreshToken过期时间列表 -1为永久

		Auths map[AuthKind][]IAuth `json:"auths,omitempty"` // 认证方式列表 TODO:GG 没有auths了之后，就是unActive了
		//LoginHistory  []*Entry `json:"loginHistory"`  // 登录历史(login)
		//EntryHistory  []any    `json:"entryHistory"`  // 进入历史(entry)
		//AccessHistory []any    `json:"accessHistory"` // 访问历史(api)
	}
)

func NewAccountEmpty() *Account {
	return &Account{
		Base:                  model.NewBaseEmpty(),
		Tokens:                make(map[OwnKind]map[uint64]string),
		TokenExpireAts:        make(map[OwnKind]map[uint64]int64),
		RefreshTokens:         make(map[OwnKind]map[uint64]string),
		RefreshTokenExpireAts: make(map[OwnKind]map[uint64]int64),
		Auths:                 make(map[AuthKind][]IAuth),
	}
}

func NewAccount(
	ownKind OwnKind, ownID uint64, userID *uint64, nickname string,
) *Account {
	return &Account{
		Base:    model.NewBase(make(data.KSMap)),
		OwnKind: ownKind, OwnID: ownID, UserID: userID, Nickname: nickname,
		Tokens:                make(map[OwnKind]map[uint64]string),
		TokenExpireAts:        make(map[OwnKind]map[uint64]int64),
		RefreshTokens:         make(map[OwnKind]map[uint64]string),
		RefreshTokenExpireAts: make(map[OwnKind]map[uint64]int64),
		Auths:                 make(map[AuthKind][]IAuth),
	}
}

const (
	AccountStatusBanned model.Status = -2 // 封禁 (不能访问任何api)
	AccountStatusLocked model.Status = -1 // 锁定 (不能登录)
	AccountStatusInit   model.Status = 0  // 初始
	AccountStatusActive model.Status = 1  // 激活 (必须有附带一个认证才能创建)
)

// CanLogin 可否登录
func (a *Account) CanLogin() bool {
	return a.Status > AccountStatusLocked
}

// CanAccess 可否访问
func (a *Account) CanAccess() bool {
	return a.Status >= AccountStatusActive
}

// GetAvailableAuthKinds 获取可用的认证方式列表
func (a *Account) GetAvailableAuthKinds() []AuthKind {
	kinds := make([]AuthKind, 0)
	for kind, auths := range a.Auths {
		if len(auths) > 0 {
			for _, auth := range auths {
				if auth != nil {
					kinds = append(kinds, kind)
					break
				}
			}
		}
	}
	return kinds
}

// InvalidateToken 使token失效
func (a *Account) InvalidateToken(ownKind OwnKind, ownID uint64) {
	if owns, ok := a.Tokens[ownKind]; ok {
		delete(owns, ownID)
	}
	if owns, ok := a.TokenExpireAts[ownKind]; ok {
		delete(owns, ownID)
	}
	if owns, ok := a.RefreshTokens[ownKind]; ok {
		delete(owns, ownID)
	}
	if owns, ok := a.RefreshTokenExpireAts[ownKind]; ok {
		delete(owns, ownID)
	}
}

// InvalidateAllTokens 使所有token失效
func (a *Account) InvalidateAllTokens() {
	a.Tokens = make(map[OwnKind]map[uint64]string)
	a.TokenExpireAts = make(map[OwnKind]map[uint64]int64)
	a.RefreshTokens = make(map[OwnKind]map[uint64]string)
	a.RefreshTokenExpireAts = make(map[OwnKind]map[uint64]int64)
}

// GetToken 获取token
func (a *Account) GetToken(ownKind OwnKind, ownID uint64) (string, bool) {
	if owns, ok := a.Tokens[ownKind]; ok {
		if token, okk := owns[ownID]; okk {
			return token, true
		}
	}
	return "", false
}

// GetRefreshToken 获取refreshToken
func (a *Account) GetRefreshToken(ownKind OwnKind, ownID uint64) (string, bool) {
	if owns, ok := a.RefreshTokens[ownKind]; ok {
		if token, okk := owns[ownID]; okk {
			return token, true
		}
	}
	return "", false
}

// IsTokenExpired 检查token是否已过期
func (a *Account) IsTokenExpired(ownKind OwnKind, ownID uint64) bool {
	if expireAts, ok := a.TokenExpireAts[ownKind]; ok {
		if expireAt, ok := expireAts[ownID]; ok {
			if expireAt < 0 { // -1表示永不过期
				return false
			}
			return time.Now().Unix() > expireAt
		}
	}
	return true // 如果找不到过期时间，视为已过期
}

// IsRefreshTokenExpired 检查refreshToken是否已过期
func (a *Account) IsRefreshTokenExpired(ownKind OwnKind, ownID uint64) bool {
	if expireAts, ok := a.RefreshTokenExpireAts[ownKind]; ok {
		if expireAt, ok := expireAts[ownID]; ok {
			if expireAt < 0 { // -1表示永不过期
				return false
			}
			return time.Now().Unix() > expireAt
		}
	}
	return true // 如果找不到过期时间，视为已过期
}

// GenerateTokens 为账号生成token
func (a *Account) GenerateTokens(
	ownKind OwnKind, ownID uint64, issuer string,
	jwtSecret string, expireSec, refExpireHou int64,
) (*Token, bool) {
	// 检查账号状态 TODO:GG 移到service?
	if !a.CanLogin() {
		return nil, false
	}
	// 旧的token
	oldToken, _ := a.GetToken(ownKind, ownID)
	// 创建新的Token
	token := NewToken(ownKind, ownID, a.ID, a.UserID, issuer, expireSec, refExpireHou)
	// 生成JWT令牌
	if err := token.GenerateJWTTokens(jwtSecret, &oldToken); err != nil {
		return nil, false
	}
	// 记录token
	if a.Tokens[ownKind] == nil {
		a.Tokens[ownKind] = make(map[uint64]string)
	}
	if a.TokenExpireAts[ownKind] == nil {
		a.TokenExpireAts[ownKind] = make(map[uint64]int64)
	}
	a.Tokens[ownKind][ownID] = token.AccessToken
	a.TokenExpireAts[ownKind][ownID] = time.Now().Add(time.Duration(token.ExpireSec) * time.Second).Unix()
	a.RefreshTokens[ownKind][ownID] = token.RefreshToken
	a.RefreshTokenExpireAts[ownKind][ownID] = time.Now().Add(time.Duration(token.RefExpireHou) * time.Hour).Unix()
	return token, true
}

// ValidateToken 验证token/refreshToken
func (a *Account) ValidateToken(
	ownKind OwnKind, ownID uint64, token string,
	jwtSecret string, checkExpire bool,
) (*TokenClaims, bool) {
	// 检查账号状态 TODO:GG 移到service?
	if !a.CanAccess() {
		return nil, false
	}

	// 验证token是否存在于账户中，是否匹配
	storedToken, exists := a.GetToken(ownKind, ownID)
	if !exists {
		return nil, false
	} else if storedToken != token {
		return nil, false
	}

	// 检查过期时间
	if checkExpire && a.IsTokenExpired(ownKind, ownID) {
		return nil, false
	}
	// 解析和验证JWT
	claims, err := ParseJWT(token, jwtSecret, checkExpire)
	return claims, err != nil
}

// AddRole 添加角色
func (a *Account) AddRole(role string) {
	roles, ok := a.GetRoles()
	if !ok {
		roles = make([]string, 0)
	}
	// 检查角色是否已存在
	for _, r := range roles {
		if r == role {
			return
		}
	}
	roles = append(roles, role)
	a.SetRoles(&roles)
}

// HasRole 检查是否拥有指定角色
func (a *Account) HasRole(role string) bool {
	roles, ok := a.GetRoles()
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

const (
	accExtraKeyAvatarId     = "avatarId"     // 头像ID
	accExtraKeyAvatarUrl    = "avatarUrl"    // 头像URL
	accExtraKeyStatusAt     = "statusAt"     // 状态时间
	accExtraKeyStatusReason = "statusReason" // 状态原因
	accExtraKeyRoles        = "roles"        // 角色列表 (默认只有org下的用户有)
)

func (a *Account) SetAvatarId(avatarId *int64) {
	a.Extra.SetInt64(accExtraKeyAvatarId, avatarId)
}

func (a *Account) GetAvatarId() (int64, bool) {
	return a.Extra.GetInt64(accExtraKeyAvatarId)
}

func (a *Account) SetAvatarUrl(avatarUrl *string) {
	a.Extra.SetString(accExtraKeyAvatarUrl, avatarUrl)
}

func (a *Account) GetAvatarUrl() (string, bool) {
	return a.Extra.GetString(accExtraKeyAvatarUrl)
}

func (a *Account) SetStatusAt(statusAt *int64) {
	a.Extra.SetInt64(accExtraKeyStatusAt, statusAt)
}

func (a *Account) GetStatusAt() (int64, bool) {
	return a.Extra.GetInt64(accExtraKeyStatusAt)
}

func (a *Account) SetStatusReason(reason *string) {
	a.Extra.SetString(accExtraKeyStatusReason, reason)
}

func (a *Account) GetStatusReason() (string, bool) {
	return a.Extra.GetString(accExtraKeyStatusReason)
}

func (a *Account) SetRoles(roles *[]string) {
	a.Extra.SetStringSlice(accExtraKeyRoles, roles)
}

func (a *Account) GetRoles() ([]string, bool) {
	return a.Extra.GetStringSlice(accExtraKeyRoles)
}
