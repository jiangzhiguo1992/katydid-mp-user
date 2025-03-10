package model

import (
	"katydid-mp-user/internal/pkg/model"
)

//var _ IAuth = (*UsernamePwd)(nil)
//var _ IAuth = (*PhoneNumber)(nil)
//var _ IAuth = (*EmailAddress)(nil)
//var _ IAuth = (*Biometric)(nil)
//var _ IAuth = (*ThirdPartyLink)(nil)

type (
	IAuth interface {
		// IsEnabled 检查认证方式是否启用
		IsEnabled() bool

		// IsActive 检查认证方式是否已激活过
		IsActive() bool

		// GetAccountID 获取关联的账号ID
		GetAccountID() uint64

		// GetKind 获取认证类型
		GetKind() AuthKind

		// Verify 验证认证信息
		Verify(credential interface{}) (bool, error)
	}

	// Auth 可验证账号基础
	Auth struct {
		*model.Base
		AccountID uint64 `json:"accountId"` // 账户Id

		Kind AuthKind `json:"kind"` // 认证类型
	}

	AuthKind int16

	// AuthPassword 用户名+密码
	AuthPassword struct {
		Auth
		Username string `json:"Username"` // 用户名

		Salt     string `json:"-"`                             // 密码盐
		Password string `json:"-" gorm:"column:password_hash"` // 密码 (md5)
	}

	// AuthPhone 手机号+短信
	AuthPhone struct {
		*Auth
		AreaCode string `json:"areaCode"` // 手机区号
		Number   string `json:"number"`   // 手机号

		Operator string `json:"operator"` // 运营商
	}

	// AuthEmail 邮箱+邮件
	AuthEmail struct {
		*Auth
		Username string `json:"username"` // 用户名
		Domain   string `json:"domain"`   // 域名
	}

	// Biometric 生物特征
	Biometric struct {
		*Auth
		Kind int16 `json:"kind"` // 平台 (eg:人脸/指纹/声纹/...)

		// extra => credential
	}

	// ThirdPartyLink 第三方账号链接
	ThirdPartyLink struct {
		*Auth
		Kind   int16  `json:"kind"`   // 平台 (eg:微信/QQ/...)
		OpenId string `json:"openId"` // 第三方平台唯一标识

		// extra => accessToken, linkId, credential
	}
)

//func NewAuth(accountId uint64) *Auth {
//	return &Auth{
//		Base:      model.NewBaseEmpty(),
//		AccountId: accountId,
//		Enable:    true,
//		IsActive:  false,
//	}
//}
//
//func NewUsernamePwdDef(accountId uint64, username, pwd string) *UsernamePwd {
//	return &UsernamePwd{
//		Auth:     NewAuthDef(accountId),
//		Username: username,
//		Password: pwd,
//	}
//}
//
//func NewPhoneNumberDef(accountId uint64, areaCode string, number string) *PhoneNumber {
//	return &PhoneNumber{
//		Auth:     NewAuthDef(accountId),
//		AreaCode: areaCode,
//		Number:   number,
//		//Operator: operator,
//	}
//}

const (
	AuthStatusBlock  model.Status = -1 // 封禁状态
	AuthStatusInit   model.Status = 0  // 初始状态
	AuthStatusActive model.Status = 1  // 激活状态

	AuthKindUserPwd    AuthKind = 1 // 密码
	AuthKindPhoneCode  AuthKind = 2 // 短信
	AuthKindEmailCode  AuthKind = 3 // 邮箱
	AuthKindBiometric  AuthKind = 4 // 生物特征
	AuthKindThirdParty AuthKind = 5 // 第三方平台

	//AuthBioKindFace   int16 = 1 // 人脸
	//AuthBioKindFinger int16 = 2 // 指纹
	//AuthBioKindVoice  int16 = 3 // 声纹
	//
	//AuthThirdKindWechat int16 = 1 // 微信
	//AuthThirdKindQQ     int16 = 2 // QQ

)

func (a *Auth) IsEnabled() bool {
	return a.Status >= AuthStatusInit
}

func (a *Auth) IsActive() bool {
	return a.Status >= AuthStatusActive
}

func (a *Auth) GetAccountID() uint64 {
	return a.AccountID
}

func (a *Auth) GetKind() AuthKind {
	return a.Kind
}

func (a *Auth) Verify(credential interface{}) (bool, error) {
	// TODO:GG
	return false, nil
}

const (
	authExtraKeyBlockedUntil = "blockedUntil" // 阻止验证直到某个时间点(过于频繁的send)
)

func (a *Auth) SetBlockedUntil(blockUtilUnix *int64) {
	a.Extra.SetInt64(authExtraKeyBlockedUntil, blockUtilUnix)
}

func (a *Auth) GetBlockedUntil() (int64, bool) {
	return a.Extra.GetInt64(authExtraKeyBlockedUntil)
}
