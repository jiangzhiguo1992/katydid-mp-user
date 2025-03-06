package model

import (
	"errors"
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/utils"
	"time"
)

// Account 账号结构
type Account struct {
	*model.Base

	UserID *uint64 `json:"userId"` // 账号拥有者Id (有些org/app不填user)

	Nickname string        `json:"nickname"` // 昵称 (没有user的app/org会用这个，放外面是方便搜索)
	Status   AccountStatus `json:"status"`   // 账号状态

	Tokens    map[TokenOwn]map[uint64]string `json:"tokens"`    // token列表(jwt)
	ActiveAts map[TokenOwn]map[uint64]int64  `json:"activeAts"` // 激活组织时间集合 (最早的就是注册的平台)
	ExpireAts map[TokenOwn]map[uint64]int64  `json:"expireAts"` // token过期时间列表 -1为永久

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
		Tokens:           make(map[TokenOwn]map[uint64]string),
		ActiveAts:        make(map[TokenOwn]map[uint64]int64),
		ExpireAts:        make(map[TokenOwn]map[uint64]int64),
		AuthBiometrics:   make([]*Biometric, 0),
		AuthThirdParties: make([]*ThirdPartyLink, 0),
		LoginHistory:     make([]*LoginInfo, 0),
	}
}

func NewAccount(userID *uint64, nickname string) *Account {
	return &Account{
		Base:             model.NewBase(make(utils.KSMap)),
		UserID:           userID,
		Nickname:         nickname,
		Status:           AccountStatusNormal,
		Tokens:           make(map[TokenOwn]map[uint64]string),
		ActiveAts:        make(map[TokenOwn]map[uint64]int64),
		ExpireAts:        make(map[TokenOwn]map[uint64]int64),
		AuthPassword:     nil,
		AuthPhone:        nil,
		AuthEmail:        nil,
		AuthBiometrics:   make([]*Biometric, 0),
		AuthThirdParties: make([]*ThirdPartyLink, 0),
		LoginHistory:     make([]*LoginInfo, 0),
	}
}

// AccountStatus 账号状态
type AccountStatus int8

const (
	AccountStatusBanned = -2 // 封禁 (不能访问任何api)
	AccountStatusLocked = -1 // 锁定 (不能登录)
	AccountStatusNormal = 0  // 正常 (必须有附带一个认证才能创建)
)

// CanLogin 可否登录
func (a *Account) CanLogin() bool {
	return a.Status > AccountStatusLocked
}

// CanAccess 可否访问
func (a *Account) CanAccess() bool {
	return a.Status > AccountStatusBanned
}

// GetAvailableAuthMethods 获取可用的认证方式列表
func (a *Account) GetAvailableAuthMethods() []AuthKind {
	methods := make([]AuthKind, 0)
	if a.AuthPassword != nil {
		methods = append(methods, AuthKindPwd)
	}
	if a.AuthPhone != nil {
		methods = append(methods, AuthKindPhone)
	}
	if a.AuthEmail != nil {
		methods = append(methods, AuthKindEmail)
	}
	if len(a.AuthBiometrics) > 0 {
		methods = append(methods, AuthKindBio)
	}
	if len(a.AuthThirdParties) > 0 {
		methods = append(methods, AuthKindThird)
	}
	return methods
}

// CountAuthMethods 计算已有认证方式数量
func (a *Account) CountAuthMethods() int {
	count := 0
	if a.AuthPassword != nil {
		count++
	}
	if a.AuthPhone != nil {
		count++
	}
	if a.AuthEmail != nil {
		count++
	}
	count += len(a.AuthBiometrics)
	count += len(a.AuthThirdParties)
	return count
}

// InvalidateToken 使token失效
func (a *Account) InvalidateToken(ownType TokenOwn, ownID uint64) {
	if owns, ok := a.Tokens[ownType]; ok {
		delete(owns, ownID)
	}
	if owns, ok := a.ExpireAts[ownType]; ok {
		delete(owns, ownID)
	}
}

// InvalidateAllTokens 使所有token失效
func (a *Account) InvalidateAllTokens() {
	a.Tokens = make(map[TokenOwn]map[uint64]string)
	a.ExpireAts = make(map[TokenOwn]map[uint64]int64)
}

// GetToken 获取token
func (a *Account) GetToken(ownType TokenOwn, ownID uint64) (string, bool) {
	if owns, ok := a.Tokens[ownType]; ok {
		if token, ok := owns[ownID]; ok {
			return token, true
		}
	}
	return "", false
}

// IsTokenExpired 检查token是否已过期
func (a *Account) IsTokenExpired(ownType TokenOwn, ownID uint64) bool {
	if expireAts, ok := a.ExpireAts[ownType]; ok {
		if expireAt, ok := expireAts[ownID]; ok {
			if expireAt < 0 { // -1表示永不过期
				return false
			}
			return time.Now().Unix() > expireAt
		}
	}
	return true // 如果找不到过期时间，视为已过期
}

// GenerateTokens 为账号生成token
func (a *Account) GenerateTokens(
	ownType TokenOwn, ownID uint64, jwtSecret string,
	expireSec, refExpireHou int64, issuer string,
) (*Token, error) {
	// 检查账号状态
	if !a.CanLogin() {
		return nil, errors.New("account cannot login")
	}
	// 旧的token
	oldToken, _ := a.GetToken(ownType, ownID)
	// 创建新的Token
	token := NewToken(a.ID, a.UserID, ownType, ownID, expireSec, refExpireHou, issuer)
	// 生成JWT令牌
	if err := token.GenerateJWTTokens(jwtSecret, &oldToken); err != nil {
		return nil, err
	}
	if a.Tokens[ownType] == nil {
		a.Tokens[ownType] = make(map[uint64]string)
	}
	if a.ActiveAts[ownType] == nil {
		a.ActiveAts[ownType] = make(map[uint64]int64)
	}
	if a.ExpireAts[ownType] == nil {
		a.ExpireAts[ownType] = make(map[uint64]int64)
	}
	// 记录token
	a.Tokens[ownType][ownID] = token.AccessToken
	a.ExpireAts[ownType][ownID] = time.Now().Add(time.Duration(token.ExpireSec) * time.Second).Unix()
	// 如果是首次激活，记录激活时间
	if _, exists := a.ActiveAts[ownType][ownID]; !exists {
		a.ActiveAts[ownType][ownID] = time.Now().Unix()
	}
	return token, nil
}

// ValidateToken 验证token
func (a *Account) ValidateToken(
	tokenStr string, ownType TokenOwn, ownID uint64,
	jwtSecret string, checkExpire bool,
) (*TokenClaims, error) {
	// 检查账号状态
	if !a.CanAccess() {
		return nil, errors.New("account cannot access")
	}

	// 验证token是否存在于账户中
	storedToken, exists := a.GetToken(ownType, ownID)
	if !exists {
		return nil, errors.New("token not found")
	}

	// 检查提供的token与存储的token是否匹配
	if storedToken != tokenStr {
		return nil, errors.New("invalid token")
	}

	// 检查过期时间
	if checkExpire && a.IsTokenExpired(ownType, ownID) {
		return nil, errors.New("token expired")
	}
	// 解析和验证JWT
	return ParseJWT(tokenStr, jwtSecret, checkExpire)
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
