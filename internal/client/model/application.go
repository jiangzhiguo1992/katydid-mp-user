package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/valid"
	"katydid-mp-user/utils"
	"reflect"
	"time"
	"unicode"
)

// Application 应用
type Application struct {
	*model.Base
	OrgId uint64 `json:"orgId" validate:"required"` // 组织
	IP    uint   `json:"IP"`                        // 系列 (eg:大富翁IP)
	Part  uint   `json:"part"`                      // 类型 (eg:单机版)

	Name      string `json:"name" validate:"required,name-format"` // 应用名称
	OnlineAt  int64  `json:"onlineAt"`                             // 上线时间 (时间没到时，只能停留在首页，提示bulletins)
	OfflineAt int64  `json:"offlineAt"`                            // 下线时间 (时间到后，强制下线+升级/等待/...)

	Maintenances []*Maintenance `json:"maintenances" gorm:"-:all"` // 维护时间(App的)
	Clients      []*Client      `json:"clients" gorm:"-:all"`      // [platform][area]平台列表
}

func NewApplicationEmpty() *Application {
	return &Application{
		Base:         model.NewBaseEmpty(),
		Maintenances: make([]*Maintenance, 0),
		Clients:      make([]*Client, 0),
	}
}

func NewApplicationDefault(
	orgId uint64, IP, part uint,
	name string,
) *Application {
	return &Application{
		Base:  model.NewBaseDefault(),
		OrgId: orgId, IP: IP, Part: part,
		Name: name, OnlineAt: 0, OfflineAt: 0,
		Maintenances: make([]*Maintenance, 0),
		Clients:      make([]*Client, 0),
	}
}

func (a *Application) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 应用名称 (1-50)
			"name-format": func(value reflect.Value, param string) bool {
				name := value.String()
				if len(name) < 1 || len(name) > 50 {
					return false
				}
				for _, r := range name {
					if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' {
						return false
					}
				}
				return true
			},
		},
	}
}

func (a *Application) ValidExtraRules() (utils.KSMap, valid.ExtraValidRules) {
	return a.Extra, valid.ExtraValidRules{
		valid.SceneAll: valid.ExtraValidRule{
			// 官网 (<1000)
			appExtraKeyWebsiteUrl: valid.ExtraValidRuleInfo{
				Field: appExtraKeyWebsiteUrl,
				ValidFn: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
			// 版权 (<100)*(<1000)
			appExtraKeyCopyrights: valid.ExtraValidRuleInfo{
				Field: appExtraKeyCopyrights,
				ValidFn: func(value interface{}) bool {
					vs, ok := value.([]string)
					if !ok {
						return false
					}
					if len(vs) > 100 {
						return false
					}
					for _, v := range vs {
						if len(v) > 1000 {
							return false
						}
					}
					return true
				},
			},
			// 服务条款URL (<1000)
			appExtraKeySupportUrl: valid.ExtraValidRuleInfo{
				Field: appExtraKeySupportUrl,
				ValidFn: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
			// 隐私政策URL (<1000)
			appExtraKeyPrivacyUrl: valid.ExtraValidRuleInfo{
				Field: appExtraKeyPrivacyUrl,
				ValidFn: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
		},
	}
}

func (a *Application) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"Name": {"format_s_input_required", false, []any{"app_name"}},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"name-format":         {"format_app_name_err", false, nil},
				appExtraKeyWebsiteUrl: {"format_website_err", false, nil},
				appExtraKeyCopyrights: {"format_copyrights_err", false, nil},
				appExtraKeySupportUrl: {"format_support_url_err", false, nil},
				appExtraKeyPrivacyUrl: {"format_privacy_url_err", false, nil},
			},
		},
	}
}

func (a *Application) OverUserMaxAccount(count int) bool {
	maxCount := a.GetUserMaxAccount()
	return (maxCount >= 0) && (count > maxCount)
}

func (a *Application) OverUserMaxToken(count int) bool {
	maxCount := a.GetUserMaxToken()
	return (maxCount >= 0) && (count > maxCount)
}

// IsOnline 是否上线
func (a *Application) IsOnline() bool {
	currentTime := time.Now().UnixMilli()
	return (a.OnlineAt > 0 && (a.OnlineAt <= currentTime)) && (a.OfflineAt == -1 || a.OfflineAt > currentTime)
}

// IsOffline 是否下线
func (a *Application) IsOffline() bool {
	currentTime := time.Now().UnixMilli()
	return (a.OfflineAt > 0 && (a.OfflineAt <= currentTime)) && (a.OnlineAt == -1 || a.OnlineAt > currentTime)
}

// IsComingOnline 是否即将上线
func (a *Application) IsComingOnline() bool {
	currentTime := time.Now().UnixMilli()
	return a.OnlineAt > currentTime && (a.OfflineAt == -1 || a.OfflineAt < currentTime)
}

// IsComingOffline 是否即将下线
func (a *Application) IsComingOffline() bool {
	currentTime := time.Now().UnixMilli()
	return a.OfflineAt > currentTime && (a.OnlineAt == -1 || a.OnlineAt < currentTime)
}

func (a *Application) GetClientsByPlatform(platform uint) []*Client {
	var clients []*Client
	for _, c := range a.Clients {
		if c.Platform == platform {
			clients = append(clients, c)
		}
	}
	return clients
}

func (a *Application) GetClientsByArea(area uint) []*Client {
	var clients []*Client
	for _, c := range a.Clients {
		if c.Area == area {
			clients = append(clients, c)
		}
	}
	return clients
}

func (a *Application) GetClient(platform, area uint) *Client {
	for _, c := range a.Clients {
		if c.Platform == platform && c.Area == area {
			return c
		}
	}
	return nil
}

func (a *Application) GetLatestVersion(platform, area uint, market uint) *Version {
	p := a.GetClient(platform, area)
	return p.GetLatestVersion(market)
}

const (
	appExtraKeyCerts = "certs"

	appExtraKeyDesc           = "desc"           // 简介
	appExtraKeyLogoUrl        = "logoUrl"        // 图标
	appExtraKeyWebsiteUrl     = "websiteUrl"     // 官网
	appExtraKeyCopyrights     = "copyrights"     // 版权
	appExtraKeySupportUrl     = "supportUrl"     // 服务条款URL
	appExtraKeyPrivacyUrl     = "privacyUrl"     // 隐私政策URL
	appExtraKeyUserMaxAccount = "userMaxAccount" // 用户最多账户数
	appExtraKeyUserMaxToken   = "userMaxToken"   // 用户最多令牌数 (同时登录最大数，防止工作室?)

	// TODO:GG
	appExtraKeyCanPwd           = "enablePwd" // 是否支持密码登录
	appExtraKeyCanSignUp        = "canSignUp" // 是否支持密码登录
	appExtraKeyCanSignUpCode    = "enableSession"
	appExtraKeyCanSignInSession = "enableSession"
	// SigninMethods // TODO:GG 支持的账号认证，必须的账号认证?
	//Providers           []*ProviderItem `xorm:"mediumtext" json:"providers"`
	//SignupItems         []*SignupItem   `xorm:"varchar(1000)" json:"signupItems"`
	//ExpireInHours        int            `json:"expireInHours"`
	//RefreshExpireInHours int            `json:"refreshExpireInHours"`
)

func (a *Application) SetCerts(certs *[]any) {
	a.Extra.SetSlice(appExtraKeyCerts, certs)
}

func (a *Application) GetCerts() []any {
	data, _ := a.Extra.GetSlice(appExtraKeyCerts)
	return data
}

func (a *Application) SetDesc(desc *string) {
	a.Extra.SetString(appExtraKeyDesc, desc)
}

func (a *Application) GetDesc() string {
	data, _ := a.Extra.GetString(appExtraKeyDesc)
	return data
}

func (a *Application) SetLogoUrl(logoUrl *string) {
	a.Extra.SetString(appExtraKeyLogoUrl, logoUrl)
}

func (a *Application) GetLogoUrl() string {
	data, _ := a.Extra.GetString(appExtraKeyLogoUrl)
	return data
}

func (a *Application) SetWebsiteUrl(website *string) {
	a.Extra.SetString(appExtraKeyWebsiteUrl, website)
}

func (a *Application) GetWebsiteUrl() string {
	data, _ := a.Extra.GetString(appExtraKeyWebsiteUrl)
	return data
}

func (a *Application) SetCopyrights(copyrights *[]string) {
	a.Extra.SetStringSlice(appExtraKeyCopyrights, copyrights)
}

func (a *Application) GetCopyrights() []string {
	data, _ := a.Extra.GetStringSlice(appExtraKeyCopyrights)
	return data
}

func (a *Application) SetSupportUrl(supportUrl *string) {
	a.Extra.SetString(appExtraKeySupportUrl, supportUrl)
}

func (a *Application) GetSupportUrl() string {
	data, _ := a.Extra.GetString(appExtraKeySupportUrl)
	return data
}

func (a *Application) SetPrivacyUrl(privacyUrl *string) {
	a.Extra.SetString(appExtraKeyPrivacyUrl, privacyUrl)
}

func (a *Application) GetPrivacyUrl() string {
	data, _ := a.Extra.GetString(appExtraKeyPrivacyUrl)
	return data
}

func (a *Application) SetUserMaxAccount(userMaxAccount *int) {
	a.Extra.SetInt(appExtraKeyUserMaxAccount, userMaxAccount)
}

func (a *Application) GetUserMaxAccount() int {
	data, _ := a.Extra.GetInt(appExtraKeyUserMaxAccount)
	return data
}

func (a *Application) SetUserMaxToken(userMaxToken *int) {
	a.Extra.SetInt(appExtraKeyUserMaxToken, userMaxToken)
}

func (a *Application) GetUserMaxToken() int {
	data, _ := a.Extra.GetInt(appExtraKeyUserMaxToken)
	return data
}
