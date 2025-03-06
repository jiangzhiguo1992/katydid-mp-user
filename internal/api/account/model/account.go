package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/utils"
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

// RandomToken 生成随机token的真实实现
func RandomToken() string {
	// 此处应实现真正的随机token生成逻辑
	// 例如生成UUID、安全随机字符串等
	return "真实随机Token"
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
