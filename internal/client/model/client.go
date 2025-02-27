package model

import (
	"katydid-mp-user/internal/pkg/model"
	"time"
)

// Client 客户端平台
type Client struct {
	*model.Base
	AppId    uint64 `json:"appId" validate:"required"`    // 应用id
	Platform uint   `json:"platform" validate:"required"` // 平台
	Area     uint   `json:"area" validate:"required"`     // 区域编号

	OnlineAt  int64  `json:"onlineAt"`  // 上线时间 (时间没到时，只能停留在首页，提示bulletins)
	OfflineAt int64  `json:"offlineAt"` // 下线时间 (时间到后，强制下线+升级/等待/...)
	PkgId     string `json:"pkgId"`     // 各平台应用唯一标识 (pkg/bundle，海外和大陆可以同时安装!)
	AppName   string `json:"appName"`   // app名称 (不同Area的区别)

	Maintenances []*Maintenance    `json:"maintenances" gorm:"-:all"` // 维护时间(Client的)
	Versions     map[uint]*Version `json:"versions" gorm:"-:all"`     // [market]最新publish版本号
}

func NewClientEmpty() *Client {
	return &Client{
		Base:         model.NewBaseEmpty(),
		Maintenances: make([]*Maintenance, 0),
		Versions:     make(map[uint]*Version),
	}
}

func NewClientDefault(
	appId uint64, platform uint, area uint,
	pkgId, appName string,
) *Client {
	return &Client{
		Base:  model.NewBaseEmpty(),
		AppId: appId, Platform: platform, Area: area,
		OnlineAt: 0, OfflineAt: 0, PkgId: pkgId, AppName: appName,
		Maintenances: make([]*Maintenance, 0),
		Versions:     make(map[uint]*Version),
	}
}

const (
	ClientPlatformLinux   uint = 1 // Linux
	ClientPlatformWindows uint = 2 // Windows
	ClientPlatformMacOS   uint = 3 // MacOS
	ClientPlatformWeb     uint = 4 // Web
	ClientPlatformAndroid uint = 5 // Android
	ClientPlatformIOS     uint = 6 // IOS
	ClientPlatformApplet  uint = 7 // Applet
	ClientPlatformTvAnd   uint = 8 // TvAnd
	ClientPlatformTvIOS   uint = 9 // TvIOS

	ClientAreaWord      uint = 1 // 全球 (默认英文，泛指海外)
	ClientAreaChinaLand uint = 2 // 中国大陆 (简体中文)
	ClientAreaChinaHMT  uint = 3 // 中国港澳台 (繁体中文)
	ClientAreaEurope    uint = 4 // 欧洲 (默认英文，GDPR)
	// 后面可能会划分更多的区域

	ClientSocialLinkEmail     uint = 1  // 邮箱
	ClientSocialLinkPhone     uint = 2  // 手机
	ClientSocialLinkQQ        uint = 3  // QQ
	ClientSocialLinkWeChat    uint = 4  // 微信
	ClientSocialLinkWeibo     uint = 5  // 微博
	ClientSocialLinkFacebook  uint = 6  // Facebook
	ClientSocialLinkTwitter   uint = 7  // Twitter
	ClientSocialLinkTelegram  uint = 8  // Telegram
	ClientSocialLinkDiscord   uint = 9  // Discord
	ClientSocialLinkInstagram uint = 10 // Instagram
	ClientSocialLinkYouTube   uint = 11 // YouTube
	ClientSocialLinkTikTok    uint = 12 // TikTok
)

func (c *Client) GetPlatformName() string {
	return GetClientPlatformName(c.Platform)
}

func (c *Client) GetAreaName() string {
	return GetClientAreaName(c.Area)
}

func (c *Client) IsOnline() bool {
	currentTime := time.Now().UnixMilli()
	return (c.OnlineAt > 0 && (c.OnlineAt <= currentTime)) && (c.OfflineAt == 0 || c.OfflineAt > currentTime)
}

func (c *Client) IsOffline() bool {
	currentTime := time.Now().UnixMilli()
	return (c.OfflineAt > 0 && (c.OfflineAt <= currentTime)) && (c.OnlineAt == 0 || c.OnlineAt > currentTime)
}

func (c *Client) IsComingOnline() bool {
	currentTime := time.Now().UnixMilli()
	return c.OnlineAt > currentTime && (c.OfflineAt == -1 || c.OfflineAt < currentTime)
}

func (c *Client) IsComingOffline() bool {
	currentTime := time.Now().UnixMilli()
	return c.OfflineAt > currentTime && (c.OnlineAt == -1 || c.OnlineAt < currentTime)
}

func (c *Client) GetSocialLink(social uint) (string, string) {
	if v, ok := c.GetSocialLinks()[social]; ok {
		return GetClientSocialLinkName(social), v
	}
	return "", ""
}

func (c *Client) GetMarketHome(market uint) (string, string) {
	if v, ok := c.GetMarketHomes()[market]; ok {
		return GetPlatformMarketName(c.Platform, market), v
	}
	return "", ""
}

func (c *Client) GetLatestVersion(market uint) *Version {
	if v, ok := c.Versions[market]; ok {
		return v
	}
	return nil
}

const (
	clientExtraKeyIconUrl     = "iconUrl"     // icon地址
	clientExtraKeySocialLinks = "socialLinks" // 社交链接 (方便控制台跳转)
	clientExtraKeyMarketHomes = "marketHomes" // 应用市场页面 (方便控制台跳转)
	clientExtraKeyIosId       = "iosId"       // apple应用市场id
)

func (c *Client) SetIconUrl(iconUrl *string) {
	c.Extra.SetString(clientExtraKeyIconUrl, iconUrl)
}

func (c *Client) GetIconUrl() string {
	data, _ := c.Extra.GetString(clientExtraKeyIconUrl)
	return data
}

func (c *Client) SetSocialLinks(socialLinks *map[uint]string) int {
	var count int
	if (socialLinks != nil) && (len(*socialLinks) > 0) {
		for k := range *socialLinks {
			ok := c.SetSocialLink(k, (*socialLinks)[k])
			if ok {
				count++
			}
		}
	} else {
		delete(c.Extra, clientExtraKeySocialLinks)
	}
	return count
}

func (c *Client) SetSocialLink(social uint, link string) bool {
	if !IsClientSocialLinkOk(social) {
		return false
	} else if len(link) <= 0 {
		if c.Extra[clientExtraKeySocialLinks] != nil {
			delete((c.Extra[clientExtraKeySocialLinks]).(map[uint]string), social)
		}
		return true
	}
	if c.Extra[clientExtraKeySocialLinks] == nil {
		c.Extra[clientExtraKeySocialLinks] = map[uint]string{}
	}
	(c.Extra[clientExtraKeySocialLinks]).(map[uint]string)[social] = link
	return true
}

func (c *Client) GetSocialLinks() map[uint]string {
	if c.Extra[clientExtraKeySocialLinks] == nil {
		return map[uint]string{}
	}
	return (c.Extra[clientExtraKeySocialLinks]).(map[uint]string)
}

func (c *Client) SetMarketHomes(marketHomes *map[uint]string) int {
	var count int
	if (marketHomes != nil) && (len(*marketHomes) > 0) {
		for k := range *marketHomes {
			ok := c.SetMarketHome(k, (*marketHomes)[k])
			if ok {
				count++
			}
		}
	} else {
		delete(c.Extra, clientExtraKeyMarketHomes)
	}
	return count
}

func (c *Client) SetMarketHome(market uint, home string) bool {
	if !IsPlatformMarketOk(c.Platform, market) {
		return false
	} else if len(home) <= 0 {
		if c.Extra[clientExtraKeyMarketHomes] != nil {
			delete((c.Extra[clientExtraKeyMarketHomes]).(map[uint]string), market)
		}
		return true
	}
	if c.Extra[clientExtraKeyMarketHomes] == nil {
		c.Extra[clientExtraKeyMarketHomes] = map[uint]string{}
	}
	(c.Extra[clientExtraKeyMarketHomes]).(map[uint]string)[market] = home
	return true
}

func (c *Client) GetMarketHomes() map[uint]string {
	if c.Extra[clientExtraKeyMarketHomes] == nil {
		return map[uint]string{}
	}
	return (c.Extra[clientExtraKeyMarketHomes]).(map[uint]string)
}

func (c *Client) SetIosId(iosId *string) {
	c.Extra.SetString(clientExtraKeyIosId, iosId)
}

func (c *Client) GetIosId() string {
	data, _ := c.Extra.GetString(clientExtraKeyIosId)
	return data
}

func IsClientPlatformOk(platform uint) bool {
	switch platform {
	case ClientPlatformLinux,
		ClientPlatformWindows,
		ClientPlatformMacOS,
		ClientPlatformWeb,
		ClientPlatformAndroid,
		ClientPlatformIOS,
		ClientPlatformApplet,
		ClientPlatformTvAnd,
		ClientPlatformTvIOS:
		return true
	}
	return false
}

func IsClientAreaOk(area uint) bool {
	switch area {
	case ClientAreaWord,
		ClientAreaChinaLand,
		ClientAreaChinaHMT,
		ClientAreaEurope:
		return true
	}
	return false
}

func IsClientSocialLinkOk(socialLink uint) bool {
	switch socialLink {
	case ClientSocialLinkEmail,
		ClientSocialLinkPhone,
		ClientSocialLinkQQ,
		ClientSocialLinkWeChat,
		ClientSocialLinkWeibo,
		ClientSocialLinkFacebook,
		ClientSocialLinkTwitter,
		ClientSocialLinkTelegram,
		ClientSocialLinkDiscord,
		ClientSocialLinkInstagram,
		ClientSocialLinkYouTube,
		ClientSocialLinkTikTok:
		return true
	}
	return false
}

var platformInfos = map[uint]string{
	ClientPlatformLinux:   "Linux",
	ClientPlatformWindows: "Windows",
	ClientPlatformMacOS:   "MacOS",
	ClientPlatformWeb:     "Web",
	ClientPlatformAndroid: "Android",
	ClientPlatformIOS:     "IOS",
	ClientPlatformApplet:  "Applet",
	ClientPlatformTvAnd:   "TvAnd",
	ClientPlatformTvIOS:   "TvIOS",
}

var areaInfos = map[uint]string{
	ClientAreaWord:      "Word",
	ClientAreaChinaLand: "ChinaLand",
	ClientAreaChinaHMT:  "ChinaHMT",
	ClientAreaEurope:    "Europe",
}

var socialLinkInfos = map[uint]string{
	ClientSocialLinkEmail:     "Email",
	ClientSocialLinkPhone:     "Phone",
	ClientSocialLinkQQ:        "QQ",
	ClientSocialLinkWeChat:    "WeChat",
	ClientSocialLinkWeibo:     "Weibo",
	ClientSocialLinkFacebook:  "Facebook",
	ClientSocialLinkTwitter:   "Twitter",
	ClientSocialLinkTelegram:  "Telegram",
	ClientSocialLinkDiscord:   "Discord",
	ClientSocialLinkInstagram: "Instagram",
	ClientSocialLinkYouTube:   "YouTube",
	ClientSocialLinkTikTok:    "TikTok",
}

func GetClientPlatformName(ClientPlatform uint) string {
	if v, ok := platformInfos[ClientPlatform]; ok {
		return v
	}
	return ""
}

func GetClientAreaName(ClientArea uint) string {
	if v, ok := areaInfos[ClientArea]; ok {
		return v
	}
	return ""
}

func GetClientSocialLinkName(ClientSocialLink uint) string {
	if v, ok := socialLinkInfos[ClientSocialLink]; ok {
		return v
	}
	return ""
}
