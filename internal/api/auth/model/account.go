package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/data"
)

type (
	// Account 账号结构
	Account struct {
		*model.Base
		OwnKind  OwnKind `json:"ownKind"`  // 账号拥有者类型(注册的)
		OwnID    uint64  `json:"ownId"`    // 账号拥有者ID(注册的)
		Nickname string  `json:"nickname"` // 昵称 (没有user的app/org会用这个，放外面是方便搜索)

		Auths map[AuthKind]IAuth `json:"auths,omitempty"` // 认证方式列表 (多对多)

		//LoginHistory  []*Entry `json:"loginHistory"`  // 登录历史(login)
		//EntryHistory  []any    `json:"entryHistory"`  // 进入历史(entry)
		//AccessHistory []any    `json:"accessHistory"` // 访问历史(api)
	}
)

func NewAccountEmpty() *Account {
	return &Account{
		Base:  model.NewBaseEmpty(),
		Auths: make(map[AuthKind]IAuth),
	}
}

func NewAccount(
	ownKind OwnKind, ownID uint64, nickname string,
) *Account {
	base := model.NewBase(make(data.KSMap))
	base.Status = AccountStatusInit
	return &Account{
		Base:    base,
		OwnKind: ownKind, OwnID: ownID, Nickname: nickname,
		Auths: make(map[AuthKind]IAuth),
	}
}

const (
	AccountStatusBanned     model.Status = -3 // 封禁 (不能获取到，不能访问任何api，包括注销/注册)
	AccountStatusUnRegister model.Status = -2 // 注销 (不能获取到，可以重新注册)
	AccountStatusLocked     model.Status = -1 // 锁定 (能获取到，暂时锁定，不能登录)
	AccountStatusInit       model.Status = 0  // 初始 (未激活状态)
	AccountStatusActive     model.Status = 1  // 激活 (能访问api，必须有附带requires认证才能创建)
)

// CanRegister 可否注册
func (a *Account) CanRegister() bool {
	return a.Status > AccountStatusBanned
}

// IsUnRegister 是否注销
func (a *Account) IsUnRegister() bool {
	return a.Status <= AccountStatusUnRegister
}

// CanLogin 可否登录
func (a *Account) CanLogin() bool {
	return a.Status > AccountStatusLocked
}

// IsNeedAuth 是否需要认证
func (a *Account) IsNeedAuth() bool {
	return a.Status == AccountStatusInit
}

// CanAccess 可否访问
func (a *Account) CanAccess() bool {
	return a.Status >= AccountStatusActive
}

func (a *Account) AddAuth(auth IAuth) {
	if a.Auths == nil {
		a.Auths = make(map[AuthKind]IAuth)
	}
	a.Auths[auth.GetKind()] = auth
	if (a.Status >= AccountStatusInit) && (a.Status < AccountStatusActive) {
		a.Status = AccountStatusActive
	}
	a.Auths[auth.GetKind()].SetAccount(a)
}

func (a *Account) DelAuth(auth IAuth) {
	if a.Auths == nil {
		return
	}
	if _, ok := a.Auths[auth.GetKind()]; ok {
		a.Auths[auth.GetKind()].DelAccount(a)
		delete(a.Auths, auth.GetKind())
	}
	if (len(a.Auths) == 0) && (a.Status >= AccountStatusActive) {
		a.Status = AccountStatusInit
	}
	a.Auths[auth.GetKind()].DelAccount(a)
}

// GetAuthKinds 获取认证方式种类列表
func (a *Account) GetAuthKinds() []AuthKind {
	kinds := make([]AuthKind, 0)
	for kind, au := range a.Auths {
		if au != nil {
			kinds = append(kinds, kind)
			break
		}
	}
	return kinds
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
	accExtraKeyAvatarID     = "avatarId"     // 头像ID
	accExtraKeyAvatarUrl    = "avatarUrl"    // 头像URL
	accExtraKeyStatusAt     = "statusAt"     // 状态时间
	accExtraKeyStatusReason = "statusReason" // 状态原因
	accExtraKeyRoles        = "roles"        // 角色列表 (默认只有org下的用户有) TODO:GG 放在extra？还是这里外键关联？还是不放？
)

func (a *Account) SetAvatarID(avatarId *int64) {
	a.Extra.SetInt64(accExtraKeyAvatarID, avatarId)
}

func (a *Account) GetAvatarID() (int64, bool) {
	return a.Extra.GetInt64(accExtraKeyAvatarID)
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
