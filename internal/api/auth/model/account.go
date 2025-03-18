package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/valid"
	"reflect"
)

type (
	// Account 账号结构
	Account struct {
		*model.Base
		OwnKind OwnKind `json:"ownKind" validate:"required,range-own"` // 账号拥有者类型(注册的)
		OwnID   uint64  `json:"ownId" validate:"required"`             // 账号拥有者ID(注册的)

		UserID   *uint64 `json:"userId"`                              // 认证用户Id (有些org/app不填user，这里是第一绑定)
		Number   *uint64 `json:"number" validate:"format-number"`     // 账号标识 (自定义数字，防止暴露ID)
		Nickname *string `json:"nickname" validate:"format-nickname"` // 昵称 (没有user的app/org会用这个，放外面是方便搜索)

		Auths map[AuthKind]IAuth `json:"auths,omitempty"` // 认证方式列表 (多对多)

		LoginHistory  []any `json:"loginHistory,omitempty"`  // 登录历史(login)
		EntryHistory  []any `json:"entryHistory,omitempty"`  // 进入历史(entry)
		AccessHistory []any `json:"accessHistory,omitempty"` // 访问历史(api)
	}
)

func NewAccountEmpty() *Account {
	return &Account{
		Base:  model.NewBaseEmpty(),
		Auths: make(map[AuthKind]IAuth),
	}
}

func (a *Account) Wash() *Account {
	a.Base = a.Base.Wash(AccountStatusInit)
	a.UserID = nil
	a.Number = nil
	a.Auths = make(map[AuthKind]IAuth)
	a.LoginHistory = nil
	a.EntryHistory = nil
	a.AccessHistory = nil
	return a
}

func (a *Account) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
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
			// 账号标识
			"format-number": func(value reflect.Value, param string) bool {
				val := value.Interface().(*uint64)
				if val == nil {
					return true
				}
				return *val > 1_000_000
			},
			// 昵称
			"format-nickname": func(value reflect.Value, param string) bool {
				val := value.Interface().(*string)
				if val == nil {
					return true
				}
				return (len(*val) >= 3) && (len(*val) <= 50)
			},
		},
	}
}

func (a *Account) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"OwnKind": {"required_account_own_kind_err", false, nil},
					"OwnID":   {"required_account_own_id_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"range-own":       {"range_account_own_err", false, nil},
				"format-number":   {"format_account_number_err", false, nil},
				"format-nickname": {"format_account_nickname_err", false, nil},
			},
		},
	}
}

const (
	AccountStatusBanned     model.Status = -4 // 封禁 (不能获取到，不能访问任何api，包括注销/注册)
	AccountStatusUnRegister model.Status = -3 // 注销 (不能获取到，可以重新注册)
	AccountStatusBlocked    model.Status = -2 // 锁定 (能获取到，暂时锁定，管理员解锁)
	AccountStatusLocked     model.Status = -1 // 锁定 (能获取到，暂时锁定，登录解锁)
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
	return a.Status > AccountStatusBlocked
}

// IsNeedAuth 是否需要认证
func (a *Account) IsNeedAuth() bool {
	return a.Status == AccountStatusInit
}

// CanAccess 可否访问
func (a *Account) CanAccess() bool {
	return a.Status >= AccountStatusActive
}

// AddAuth 添加认证方式
func (a *Account) AddAuth(auth IAuth) bool {
	if a.Auths == nil {
		a.Auths = make(map[AuthKind]IAuth)
	}

	a.Auths[auth.GetKind()] = auth
	auth.SetAccount(a)

	if (a.Status >= AccountStatusInit) && (a.Status < AccountStatusActive) {
		a.Status = AccountStatusActive
		return true
	}
	return false
}

// DelAuth 删除认证方式
func (a *Account) DelAuth(auth IAuth) bool {
	if a.Auths == nil {
		return a.Status > AccountStatusInit
	}

	if _, ok := a.Auths[auth.GetKind()]; ok {
		auth.DelAccount(a.OwnKind, a.OwnID)
		delete(a.Auths, auth.GetKind())
	}

	if (len(a.Auths) == 0) && (a.Status >= AccountStatusActive) {
		a.Status = AccountStatusInit
		return true
	}
	return false
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

// HasAuthKind 检查是否拥有指定的认证方式
func (a *Account) HasAuthKind(authKind AuthKind) bool {
	if a.Auths == nil {
		return false
	}
	for kind := range a.Auths {
		if kind == authKind {
			return true
		}
	}
	return false
}

// FirstAuth 获取第一个认证方式
func (a *Account) FirstAuth() IAuth {
	if a.Auths == nil {
		return nil
	}
	for _, au := range a.Auths {
		if au != nil {
			return au
		}
	}
	return nil
}

// AddRole 添加角色 TODO:GG 在这里?
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

// HasRole 检查是否拥有指定角色 TODO:GG 在这里?
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
	accExtraKeyAvatarID  = "avatarId"  // 头像ID
	accExtraKeyAvatarUrl = "avatarUrl" // 头像URL
	accExtraKeyRoles     = "roles"     // 角色列表 (默认只有org下的用户有) TODO:GG 放在extra？还是这里外键关联？还是不放？
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

func (a *Account) SetRoles(roles *[]string) {
	a.Extra.SetStringSlice(accExtraKeyRoles, roles)
}

func (a *Account) GetRoles() ([]string, bool) {
	return a.Extra.GetStringSlice(accExtraKeyRoles)
}
