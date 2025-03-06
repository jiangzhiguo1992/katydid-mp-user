package model

import (
	"katydid-mp-user/internal/pkg/model"
	"time"
)

const (
	AccountKindPwd   uint16 = 1 // 账号+密码
	AccountKindPhone uint16 = 2 // 手机+短信码
	AccountKindEmail uint16 = 3 // 邮箱+邮件码
	AccountKindBio   uint16 = 4 // 生物特征
	AccountKingThird uint16 = 5 // 第三方+平台验证
)

// Account 账户
type Account struct {
	*model.Base

	UserId uint64 `json:"userId"` // 用户Id

	//AuthKinds []int             `json:"authKinds"` // 可用验证类型
	//AuthOnAts map[int][]int64   `json:"authOnAts"` // [kind][]验证发送时间
	//AuthBodys map[int]*AuthInfo `json:"authBodys"` // [kind]验证内容
	//AuthOkAts map[int][]int64   `json:"authOkAts"` // [kind][]验证成功时间

	clientId int64 `json:"clientId"` // 客户端Id
	orgId    int64 `json:"orgId"`    // 组织Id
	//Logins   map[int64][]LoginInfo     `json:"logins"`   // [clientId]登录信息

	ActiveAts   map[uint64][]int64  `json:"activeAts"`   // [clientId]激活/登录客户端时间集合 (最早的就是注册的平台)
	Tokens      map[uint64][]string `json:"tokens"`      // [clientId]访问tokens
	ToExpireAts map[uint64][]int64  `json:"toExpireAts"` // [clientId]访问tokens过期时间 -1为永久
	Enables     map[uint64]bool     `json:"enables"`     // [clientId]是否可用 (-1为都不可用)
	Reasons     map[uint64]string   `json:"reasons"`     // [clientId]拒绝原因 (-1为都不可用)

	AuthKinds []int `json:"authKinds"` // 可用验证类型

	//LastSigninTime int64
	//LastSigninIp   int64
	// TODO:GG 下面两个要不要精简上去
	//LastLoginAt time.Time `json:"lastLoginAt"`
	// Actives
	//ClientIds uint64 `json:"clientIds"` // 客户端Id
	//Logins   map[int64][]LoginInfo     `json:"logins"`   // [clientId]登录信息
	//Versions map[int64][]ClientVersion `json:"versions"` // 版本分布
}

func NewAccountDef(userId uint64) *Account {
	return &Account{
		Base:   model.NewBaseEmpty(),
		UserId: userId,
		//AuthKinds: []int{},
		//AuthOnAts: map[int][]int64{},
		//AuthBodys: map[int]*AuthInfo{},
		//AuthOkAts: map[int][]int64{},
		//
		//ActiveAts:   map[uint64][]int64{},
		//Tokens:      map[uint64][]string{},
		//ToExpireAts: map[uint64][]int64{},
		//Enables:     map[uint64]bool{},
		//Reasons:     map[uint64]string{},

		//Extra: map[string]any{},

		//Logins:   map[int64][]LoginInfo{},
		//Versions: map[int64][]ClientVersion{},
	}
}

func NewAccountEmpty() *Account {
	return &Account{
		//*base.DBModel
		//AuthKinds: []int{},
		//AuthOnAts: map[int][]int64{},
		//AuthBodys: map[int]*AuthInfo{},
		//AuthOkAts: map[int][]int64{},
		//
		//ActiveAts:   map[uint64][]int64{},
		//Tokens:      map[uint64][]string{},
		//ToExpireAts: map[uint64][]int64{},
		//Enables:     map[uint64]bool{},
		//Reasons:     map[uint64]string{},

		//Extra: map[string]any{},

		//Logins:   map[int64][]LoginInfo{},
		//Versions: map[int64][]ClientVersion{},
	}
}

func (a *Account) IsTokenOk(clientId uint64, token string) (bool, bool) {
	isTokenOk := false
	isTokenExpireOk := false
	if tokens, toOk := a.Tokens[clientId]; toOk {
		for toIndex, to := range tokens {
			if to == token {
				isTokenOk = true
				if expires, exOk := a.ToExpireAts[clientId]; exOk {
					if len(expires) > toIndex {
						expire := expires[toIndex]
						if (expire == -1) || (expire > time.Now().UnixMilli()) {
							isTokenExpireOk = true
						}
					}
				}
				break
			}
		}
	}
	return isTokenOk, isTokenExpireOk
}

func (a *Account) NewToken(clientId uint64, userTokenMax int) (string, int64) {
	if _, ok := a.Tokens[clientId]; !ok {
		a.Tokens[clientId] = []string{}
	}
	if _, ok := a.ToExpireAts[clientId]; !ok {
		a.ToExpireAts[clientId] = []int64{}
	}
	if len(a.Tokens[clientId]) >= userTokenMax {
		a.Tokens[clientId] = a.Tokens[clientId][1:]
		a.ToExpireAts[clientId] = a.ToExpireAts[clientId][1:]
	}
	newToken := "随机生成"
	newTokenExpireAt := time.Now().UnixMilli() + 3600
	a.Tokens[clientId] = append(a.Tokens[clientId], "随机生成")
	a.ToExpireAts[clientId] = append(a.ToExpireAts[clientId], newTokenExpireAt)
	return newToken, newTokenExpireAt
}

// SetAvatarUrl 头像
func (a *Account) SetAvatarUrl(supportUrl *string) {
	if (supportUrl != nil) && (len(*supportUrl) > 0) {
		a.Extra["avatarUrl"] = *supportUrl
	} else {
		delete(a.Extra, "avatarUrl")
	}
}

func (a *Account) GetSupportUrl() string {
	if a.Extra["avatarUrl"] == nil {
		return ""
	}
	return a.Extra["avatarUrl"].(string)
}
