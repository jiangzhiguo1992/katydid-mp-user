package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/utils"
	"time"
)

type AuthKind int16

const (
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

//var _ IAuth = (*UsernamePwd)(nil)
//var _ IAuth = (*PhoneNumber)(nil)
//var _ IAuth = (*EmailAddress)(nil)
//var _ IAuth = (*Biometric)(nil)
//var _ IAuth = (*ThirdPartyLink)(nil)

type (
	IAuth interface {
		// GetID 获取认证记录的ID
		GetID() uint64

		// GetAccountID 获取关联的账号ID
		GetAccountID() uint64

		// IsEnabled 检查认证方式是否启用
		IsEnabled() bool

		// IsActive 检查认证方式是否已激活过
		IsActive() bool

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

		// TODO:GG verifys 关联
		//AuthOnAts  map[int][]int64 `json:"authOnAts"`  // [kind][]验证发送时间
		//AuthBodies map[int]*any    `json:"authBodies"` // [kind]验证内容
		//AuthOkAts  map[int][]int64 `json:"authOkAts"`  // [kind][]验证成功时间
	}

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

func NewAuthDef(accountId uint64) *Auth {
	return &Auth{
		Base:      model.NewBaseEmpty(),
		AccountId: accountId,
		Enable:    true,
		IsActive:  false,
	}
}

func NewUsernamePwdDef(accountId uint64, username, pwd string) *UsernamePwd {
	return &UsernamePwd{
		Auth:     NewAuthDef(accountId),
		Username: username,
		Password: pwd,
	}
}

func NewPhoneNumberDef(accountId uint64, areaCode string, number string) *PhoneNumber {
	return &PhoneNumber{
		Auth:     NewAuthDef(accountId),
		AreaCode: areaCode,
		Number:   number,
		//Operator: operator,
	}
}

func (a *Auth) IsEnabled() bool {
	return a.Status >= model.StatusInit
}

func (a *Auth) IsActive() bool {
	return a.Status >= model.StatusActive
}

const (
	authExtraKeyPhoneCode = "phoneCode" // 手机验证码
)

func (a *Auth) AddPhoneCode(timestamp int64, code *string) {
	datas, ok := a.Extra.Get(authExtraKeyPhoneCode)
	if !ok || datas == nil {
		datas = make(map[int64]utils.KSMap)
	} else if _, ok := datas.(map[int64]utils.KSMap); !ok {
		datas = make(map[int64]utils.KSMap)
	}
	codes := datas.(map[int64]utils.KSMap)
	codes[timestamp] = make(utils.KSMap)
	codes[timestamp].SetString("code", code)
	a.Extra.Set(authExtraKeyPhoneCode, codes)
}

func (a *Auth) SetPhoneCodes(codes *map[int64]utils.KSMap) {
	a.Extra.SetPtr(authExtraKeyPhoneCode, codes)
}

func (a *Auth) GetPhoneCodes() (map[int64]utils.KSMap, bool) {
	datas, ok := a.Extra.Get(authExtraKeyPhoneCode)
	if ok && datas != nil {
		if _, ok := datas.(map[int64]utils.KSMap); ok {
			return datas.(map[int64]utils.KSMap), true
		}
	}
	return nil, ok
}
