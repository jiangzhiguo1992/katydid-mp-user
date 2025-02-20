package model

import (
	"katydid-mp-user/internal/pkg/model"
)

const (
	AuthKindPwd   int16 = 1 // 密码
	AuthKindPhone int16 = 2 // 短信
	AuthKindEmail int16 = 3 // 邮箱
	AuthKindBio   int16 = 4 // 生物特征
	AuthKindThird int16 = 5 // 第三方平台

	AuthBioKindFace   int16 = 1 // 人脸
	AuthBioKindFinger int16 = 2 // 指纹
	AuthBioKindVoice  int16 = 3 // 声纹

	AuthThirdKindWechat int16 = 1 // 微信
	AuthThirdKindQQ     int16 = 2 // QQ
)

type (
	// Auth 可验证账号基础
	Auth struct {
		*model.Base
		AccountId uint64 `json:"accountId"` // 账户Id

		Enable   bool `json:"enable"`   // 是否可用 (类似拉黑)
		IsActive bool `json:"isActive"` // 是否激活
	}

	// UsernamePwd 用户名+密码
	UsernamePwd struct {
		*Auth
		UserName string `json:"userName"` // 用户名

		Password string `json:"password"` // 密码 (md5)
	}

	// PhoneNumber 手机号+短信
	PhoneNumber struct {
		*Auth
		AreaCode int    `json:"areaCode"` // 手机区号
		Number   string `json:"number"`   // 手机号

		Operator string `json:"operator"` // 运营商
	}

	// EmailAddress 邮箱+邮件
	EmailAddress struct {
		*Auth
		Address string `json:"address"` // 用户名
		Domain  string `json:"domain"`  // 域名
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

func NewAuthEmpty(accountId uint64) *Auth {
	return &Auth{
		Base:      model.NewBaseEmpty(),
		AccountId: accountId,
		Enable:    true,
		IsActive:  false,
	}
}

func NewUsernamePwdEmpty(accountId uint64, username, pwd string) *UsernamePwd {
	return &UsernamePwd{
		Auth:     NewAuthEmpty(accountId),
		UserName: username,
		Password: pwd,
	}
}

func NewPhoneNumberEmpty(accountId uint64, areaCode int, number string) *PhoneNumber {
	return &PhoneNumber{
		Auth:     NewAuthEmpty(accountId),
		AreaCode: areaCode,
		Number:   number,
		//Operator: operator,
	}
}
