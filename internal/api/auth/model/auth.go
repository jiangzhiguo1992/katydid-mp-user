package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/data"
	"katydid-mp-user/pkg/valid"
	"reflect"
	"unicode"
)

var _ IAuth = (*AuthPassword)(nil)
var _ IAuth = (*AuthCellphone)(nil)
var _ IAuth = (*AuthEmail)(nil)

type (
	// IAuth 认证接口
	IAuth interface {
		IsBlocked() bool    // 检查认证方式是否被封禁
		IsEnabled() bool    // 检查认证方式是否启用
		IsActive() bool     // 检查认证方式是否已激活过
		GetKind() AuthKind  // 获取认证类型
		GetUserID() *uint64 // 认证用户Id

		SetAccount(*Account)                             // 关联账号信息
		DelAccount(*Account)                             // 删除关联账号信息
		GetAccAccounts() map[OwnKind]map[uint64]*Account // 获取关联的账号ID
		GetAccount(OwnKind, uint64) *Account             // 获取关联的账号ID

		GetVerify(VerifyApply) *Verify // 获取认证信息
	}

	// Auth 可验证账号基础
	Auth struct {
		*model.Base
		Kind   AuthKind `json:"kind" validate:"required,check-kind"` // 认证类型
		UserID *uint64  `json:"userId"`                              // 认证用户Id (有些org/app不填user，这里是第一绑定)

		// implements

		Accounts map[OwnKind]map[uint64]*Account `json:"-"` // 账户Id (多对多表)
		Verifies map[VerifyApply]*Verify         `json:"-"` // 认证信息
	}

	// AuthPassword 用户名+密码
	AuthPassword struct {
		*Auth
		Username *string `json:"username" validate:"required,format-username"` // 用户名(可能为空)

		PasswordMD5 string `json:"omitempty" validate:"required"` // 密码
	}

	// AuthCellphone 移动手机号+短信
	AuthCellphone struct {
		*Auth
		Code   string `json:"code" validate:"required,range-code"` // 国家区号
		Number string `json:"number" validate:"required"`          // 手机号

		Operator *string `json:"operator"` // 运营商
	}

	// AuthEmail 邮箱+邮件
	AuthEmail struct {
		*Auth
		Username string `json:"username" validate:"required,format-username"` // 用户名
		Domain   string `json:"domain" validate:"required,format-domain"`     // 域名

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

func NewAuthPasswordEmpty() *AuthPassword {
	return &AuthPassword{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthCellphoneEmpty() *AuthCellphone {
	return &AuthCellphone{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthEmailEmpty() *AuthEmail {
	return &AuthEmail{
		Auth: NewAuthEmpty(),
	}
}

func NewAuth(kind AuthKind) *Auth {
	base := model.NewBase(make(data.KSMap))
	base.Status = model.StatusInit
	return &Auth{
		Base: base,
		Kind: kind,
	}
}

func NewAuthPassword(
	username *string, passwordMD5 string,
) *AuthPassword {
	return &AuthPassword{
		Auth:     NewAuth(AuthKindPassword),
		Username: username, PasswordMD5: passwordMD5,
	}
}

func NewAuthCellphone(
	code, number string,
) *AuthCellphone {
	return &AuthCellphone{
		Auth: NewAuth(AuthKindCellphone),
		Code: code, Number: number,
		Operator: nil,
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

func (a *AuthPassword) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 认证类型
			"check-kind": func(value reflect.Value, param string) bool {
				val := value.Interface().(AuthKind)
				return val == AuthKindPassword
			},
			// 用户名
			"format-username": func(value reflect.Value, param string) bool {
				val := value.Interface().(string)
				// 长度检查：3-30个字符
				if len(val) < 3 || len(val) > 30 {
					return false
				}

				// 必须以字母开头
				if !unicode.IsLetter(rune(val[0])) {
					return false
				}

				// 不能以连字符或下划线结尾
				lastChar := val[len(val)-1]
				if lastChar == '-' || lastChar == '_' {
					return false
				}

				// 不能有连续的连字符或下划线
				hasContinuousSpecial := false
				lastIsSpecial := false

				for i, char := range val {
					if i == 0 {
						continue // 已经检查过第一个字符
					}
					// 只允许字母、数字、下划线和连字符
					if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
						return false
					}
					// 检查连续的特殊字符
					isSpecial := char == '_' || char == '-'
					if isSpecial && lastIsSpecial {
						hasContinuousSpecial = true
						break
					}
					lastIsSpecial = isSpecial
				}
				return !hasContinuousSpecial

			},
		},
	}
}

func (a *AuthCellphone) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 认证类型
			"check-kind": func(value reflect.Value, param string) bool {
				val := value.Interface().(AuthKind)
				return val == AuthKindCellphone
			},
			// 手机区号
			"range-code": func(value reflect.Value, param string) bool {
				val := value.Interface().(string)
				_, ok := valid.IsPhoneCountryCode(val)
				return ok
			},
		},
	}
}

func (a *AuthEmail) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 认证类型
			"kind-check": func(value reflect.Value, param string) bool {
				val := value.Interface().(AuthKind)
				return val == AuthKindEmail
			},
			// 邮箱用户名
			"format-username": func(value reflect.Value, param string) bool {
				val := value.Interface().(string)
				return valid.IsEmailUsername(val)
			},
			// 邮箱域名
			"format-domain": func(value reflect.Value, param string) bool {
				val := value.Interface().(string)
				return valid.IsEmailDomain(val)
			},
		},
	}
}

func (a *AuthPassword) ValidExtraRules() (data.KSMap, valid.ExtraValidRules) {
	return a.Extra, valid.ExtraValidRules{
		valid.SceneSave: map[valid.Tag]valid.ExtraValidRuleInfo{
			// 密码盐
			authExtraKeyPasswordSalt: {
				Field: authExtraKeyPasswordSalt,
				ValidFn: func(value any) bool {
					val, ok := value.(string)
					if !ok {
						return false
					}
					return len(val) <= 0
				},
			},
		},
	}
}

func (a *AuthCellphone) ValidStructRules(scene valid.Scene, fn valid.FuncReportError) {
	switch scene {
	case valid.SceneAll:
		_, _, ok := valid.IsPhoneNumber(a.Code, a.Number)
		if !ok {
			// 手机区号+号码
			fn(a.Number, "Number", valid.TagFormat, "")
		}
	}
}

func (a *AuthPassword) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"AuthKind":    {"required_auth_kind_err", false, nil},
					"Username":    {"required_auth_username_err", false, nil},
					"PasswordMD5": {"required_auth_password_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"check-kind":             {"check_auth_kind_err", false, nil},
				"format-username":        {"format_auth_username_err", false, nil},
				authExtraKeyPasswordSalt: {"require_auth_password_salt_err", false, nil},
			},
		},
	}
}

func (a *AuthCellphone) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"AuthKind": {"required_auth_kind_err", false, nil},
					"Code":     {"required_auth_cellphone_code_err", false, nil},
					"Number":   {"required_auth_cellphone_number_err", false, nil},
				},
				valid.TagFormat: {
					"Number": {"format_auth_cellphone_number_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"check-kind": {"check_auth_kind_err", false, nil},
				"range-code": {"range_auth_cellphone_code_err", false, nil},
			},
		},
	}
}

func (a *AuthEmail) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"AuthKind": {"required_auth_kind_err", false, nil},
					"Username": {"required_auth_email_username_err", false, nil},
					"Domain":   {"required_auth_email_domain_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"check-kind":      {"check_auth_kind_err", false, nil},
				"format-username": {"format_auth_email_username_err", false, nil},
				"format-domain":   {"format_auth_email_domain_err", false, nil},
			},
		},
	}
}

const (
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
	return a.IsBlack()
}

func (a *Auth) IsEnabled() bool {
	return a.Status >= model.StatusInit
}

func (a *Auth) IsActive() bool {
	return a.IsWhite()
}

func (a *Auth) GetKind() AuthKind {
	return a.Kind
}

func (a *Auth) GetUserID() *uint64 {
	return a.UserID
}

func (a *Auth) SetAccount(account *Account) {
	if _, ok := a.Accounts[account.OwnKind]; !ok {
		a.Accounts[account.OwnKind] = make(map[uint64]*Account)
	}
	a.Accounts[account.OwnKind][account.ID] = account
	if (a.Status >= model.StatusInit) && (a.Status < model.StatusWhite) {
		a.Status = model.StatusWhite
	}
	if account.Auths[a.Kind] == nil {
		account.AddAuth(a)
	}
}

func (a *Auth) DelAccount(account *Account) {
	if owns, ok := a.Accounts[account.OwnKind]; ok {
		delete(owns, account.ID)
		if len(owns) == 0 {
			delete(a.Accounts, account.OwnKind)
		}
	}
	if (a.Status >= model.StatusWhite) && (len(a.Accounts) == 0) {
		a.Status = model.StatusInit
	}
	if account.Auths[a.Kind] != nil {
		account.DelAuth(a)
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

// FullNumber 返回完整的手机号
func (a *AuthCellphone) FullNumber() string {
	return "+" + a.Code + " " + a.Number
}

// EmailAddress 返回邮箱地址
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
