package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/data"
)

var _ IAuth = (*AuthPassword)(nil)
var _ IAuth = (*AuthCellphone)(nil)
var _ IAuth = (*AuthEmail)(nil)

type (
	IAuth interface {
		IsBlocked() bool   // 检查认证方式是否被封禁
		IsEnabled() bool   // 检查认证方式是否启用
		IsActive() bool    // 检查认证方式是否已激活过
		GetKind() AuthKind // 获取认证类型

		SetAccount(*Account)                             // 关联账号信息
		DelAccount(*Account)                             // 删除关联账号信息
		GetAccAccounts() map[OwnKind]map[uint64]*Account // 获取关联的账号ID
		GetAccount(OwnKind, uint64) *Account             // 获取关联的账号ID

		GetVerify(VerifyApply) *Verify // 获取认证信息
	}

	// Auth 可验证账号基础
	Auth struct {
		*model.Base
		Kind AuthKind `json:"kind"` // 认证类型

		// implements

		Accounts map[OwnKind]map[uint64]*Account `json:"accountId"`          // 账户Id (多对多)
		Verifies map[VerifyApply]*Verify         `json:"verifies,omitempty"` // 认证信息
	}

	// AuthPassword 用户名+密码
	AuthPassword struct {
		*Auth
		Username *string `json:"username"` // 用户名(可能为空)

		Password string `json:"omitempty" gorm:"column:password_hash"` // 密码 (md5)
	}

	// AuthCellphone 移动手机号+短信
	AuthCellphone struct {
		*Auth
		Code   string `json:"code"`   // 国家区号
		Number string `json:"number"` // 手机号

		Operator *string `json:"operator"` // 运营商
	}

	// AuthEmail 邮箱+邮件
	AuthEmail struct {
		*Auth
		Username string `json:"username"` // 用户名
		Domain   string `json:"domain"`   // 域名

		Entity *string `json:"entity"` // 邮箱服务商 (eg:163/qq/...)
		TLD    *string `json:"tld"`    // 顶级域名 (eg:com/cn/org/...)
	}

	// OwnKind 认证拥有者类型
	OwnKind int16

	// AuthKind 认证类型
	AuthKind int16
)

func NewAuthEmpty() *Auth {
	return &Auth{
		Base: model.NewBaseEmpty(),
	}
}

// NewAuth 创建指定类型的认证实例
func NewAuth(kind AuthKind) *Auth {
	base := model.NewBase(make(data.KSMap))
	base.Status = AuthStatusInit
	return &Auth{
		Base: base,
		Kind: kind,
	}
}

func NewAuthPasswordEmpty() *AuthPassword {
	return &AuthPassword{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthPassword(
	username *string, password string,
) *AuthPassword {
	return &AuthPassword{
		Auth:     NewAuth(AuthKindPassword),
		Username: username, Password: password,
	}
}

func NewAuthPhoneEmpty() *AuthCellphone {
	return &AuthCellphone{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthPhone(
	code, number string,
) *AuthCellphone {
	return &AuthCellphone{
		Auth: NewAuth(AuthKindCellphone),
		Code: code, Number: number,
		Operator: nil,
	}
}

func NewAuthEmailEmpty() *AuthEmail {
	return &AuthEmail{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthEmail(
	username, domain string,
) *AuthEmail {
	return &AuthEmail{
		Auth:     NewAuth(AuthKindEmail),
		Username: username, Domain: domain,
		Entity: nil, TLD: nil,
	}
}

const (
	AuthStatusBlock  model.Status = -1 // 封禁状态
	AuthStatusInit   model.Status = 0  // 初始状态
	AuthStatusActive model.Status = 1  // 激活状态

	OwnKindOrg    OwnKind = 10 // 组织
	OwnKindRole   OwnKind = 11 // 角色
	OwnKindApp    OwnKind = 20 // 应用
	OwnKindClient OwnKind = 21 // 客户端
	OwnKindUser   OwnKind = 30 // 用户

	AuthKindPassword    AuthKind = 10  // 密码
	AuthKindCellphone   AuthKind = 20  // 短信
	AuthKindEmail       AuthKind = 30  // 邮箱
	AuthKindBioFace     AuthKind = 40  // 生物特征-人脸
	AuthKindBioFinger   AuthKind = 41  // 生物特征-指纹
	AuthKindBioVoice    AuthKind = 42  // 生物特征-声纹
	AuthKindBioIris     AuthKind = 43  // 生物特征-虹膜
	AuthKindThirdGoogle AuthKind = 100 // 三方平台-Google
	AuthKindThirdApple  AuthKind = 101 // 三方平台-Apple
	AuthKindThirdWechat AuthKind = 102 // 三方平台-微信
	AuthKindThirdQQ     AuthKind = 103 // 三方平台-QQ
	AuthKindThirdIns    AuthKind = 104 // 三方平台-Instagram
	AuthKindThirdFB     AuthKind = 105 // 三方平台-Facebook
)

func (a *Auth) IsBlocked() bool {
	return a.Status <= AuthStatusBlock
}

func (a *Auth) IsEnabled() bool {
	return a.Status >= AuthStatusInit
}

func (a *Auth) IsActive() bool {
	return a.Status >= AuthStatusActive
}

func (a *Auth) GetKind() AuthKind {
	return a.Kind
}

func (a *Auth) SetAccount(account *Account) {
	if _, ok := a.Accounts[account.OwnKind]; !ok {
		a.Accounts[account.OwnKind] = make(map[uint64]*Account)
	}
	a.Accounts[account.OwnKind][account.ID] = account
}

func (a *Auth) DelAccount(account *Account) {
	if owns, ok := a.Accounts[account.OwnKind]; ok {
		delete(owns, account.ID)
		if len(owns) == 0 {
			delete(a.Accounts, account.OwnKind)
		}
	}
}

func (a *Auth) GetAccAccounts() map[OwnKind]map[uint64]*Account {
	return a.Accounts
}

func (a *Auth) GetAccount(ownKind OwnKind, ownID uint64) *Account {
	if owns, ok := a.Accounts[ownKind]; ok {
		if account, ok := owns[ownID]; ok {
			return account
		}
	}
	return nil
}

func (a *Auth) GetVerify(apply VerifyApply) *Verify {
	if verify, ok := a.Verifies[apply]; ok {
		return verify
	}
	return nil
}

func (a *AuthCellphone) FullNumber() string {
	return "+" + a.Code + " " + a.Number
}

func (a *AuthEmail) EmailAddress() string {
	return a.Username + "@" + a.Domain
}

const (
	authExtraKeyPasswordSalt = "passwordSalt" // 密码盐 TODO:GG 不response
)

func (a *Auth) SetPasswordSalt(salt *string) {
	a.Extra.SetString(authExtraKeyPasswordSalt, salt)
}

func (a *Auth) GetPasswordSalt() (string, bool) {
	return a.Extra.GetString(authExtraKeyPasswordSalt)
}
