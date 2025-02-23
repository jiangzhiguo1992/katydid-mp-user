package model

import (
	"katydid-mp-user/internal/pkg/model"
	"time"
)

var _ model.IValid = (*Client)(nil)

// Client 客户端
type Client struct {
	*model.Base
	TeamId uint64 `json:"teamId"` // 团队
	IP     uint   `json:"IP"`     // 系列 (eg:大富翁IP)
	Part   uint   `json:"part"`   // 类型 (eg:单机版)

	Enable    bool  `json:"enable"`    // 是否可用 (一般不用，下架之类的，没有reason)
	OnlineAt  int64 `json:"onlineAt"`  // 上线时间 (时间没到时，只能停留在首页，提示bulletins)
	OfflineAt int64 `json:"offlineAt"` // 下线时间 (时间到后，强制下线+升级/等待/...)

	Name string `json:"name" valid:"client-name"` // 客户端名称

	//Platforms map[uint16]map[uint16]*Platform `json:"platforms" gorm:"-:all"` // [platform][area]平台列表
}

func NewClientEmpty() *Client {
	return &Client{Base: model.NewBaseEmpty()}
}

func NewClientDefault(
	teamId uint64, ip, part uint,
	enable bool,
	name string,
) *Client {
	client := &Client{
		Base:   model.NewBaseEmpty(),
		TeamId: teamId, IP: ip, Part: part,
		Enable: enable, OnlineAt: 0, OfflineAt: 0,
		Name: name,
		//Platforms: make(map[uint16]map[uint16]*Platform),
	}
	return client
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

// SetWebsite 官网
func (c *Client) SetWebsite(website *string) {
	if (website != nil) && (len(*website) > 0) {
		c.Extra["website"] = *website
	} else {
		delete(c.Extra, "website")
	}
}

func (c *Client) GetWebsite() string {
	if c.Extra["website"] == nil {
		return ""
	}
	return c.Extra["website"].(string)
}

// SetCopyrights 版权
func (c *Client) SetCopyrights(copyrights *[]string) {
	if (copyrights != nil) && (len(*copyrights) > 0) {
		c.Extra["copyrights"] = *copyrights
	} else {
		delete(c.Extra, "copyrights")
	}
}

func (c *Client) GetCopyrights() []string {
	if c.Extra["copyrights"] == nil {
		return []string{}
	}
	return c.Extra["copyrights"].([]string)
}

// SetSupportUrl 服务条款URL
func (c *Client) SetSupportUrl(supportUrl *string) {
	if (supportUrl != nil) && (len(*supportUrl) > 0) {
		c.Extra["supportUrl"] = *supportUrl
	} else {
		delete(c.Extra, "supportUrl")
	}
}

func (c *Client) GetSupportUrl() string {
	if c.Extra["supportUrl"] == nil {
		return ""
	}
	return c.Extra["supportUrl"].(string)
}

// SetPrivacyUrl 隐私政策URL
func (c *Client) SetPrivacyUrl(privacyUrl *string) {
	if (privacyUrl != nil) && (len(*privacyUrl) > 0) {
		c.Extra["privacyUrl"] = *privacyUrl
	} else {
		delete(c.Extra, "privacyUrl")
	}
}

func (c *Client) GetPrivacyUrl() string {
	if c.Extra["privacyUrl"] == nil {
		return ""
	}
	return c.Extra["privacyUrl"].(string)
}

// SetUserMaxAccount 用户最多账户数 (身份证/护照/...)
func (c *Client) SetUserMaxAccount(userMaxAccount *int) {
	if (userMaxAccount != nil) && (*userMaxAccount > 0) {
		c.Extra["userMaxAccount"] = *userMaxAccount
	} else {
		delete(c.Extra, "userMaxAccount")
	}
}

func (c *Client) GetUserMaxAccount() int {
	if c.Extra["userMaxAccount"] == nil {
		return 0
	}
	return c.Extra["userMaxAccount"].(int)
}

func (c *Client) OverUserMaxAccount(count int) bool {
	maxCount := c.GetUserMaxAccount()
	if maxCount <= 0 {
		return false
	}
	return count > maxCount
}

// SetUserMaxToken 用户最多令牌数 (同时登录最大数，防止工作室?)
func (c *Client) SetUserMaxToken(userMaxToken *int) {
	if (userMaxToken != nil) && (*userMaxToken > 0) {
		c.Extra["userMaxToken"] = *userMaxToken
	} else {
		delete(c.Extra, "userMaxToken")
	}
}

func (c *Client) GetUserMaxToken() int {
	if c.Extra["userMaxToken"] == nil {
		return 0
	}
	return c.Extra["userMaxToken"].(int)
}

func (c *Client) OverUserMaxToken(count int) bool {
	maxCount := c.GetUserMaxToken()
	if maxCount <= 0 {
		return false
	}
	return count > maxCount
}

//func (c *Client) GetPlatform(platform, area uint16) *Platform {
//	if _, ok := c.Platforms[platform]; !ok {
//		return nil
//	}
//	return c.Platforms[platform][area]
//}
//
//func (c *Client) GetLatestVersion(platform, area uint16, market uint) *Version {
//	p := c.GetPlatform(platform, area)
//	return p.GetLatestVersion(market)
//}
