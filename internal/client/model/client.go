package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/valid"
	"katydid-mp-user/utils"
	"reflect"
	"time"
	"unicode"
)

const (
	clientExtraKeyWebsite        = "website"        // 官网
	clientExtraKeyCopyrights     = "copyrights"     // 版权
	clientExtraKeySupportUrl     = "supportUrl"     // 服务条款URL
	clientExtraKeyPrivacyUrl     = "privacyUrl"     // 隐私政策URL
	clientExtraKeyUserMaxAccount = "userMaxAccount" // 用户最多账户数
	clientExtraKeyUserMaxToken   = "userMaxToken"   // 用户最多令牌数 (同时登录最大数，防止工作室?)
)

// Client 客户端
type Client struct {
	*model.Base
	TeamId uint64 `json:"teamId"` // 团队
	IP     uint   `json:"IP"`     // 系列 (eg:大富翁IP)
	Part   uint   `json:"part"`   // 类型 (eg:单机版)

	Enable bool   `json:"enable"`                               // 是否可用 (一般不用，下架之类的，没有reason)
	Name   string `json:"name" validate:"required,name-format"` // 客户端名称

	OnlineAt  int64 `json:"onlineAt"`  // 上线时间 (时间没到时，只能停留在首页，提示bulletins)
	OfflineAt int64 `json:"offlineAt"` // 下线时间 (时间到后，强制下线+升级/等待/...)
	//RemainingTime // TODO:GG 维护信息

	Platforms []*Platform `json:"platforms" gorm:"-:all"` // [platform][area]平台列表
}

func NewClientEmpty() *Client {
	return &Client{Base: model.NewBaseEmpty()}
}

func NewClientDefault(
	teamId uint64, IP, part uint,
	enable bool, name string,
) *Client {
	client := &Client{
		Base:   model.NewBaseDefault(),
		TeamId: teamId, IP: IP, Part: part,
		Enable: enable, Name: name,
		OnlineAt: 0, OfflineAt: 0,
		Platforms: []*Platform{},
	}
	return client
}

func (c *Client) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 客户端名称 (1-50)
			"name-format": func(value reflect.Value) bool {
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

func (c *Client) ValidExtraRules() (utils.KSMap, valid.ExtraValidRules) {
	return c.Extra, valid.ExtraValidRules{
		valid.SceneAll: valid.ExtraValidRule{
			// 官网 (<1000)
			clientExtraKeyWebsite: valid.ExtraValidRuleInfo{
				Field: clientExtraKeyWebsite,
				Validate: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
			// 版权 (<100)*(<1000)
			clientExtraKeyCopyrights: valid.ExtraValidRuleInfo{
				Field: clientExtraKeyCopyrights,
				Validate: func(value interface{}) bool {
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
			clientExtraKeySupportUrl: valid.ExtraValidRuleInfo{
				Field: clientExtraKeySupportUrl,
				Validate: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
			// 隐私政策URL (<1000)
			clientExtraKeyPrivacyUrl: valid.ExtraValidRuleInfo{
				Field: clientExtraKeyPrivacyUrl,
				Validate: func(value interface{}) bool {
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

func (c *Client) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName][3]interface{}{
				valid.TagRequired: {
					"Name": {"format_s_input_required", false, []any{"client_name"}},
				},
			}, Rule2: map[valid.Tag][3]interface{}{
				"name-format":            {"format_client_name_err", false, nil},
				clientExtraKeyWebsite:    {"format_website_err", false, nil},
				clientExtraKeyCopyrights: {"format_copyrights_err", false, nil},
				clientExtraKeySupportUrl: {"format_support_url_err", false, nil},
				clientExtraKeyPrivacyUrl: {"format_privacy_url_err", false, nil},
			},
		},
	}
}

func (c *Client) SetWebsite(website *string) {
	c.Extra.SetString(clientExtraKeyWebsite, website)
}

func (c *Client) GetWebsite() string {
	data, _ := c.Extra.GetString(clientExtraKeyWebsite)
	return data
}

func (c *Client) SetCopyrights(copyrights *[]string) {
	c.Extra.SetStringSlice(clientExtraKeyCopyrights, copyrights)
}

func (c *Client) GetCopyrights() []string {
	data, _ := c.Extra.GetStringSlice(clientExtraKeyCopyrights)
	return data
}

func (c *Client) SetSupportUrl(supportUrl *string) {
	c.Extra.SetString(clientExtraKeySupportUrl, supportUrl)
}

func (c *Client) GetSupportUrl() string {
	data, _ := c.Extra.GetString(clientExtraKeySupportUrl)
	return data
}

func (c *Client) SetPrivacyUrl(privacyUrl *string) {
	c.Extra.SetString(clientExtraKeyPrivacyUrl, privacyUrl)
}

func (c *Client) GetPrivacyUrl() string {
	data, _ := c.Extra.GetString(clientExtraKeyPrivacyUrl)
	return data
}

func (c *Client) SetUserMaxAccount(userMaxAccount *int) {
	c.Extra.SetInt(clientExtraKeyUserMaxAccount, userMaxAccount)
}

func (c *Client) GetUserMaxAccount() int {
	data, _ := c.Extra.GetInt(clientExtraKeyUserMaxAccount)
	return data
}

func (c *Client) OverUserMaxAccount(count int) bool {
	maxCount := c.GetUserMaxAccount()
	return (maxCount >= 0) && (count > maxCount)
}

func (c *Client) SetUserMaxToken(userMaxToken *int) {
	c.Extra.SetInt(clientExtraKeyUserMaxToken, userMaxToken)
}

func (c *Client) GetUserMaxToken() int {
	data, _ := c.Extra.GetInt(clientExtraKeyUserMaxToken)
	return data
}

func (c *Client) OverUserMaxToken(count int) bool {
	maxCount := c.GetUserMaxToken()
	return (maxCount >= 0) && (count > maxCount)
}

// IsOnline 是否上线
func (c *Client) IsOnline() bool {
	currentTime := time.Now().UnixMilli()
	return (c.OnlineAt > 0 && (c.OnlineAt <= currentTime)) && (c.OfflineAt == -1 || c.OfflineAt > currentTime)
}

// IsOffline 是否下线
func (c *Client) IsOffline() bool {
	currentTime := time.Now().UnixMilli()
	return (c.OfflineAt > 0 && (c.OfflineAt <= currentTime)) && (c.OnlineAt == -1 || c.OnlineAt > currentTime)
}

// IsComingOnline 是否即将上线
func (c *Client) IsComingOnline() bool {
	currentTime := time.Now().UnixMilli()
	return c.OnlineAt > currentTime && (c.OfflineAt == -1 || c.OfflineAt < currentTime)
}

// IsComingOffline 是否即将下线
func (c *Client) IsComingOffline() bool {
	currentTime := time.Now().UnixMilli()
	return c.OfflineAt > currentTime && (c.OnlineAt == -1 || c.OnlineAt < currentTime)
}

func (c *Client) GetPlatformsByPlat(platform uint16) []*Platform {
	var platforms []*Platform
	for _, p := range c.Platforms {
		if p.Platform == platform {
			platforms = append(platforms, p)
		}
	}
	return platforms
}

func (c *Client) GetPlatformsByArea(area uint) []*Platform {
	var platforms []*Platform
	for _, p := range c.Platforms {
		if p.Area == area {
			platforms = append(platforms, p)
		}
	}
	return platforms
}

func (c *Client) GetPlatform(platform uint16, area uint) *Platform {
	for _, p := range c.Platforms {
		if p.Platform == platform && p.Area == area {
			return p
		}
	}
	return nil
}

//func (c *Client) GetLatestVersion(platform, area uint16, market uint) *Version {
//	p := c.GetPlatform(platform, area)
//	return p.GetLatestVersion(market)
//}
