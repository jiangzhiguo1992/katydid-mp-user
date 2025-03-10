package model

import (
	"errors"
	"katydid-mp-user/internal/pkg/model"
)

var _ IAuth = (*AuthPassword)(nil)
var _ IAuth = (*AuthPhone)(nil)
var _ IAuth = (*AuthEmail)(nil)
var _ IAuth = (*AuthBiometric)(nil)
var _ IAuth = (*AuthThirdParty)(nil)

type (
	IAuth interface {
		IsEnabled() bool                             // 检查认证方式是否启用
		IsActive() bool                              // 检查认证方式是否已激活过
		GetAccountID() uint64                        // 获取关联的账号ID
		GetKind() AuthKind                           // 获取认证类型
		Verify(credential interface{}) (bool, error) // 验证认证信息
	}

	// Auth 可验证账号基础
	Auth struct {
		*model.Base
		AccountID uint64 `json:"accountId"` // 账户Id

		Kind AuthKind `json:"kind"` // 认证类型
	}

	// AuthKind 认证类型
	AuthKind int16

	// AuthPassword 用户名+密码
	AuthPassword struct {
		*Auth

		Username string `json:"Username"`                      // 用户名
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

	// AuthBiometric 生物特征
	AuthBiometric struct {
		*Auth

		Kind AuthBioKind `json:"kind"` // 生物特征类型 (eg:人脸/指纹/声纹/...)
		// extra => credential
	}

	// AuthThirdParty 第三方账号链接
	AuthThirdParty struct {
		*Auth

		Provider AuthTPProvider `json:"provider" gorm:"index"` // 平台 (eg:微信/QQ/...)
		OpenId   string         `json:"openId"`                // 第三方平台唯一标识
		// extra => AccessToken,RefreshToken, TokenExpiredAt, linkId, credential, UserInfo
	}

	// AuthBioKind 生物特征类型
	AuthBioKind int16

	// AuthTPProvider 第三方平台类型
	AuthTPProvider int16
)

// NewAuth 创建指定类型的认证实例
func NewAuth(kind AuthKind, accountID uint64) IAuth {
	auth := &Auth{
		AccountID: accountID,
		Kind:      kind,
	}
	switch kind {
	case AuthKindPassword:
		return &AuthPassword{Auth: auth}
	case AuthKindPhone:
		return &AuthPhone{Auth: auth}
	case AuthKindEmail:
		return &AuthEmail{Auth: auth}
	case AuthKindBiometric:
		return &AuthBiometric{Auth: auth}
	case AuthKindThirdParty:
		return &AuthThirdParty{Auth: auth}
	default:
		return nil
	}
}

const (
	AuthStatusBlock  model.Status = -1 // 封禁状态
	AuthStatusInit   model.Status = 0  // 初始状态
	AuthStatusActive model.Status = 1  // 激活状态

	AuthKindPassword   AuthKind = 1 // 密码
	AuthKindPhone      AuthKind = 2 // 短信
	AuthKindEmail      AuthKind = 3 // 邮箱
	AuthKindBiometric  AuthKind = 4 // 生物特征
	AuthKindThirdParty AuthKind = 5 // 第三方平台

	AuthBioKindFace   AuthBioKind = 1 // 人脸
	AuthBioKindFinger AuthBioKind = 2 // 指纹
	AuthBioKindVoice  AuthBioKind = 3 // 声纹

	AuthThirdKindWechat AuthTPProvider = 1 // 微信
	AuthThirdKindQQ     AuthTPProvider = 2 // QQ
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

func (a *Auth) Verify(_ interface{}) (bool, error) {
	return false, errors.New("verify not implemented")
}

// Verify 实现密码验证
func (p *AuthPassword) Verify(credential interface{}) (bool, error) {
	cred, ok := credential.(struct {
		Username string
		Password string
	})

	if !ok {
		return false, errors.New("invalid credential format")
	}

	// TODO:GG 实际场景中应该使用安全的密码哈希比较
	if p.Username == cred.Username && p.Password == cred.Password {
		return true, nil
	}
	return false, nil
}

// Verify 实现手机号验证
func (p *AuthPhone) Verify(credential interface{}) (bool, error) {
	cred, ok := credential.(struct {
		AreaCode string
		Number   string
	})

	if !ok {
		return false, errors.New("invalid credential format")
	}
	if p.AreaCode == cred.AreaCode && p.Number == cred.Number {
		return true, nil
	}
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
