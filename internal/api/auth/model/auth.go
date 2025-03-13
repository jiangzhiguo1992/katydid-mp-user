package model

import (
	"errors"
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/data"
	"time"
)

var _ IAuth = (*AuthPassword)(nil)
var _ IAuth = (*AuthPhone)(nil)
var _ IAuth = (*AuthEmail)(nil)
var _ IAuth = (*AuthBiometric)(nil)
var _ IAuth = (*AuthThirdParty)(nil)

type (
	IAuth interface {
		IsEnabled() bool      // 检查认证方式是否启用
		IsBlocked() bool      // 检查认证方式是否被封禁
		IsActive() bool       // 检查认证方式是否已激活过
		GetAccountID() uint64 // 获取关联的账号ID
		GetKind() AuthKind    // 获取认证类型
		//Verify(credential any) bool // 验证认证信息
	}

	// Auth 可验证账号基础 TODO:GG 检查唯一性
	Auth struct {
		*model.Base
		AccountID uint64   `json:"accountId"` // 账户Id
		Kind      AuthKind `json:"kind"`      // 认证类型
	}

	// AuthPassword 用户名+密码
	AuthPassword struct {
		*Auth
		Username string `json:"username"` // 用户名

		Password string `json:"omitempty" gorm:"column:password_hash"` // 密码 (md5)
	}

	// AuthPhone 手机号+短信
	AuthPhone struct {
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

	// AuthBiometric 生物特征
	AuthBiometric struct {
		*Auth
		Kind AuthBioKind `json:"kind"` // 生物特征类型 (eg:人脸/指纹/声纹/...)
		// TODO:GG 有一个id?
	}

	// AuthThirdParty 第三方账号链接
	AuthThirdParty struct {
		*Auth
		Kind   AuthTPKind `json:"kind"`   // 平台 (eg:微信/QQ/...)
		OpenId string     `json:"openId"` // 第三方平台唯一标识
	}

	// AuthKind 认证类型
	AuthKind int16

	// AuthBioKind 生物特征类型
	AuthBioKind int16

	// AuthTPKind 第三方平台类型
	AuthTPKind int16
)

func NewAuthEmpty() *Auth {
	return &Auth{
		Base: model.NewBaseEmpty(),
	}
}

// NewAuth 创建指定类型的认证实例
func NewAuth(accountID uint64, kind AuthKind) *Auth {
	return &Auth{
		Base:      model.NewBase(make(data.KSMap)),
		AccountID: accountID, Kind: kind,
	}
}

func NewAuthPasswordEmpty() *AuthPassword {
	return &AuthPassword{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthPassword(
	accountID uint64,
	username, password string,
) *AuthPassword {
	return &AuthPassword{
		Auth:     NewAuth(accountID, AuthKindPassword),
		Username: username, Password: password,
	}
}

func NewAuthPhoneEmpty() *AuthPhone {
	return &AuthPhone{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthPhone(
	accountID uint64,
	code, number string,
) *AuthPhone {
	return &AuthPhone{
		Auth: NewAuth(accountID, AuthKindPhone),
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
	accountID uint64,
	username, domain string,
) *AuthEmail {
	return &AuthEmail{
		Auth:     NewAuth(accountID, AuthKindEmail),
		Username: username, Domain: domain,
		Entity: nil, TLD: nil,
	}
}

func NewAuthBiometricEmpty() *AuthBiometric {
	return &AuthBiometric{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthBiometric(
	accountID uint64,
	kind AuthBioKind,
) *AuthBiometric {
	return &AuthBiometric{
		Auth: NewAuth(accountID, AuthKindBiometric),
		Kind: kind,
	}
}

func NewAuthThirdPartyEmpty() *AuthThirdParty {
	return &AuthThirdParty{
		Auth: NewAuthEmpty(),
	}
}

func NewAuthThirdParty(
	accountID uint64,
	kind AuthTPKind, openId string,
) *AuthThirdParty {
	return &AuthThirdParty{
		Auth: NewAuth(accountID, AuthKindThirdParty),
		Kind: kind, OpenId: openId,
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
	AuthBioKindIris   AuthBioKind = 4 // 虹膜

	AuthTPKindWechat   AuthTPKind = 1 // 微信
	AuthTPKindQQ       AuthTPKind = 2 // QQ
	AuthTPKindIns      AuthTPKind = 3 // Instagram
	AuthTPKindFacebook AuthTPKind = 4 // Facebook
	AuthTPKindGoogle   AuthTPKind = 5 // Google
	AuthTPKindApple    AuthTPKind = 6 // Apple
)

func (a *Auth) IsEnabled() bool {
	return a.Status >= AuthStatusInit
}

func (a *Auth) IsBlocked() bool {
	blockedUntil, ok := a.GetBlockedUntil()
	if !ok {
		return false
	}
	blockTime := time.Unix(blockedUntil, 0)
	return time.Now().Before(blockTime)
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

func (a *Auth) Verify(_ any) bool {
	return a.IsEnabled() && !a.IsBlocked()
}

// Verify 实现密码验证
func (a *AuthPassword) Verify(credential any) bool {
	if !a.Auth.Verify(nil) {
		return false
	}
	cred, ok := credential.(struct {
		Username string
		Password string
	})
	if !ok {
		return false
	}

	//salt, _ := a.GetPasswordSalt()
	// 实际场景中应该使用安全的密码哈希比较
	// 例如: hashedPassword := HashPassword(cred.Password, salt)
	// TODO: 比较 hashedPassword 和 a.Password

	if (a.Username != cred.Username) || (a.Password != cred.Password) {
		return false
	}
	return true
}

// Verify 实现手机号验证
func (a *AuthPhone) Verify(credential any) bool {
	if !a.Auth.Verify(nil) {
		return false
	}
	cred, ok := credential.(struct {
		Code   string
		Number string
		Body   string
	})
	if !ok {
		return false
	}

	if (a.Code != cred.Code) || (a.Number != cred.Number) {
		return false
	}
	// else if cred.Body == nil {
	//	return false, nil
	//}
	//body, ok := cred.Verify.GetBody()
	//if !ok || body == "" {
	//	return false, nil
	//}
	//success := cred.Code == body
	//if !success {
	//	return false, nil
	//}
	return true
}

// Verify 实现邮箱验证码验证
func (a *AuthEmail) Verify(credential any) (bool, error) {
	//cred, ok := credential.(struct {
	//	Email string
	//	Code  string
	//})
	//
	//if !ok {
	//	return false, errors.New("invalid credential format")
	//}
	//
	//if a.EmailAddress() != cred.Email {
	//	return false, errors.New("email mismatch")
	//}
	//body, ok := verify.GetBody()
	//if !ok || body == "" {
	//	return false, errors.New("no verification code found")
	//}
	//
	//if verify.IsExpired() {
	//	return false, errors.New("verification code expired")
	//}
	//return body == cred.Code, nil
	return true, nil
}

// Verify 实现生物特征验证
func (a *AuthBiometric) Verify(credential any) (bool, error) {
	// 生物特征验证通常需要专门的SDK或API
	// 这里只是示例框架
	return false, errors.New("biometric verification not implemented")
}

// Verify 实现第三方平台验证
func (a *AuthThirdParty) Verify(credential any) (bool, error) {
	//cred, ok := credential.(struct {
	//	Provider    string
	//	AccessToken string
	//})
	//
	//if !ok {
	//	return false, errors.New("invalid credential format")
	//}
	//
	//if a.Provider != cred.Provider {
	//	return false, errors.New("provider mismatch")
	//}
	//
	//// 在实际应用中，应该调用第三方API验证token
	//// 这里简化处理
	//if a.AccessToken != cred.AccessToken {
	//	return false, errors.New("invalid access token")
	//}
	//
	//// 检查token是否过期
	//if a.TokenExpiredAt != nil && time.Now().After(*a.TokenExpiredAt) {
	//	return false, errors.New("access token expired")
	//}
	return true, nil
}

func (a *AuthPhone) PhoneNumber() string {
	return "+" + a.Code + " " + a.Number
}

func (a *AuthEmail) EmailAddress() string {
	return a.Username + "@" + a.Domain
}

const (
	authExtraKeyBlockedUntil   = "blockedUntil"   // 阻止验证直到某个时间点(过于频繁的send)
	authExtraKeyPasswordSalt   = "passwordSalt"   // 密码盐 TODO:GG 不response
	authExtraKeyUserSecondPwd  = "userSecondPwd"  // 二级密码 TODO:GG 不response
	authExtraKeyUserScreenPwd  = "userScreenPwd"  // 锁屏密码 TODO:GG 不response
	authExtraKeyPhoneUrgent    = "phoneUrgent"    // 紧急联系人 TODO:GG 不response
	authExtraKeyBioCredential  = "bioCredential"  // 生物特征凭证 TODO:GG 不response
	authExtraKeyTPAccessToken  = "tpAccessToken"  // 第三方平台访问令牌 TODO:GG 不response
	authExtraKeyTPRefreshToken = "tpRefreshToken" // 刷新令牌 TODO:GG 不response
	authExtraKeyTPExpiredAt    = "tpExpiredAt"    // 过期时间 TODO:GG 不response
	authExtraKeyTPUserInfo     = "tpUserInfo"     // 第三方平台用户信息 TODO:GG 不response
	authExtraKeyTPLinkId       = "tpLinkId"       // 关联ID TODO:GG 不response
)

func (a *Auth) SetBlockedUntil(blockUtilUnix *int64) {
	a.Extra.SetInt64(authExtraKeyBlockedUntil, blockUtilUnix)
}

func (a *Auth) GetBlockedUntil() (int64, bool) {
	return a.Extra.GetInt64(authExtraKeyBlockedUntil)
}

func (a *Auth) SetPasswordSalt(salt *string) {
	a.Extra.SetString(authExtraKeyPasswordSalt, salt)
}

func (a *Auth) GetPasswordSalt() (string, bool) {
	return a.Extra.GetString(authExtraKeyPasswordSalt)
}
