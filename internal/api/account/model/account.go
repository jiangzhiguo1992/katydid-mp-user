package model

import (
	"fmt"
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/utils"
	"time"
)

// Account 账号结构
type Account struct {
	*model.Base

	UserID *uint64 `json:"userId"` // 账号拥有者Id (有些org/app不填user)

	Nickname string        `json:"nickname"` // 昵称 (没有user的app/org会用这个，放外面是方便搜索)
	Status   AccountStatus `json:"status"`   // 账号状态

	OrgActiveAts map[uint64]int64  `json:"orgActiveAts"` // [org]激活组织时间集合 (最早的就是注册的平台)
	OrgTokens    map[uint64]string `json:"orgTokens"`    // [org]token列表(jwt)
	OrgExpireAts map[uint64]int64  `json:"orgExpireAts"` // [org]token过期时间列表 -1为永久
	AppActiveAts map[uint64]int64  `json:"appActiveAts"` // [app]激活应用时间集合 (最早的就是注册的平台)
	AppTokens    map[uint64]string `json:"appTokens"`    // [app]token列表(jwt)
	AppExpireAts map[uint64]int64  `json:"appExpireAts"` // [app]token过期时间列表 -1为永久

	// 认证方式引用 TODO:GG 是否可认证和必须认证的方式，由org/client决定
	AuthPassword     *UsernamePwd      `json:"authPassword,omitempty"`     // 用户名密码认证
	AuthPhone        *PhoneNumber      `json:"authPhone,omitempty"`        // 手机号认证
	AuthEmail        *EmailAddress     `json:"authEmail,omitempty"`        // 邮箱认证
	AuthBiometrics   []*Biometric      `json:"authBiometrics,omitempty"`   // 生物特征认证列表
	AuthThirdParties []*ThirdPartyLink `json:"authThirdParties,omitempty"` // 第三方认证列表

	LoginHistory []*LoginInfo `json:"loginHistory,omitempty"` // 登录历史
}

func NewAccountEmpty() *Account {
	return &Account{
		Base:             model.NewBaseEmpty(),
		OrgTokens:        make(map[uint64]string),
		OrgActiveAts:     make(map[uint64]int64),
		OrgExpireAts:     make(map[uint64]int64),
		AppTokens:        make(map[uint64]string),
		AppActiveAts:     make(map[uint64]int64),
		AppExpireAts:     make(map[uint64]int64),
		AuthBiometrics:   make([]*Biometric, 0),
		AuthThirdParties: make([]*ThirdPartyLink, 0),
	}
}

func NewAccount(nickname string) *Account {
	return &Account{
		Base:             model.NewBase(make(utils.KSMap)),
		Nickname:         nickname,
		Status:           AccountStatusNormal,
		OrgTokens:        make(map[uint64]string),
		OrgActiveAts:     make(map[uint64]int64),
		OrgExpireAts:     make(map[uint64]int64),
		AppTokens:        make(map[uint64]string),
		AppActiveAts:     make(map[uint64]int64),
		AppExpireAts:     make(map[uint64]int64),
		AuthPassword:     nil,
		AuthPhone:        nil,
		AuthEmail:        nil,
		AuthBiometrics:   make([]*Biometric, 0),
		AuthThirdParties: make([]*ThirdPartyLink, 0),
	}
}

// AccountStatus 账号状态
type AccountStatus int8

const (
	AccountStatusBanned = -2 // 封禁 (不能访问任何api)
	AccountStatusLocked = -1 // 锁定 (不能登录)
	AccountStatusNormal = 0  // 正常 (必须有附带一个认证才能创建)
)

// Lock 锁定账号
func (a *Account) Lock() {
	a.Status = AccountStatusLocked
}

// Unlock 解锁账号
func (a *Account) Unlock() {
	a.Status = AccountStatusNormal
}

// Ban 封禁账号
func (a *Account) Ban() {
	a.Status = AccountStatusBanned
}

// CountAuthMethods 计算已有认证方式数量
func (a *Account) CountAuthMethods() int {
	count := 0
	if a.AuthPassword != nil {
		count++
	}
	if a.AuthPhone != nil {
		count++
	}
	if a.AuthEmail != nil {
		count++
	}
	count += len(a.AuthBiometrics)
	count += len(a.AuthThirdParties)
	return count
}

// InvalidateToken 使token失效
func (a *Account) InvalidateToken(ownType TokenOwn, ownID uint64) {
	switch ownType {
	case TokenOwnOrg:
		delete(a.OrgTokens, ownID)
		delete(a.OrgExpireAts, ownID)
	case TokenOwnApp:
		delete(a.AppTokens, ownID)
		delete(a.AppExpireAts, ownID)
	}
}

// GenerateToken 为账号生成token
func (a *Account) GenerateToken(
	ownType TokenOwn, ownID uint64, jwtSecret string,
	expireSec, refExpireHou int64, issuer string,
) (*Token, error) {
	// 创建新的Token
	token := NewToken(a.Id, ownType, ownID, expireSec, refExpireHou, issuer)
	// 生成JWT令牌
	if err := token.GenerateAccessJWTToken(jwtSecret); err != nil {
		return nil, err
	}
	// 生成刷新令牌
	if err := token.GenerateRefreshJWTToken(jwtSecret); err != nil {
		return nil, err
	}
	// 存储token
	if ownType == TokenOwnOrg {
		a.OrgTokens[ownID] = token.AccessToken
		a.OrgExpireAts[ownID] = time.Now().Add(time.Duration(token.ExpireSec) * time.Second).Unix()
		// 如果是首次激活，记录激活时间
		if _, exists := a.OrgActiveAts[ownID]; !exists {
			a.OrgActiveAts[ownID] = time.Now().Unix()
		}
	} else if ownType == TokenOwnApp {
		a.AppTokens[ownID] = token.AccessToken
		a.AppExpireAts[ownID] = time.Now().Add(time.Duration(token.ExpireSec) * time.Second).Unix()
		// 如果是首次激活，记录激活时间
		if _, exists := a.AppActiveAts[ownID]; !exists {
			a.AppActiveAts[ownID] = time.Now().Unix()
		}
	}
	return token, nil
}

// ValidateToken 验证token
func (a *Account) ValidateToken(
	tokenStr string, ownType TokenOwn, ownID uint64,
	jwtSecret string, checkExpire bool,
) (*TokenClaims, error) {
	// 验证token是否存在于账户中
	switch ownType {
	case TokenOwnOrg:
		storedToken, ok := a.OrgTokens[ownID]
		if !ok || (storedToken != tokenStr) {
			return nil, fmt.Errorf("token不存在或不匹配")
		}
		// 检查过期时间
		expireAt, ok := a.OrgExpireAts[ownID]
		if ok && (expireAt > 0) && (time.Now().Unix() > expireAt) {
			return nil, fmt.Errorf("token已过期")
		}
	case TokenOwnApp:
		storedToken, ok := a.AppTokens[ownID]
		if !ok || (storedToken != tokenStr) {
			return nil, fmt.Errorf("token不存在或不匹配")
		}
		// 检查过期时间
		expireAt, ok := a.AppExpireAts[ownID]
		if ok && (expireAt > 0) && (time.Now().Unix() > expireAt) {
			return nil, fmt.Errorf("token已过期")
		}
	default:
		return nil, fmt.Errorf("未知的客户端类型: %s", ownType)
	}
	// 解析和验证JWT
	return ParseJWT(tokenStr, jwtSecret, checkExpire)
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
	accExtraKeyAvatarId     = "avatarId"      // 头像ID
	accExtraKeyAvatarUrl    = "avatarUrl"     // 头像URL
	accExtraKeyStatusAt     = "statusAt"      // 状态时间
	accExtraKeyStatusReason = "statusReason"  // 状态原因
	accExtraKeyRoles        = "roles"         // 角色列表 (默认只有org下的用户有)
	accExtraBioCredential   = "bioCredential" // 生物特征凭据
	accExtraTPAccessToken   = "tpAccessToken" // 第三方认证访问令牌
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

func (a *Account) SetBioCredential(bioCredential *string) {
	a.Extra.SetString(accExtraBioCredential, bioCredential)
}

func (a *Account) GetBioCredential() (string, bool) {
	return a.Extra.GetString(accExtraBioCredential)
}

func (a *Account) SetTPAccessToken(tpAccessToken *string) {
	a.Extra.SetString(accExtraTPAccessToken, tpAccessToken)
}

func (a *Account) GetTPAccessToken() (string, bool) {
	return a.Extra.GetString(accExtraTPAccessToken)
}
