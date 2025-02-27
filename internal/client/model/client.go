package model

import "katydid-mp-user/internal/pkg/model"

// import (
//
//	"fmt"
//	"katydid_base_api/internal/pkg/base"
//	"katydid_base_api/internal/pkg/utils"
//	"katydid_base_api/tools"
//	"time"
//
// )
//

const (
	versionExtraKeyCert = "cert"
)

// Client 客户端平台
type Client struct {
	*model.Base
	ClientId uint64 `json:"clientId"` // 客户端id
	Platform uint16 `json:"platform"` // 平台
	Area     uint   `json:"area"`     // 区域编号

	Enable    bool   `json:"enable"`    // 是否可用 (一般不用，下架之类的，没有reason)
	OnlineAt  int64  `json:"onlineAt"`  // 上线时间 (时间没到时，只能停留在首页，提示bulletins)
	OfflineAt int64  `json:"offlineAt"` // 下线时间 (时间到后，强制下线+升级/等待/...)
	AppId     string `json:"appId"`     // 各平台应用唯一标识 (pkg/bundle，海外和大陆可以同时安装!)
	AppName   string `json:"appName"`   // app名称 (不同Area的区别)

	Versions map[uint]*Version `json:"versions" gorm:"-:all"` // [market]最新publish版本号
}

//
//func NewPlatformEmpty() *Client {
//	return &Client{DBModel: &base.DBModel{}}
//}
//
//func NewPlatformDefault(
//	clientId uint64, platform uint16, area uint,
//	enable bool,
//	appId string, appName string,
//) *Client {
//	return &Client{
//		DBModel:  base.NewDBModelEmpty(),
//		ClientId: clientId, Client: platform, Area: area,
//		Enable: enable, OnlineAt: 0, OfflineAt: 0,
//		AppId: appId, AppName: appName,
//		Extra:    map[string]interface{}{},
//		Versions: make(map[uint]*Version),
//	}
//}
//
//// IsOnline 是否上线
//func (p *Client) IsOnline() bool {
//	currentTime := time.Now().UnixMilli()
//	return (p.OnlineAt > 0 && (p.OnlineAt <= currentTime)) && (p.OfflineAt == 0 || p.OfflineAt > currentTime)
//}
//
//// IsOffline 是否下线
//func (p *Client) IsOffline() bool {
//	currentTime := time.Now().UnixMilli()
//	return (p.OfflineAt > 0 && (p.OfflineAt <= currentTime)) && (p.OnlineAt == 0 || p.OnlineAt > currentTime)
//}
//
//// IsComingOnline 是否即将上线
//func (p *Client) IsComingOnline() bool {
//	currentTime := time.Now().UnixMilli()
//	return p.OnlineAt > currentTime && (p.OfflineAt == -1 || p.OfflineAt < currentTime)
//}
//
//// IsComingOffline 是否即将下线
//func (p *Client) IsComingOffline() bool {
//	currentTime := time.Now().UnixMilli()
//	return p.OfflineAt > currentTime && (p.OnlineAt == -1 || p.OnlineAt < currentTime)
//}
//
//// SetSocialLinks 社交链接 (方便控制台跳转)
//func (p *Client) SetSocialLinks(socialLinks *map[uint16]string) int {
//	var count int
//	if (socialLinks != nil) && (len(*socialLinks) > 0) {
//		for k := range *socialLinks {
//			ok := p.SetSocialLink(k, (*socialLinks)[k])
//			if ok {
//				count++
//			}
//		}
//	} else {
//		delete(p.Extra, "socialLinks")
//	}
//	return count
//}
//
//func (p *Client) SetSocialLink(social uint16, link string) bool {
//	if !isSocialLinkTypeOk(social) {
//		return false
//	} else if len(link) <= 0 {
//		if p.Extra["socialLinks"] != nil {
//			delete((p.Extra["socialLinks"]).(map[uint16]string), social)
//		}
//		return true
//	}
//	if p.Extra["socialLinks"] == nil {
//		p.Extra["socialLinks"] = map[uint16]string{}
//	}
//	(p.Extra["socialLinks"]).(map[uint16]string)[social] = link
//	return true
//}
//
//func (p *Client) GetSocialLinks() map[uint16]string {
//	if p.Extra["socialLinks"] == nil {
//		return map[uint16]string{}
//	}
//	return (p.Extra["socialLinks"]).(map[uint16]string)
//}
//
//func (p *Client) GetSocialLink(social uint16) (string, string) {
//	if v, ok := p.GetSocialLinks()[social]; ok {
//		return socialLinkName(social), v
//	}
//	return "", ""
//}
//
//// SetMarketHomes 应用市场页面 (方便控制台跳转)
//func (p *Client) SetMarketHomes(marketHomes *map[uint]string) int {
//	var count int
//	if (marketHomes != nil) && (len(*marketHomes) > 0) {
//		for k := range *marketHomes {
//			ok := p.SetMarketHome(k, (*marketHomes)[k])
//			if ok {
//				count++
//			}
//		}
//	} else {
//		delete(p.Extra, "marketHomes")
//	}
//	return count
//}
//
//func (p *Client) SetMarketHome(market uint, home string) bool {
//	if !isPlatformMarketTypeOk(p.Client, market) {
//		return false
//	} else if len(home) <= 0 {
//		if p.Extra["marketHomes"] != nil {
//			delete((p.Extra["marketHomes"]).(map[uint]string), market)
//		}
//		return true
//	}
//	if p.Extra["marketHomes"] == nil {
//		p.Extra["marketHomes"] = map[uint]string{}
//	}
//	(p.Extra["marketHomes"]).(map[uint]string)[market] = home
//	return true
//}
//
//func (p *Client) GetMarketHomes() map[uint]string {
//	if p.Extra["marketHomes"] == nil {
//		return map[uint]string{}
//	}
//	return (p.Extra["marketHomes"]).(map[uint]string)
//}
//
//func (p *Client) GetMarketHome(market uint) (string, string) {
//	if v, ok := p.GetMarketHomes()[market]; ok {
//		return platformMarketName(p.Client, market), v
//	}
//	return "", ""
//}
//
//// SetIosId apple应用市场id
//func (p *Client) SetIosId(iosId *string) {
//	if (iosId != nil) && (len(*iosId) > 0) {
//		p.Extra["iosId"] = *iosId
//	} else {
//		delete(p.Extra, "iosId")
//	}
//}
//
//func (p *Client) GetIosId() string {
//	if p.Extra["iosId"] == nil {
//		return ""
//	}
//	return p.Extra["iosId"].(string)
//}
//
//func (p *Client) GetLatestVersion(market uint) *Version {
//	if v, ok := p.Versions[market]; ok {
//		return v
//	}
//	return nil
//}
//
//const (
//	checkClientPlatformAppIdLen   = 100
//	checkClientPlatformAppNameLen = 100
//
//	checkClientPlatformSocialLinksNum = 100
//	checkClientPlatformSocialLinkLen  = 500
//	checkClientPlatformMarketHomesNum = 100
//	checkClientPlatformMarketHomeLen  = 500
//	checkClientPlatformIosIdLen       = 50
//)
//
//// CheckFields 检查字段
//func (p *Client) CheckFields() []*tools.CodeError {
//	var errors []*tools.CodeError
//	if !isPlatformTypeOk(p.Client) {
//		errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldRange).WithPrefix("Client"))
//	}
//	if !isAreaTypeOk(p.Area) {
//		errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldRange).WithPrefix("Area"))
//	}
//	if len(p.AppId) <= 0 {
//		errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldNil).WithPrefix("AppId"))
//	} else if len(p.AppId) > checkClientPlatformAppIdLen {
//		errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldLarge).WithPrefix("AppId"))
//	}
//	if len(p.AppName) <= 0 {
//		errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldNil).WithPrefix("AppName"))
//	} else if len(p.AppName) > checkClientPlatformAppNameLen {
//		errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldLarge).WithPrefix("AppName"))
//	}
//	if len(p.Extra) > base.ExtraMaxCount {
//		errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldMax).WithPrefix("Extra"))
//	}
//	for k, v := range p.Extra {
//		switch k {
//		case "socialLinks":
//			if len(v.(map[uint16]string)) > checkClientPlatformSocialLinksNum {
//				errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldMax).WithPrefix(k))
//			}
//			for kk, vv := range v.(map[uint16]string) {
//				if len(vv) > checkClientPlatformSocialLinkLen {
//					errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldLarge).WithPrefix(fmt.Sprintf("%s[%d] ", k, kk)))
//				}
//			}
//		case "marketHomes":
//			if len(v.(map[uint]string)) > checkClientPlatformMarketHomesNum {
//				errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldMax).WithPrefix(k))
//			}
//			for kk, vv := range v.(map[uint]string) {
//				if len(vv) > checkClientPlatformMarketHomeLen {
//					errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldLarge).WithPrefix(fmt.Sprintf("%s[%d] ", k, kk)))
//				}
//			}
//		case "iosId":
//			if len(v.(string)) > checkClientPlatformIosIdLen {
//				errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldLarge).WithPrefix(k))
//			}
//		default:
//			if len(v.(string)) > base.ExtraItemMaxLen {
//				errors = append(errors, utils.MatchErrorByCode(utils.ErrorCodeDBFieldLarge).WithPrefix(fmt.Sprintf("Extra[%s]", k)))
//			}
//		}
//	}
//	return errors
//}
//
//func (p *Client) GetPlatformName() string {
//	return platformName(p.Client)
//}
//
//func (p *Client) GetAreaName() string {
//	return areaName(p.Area)
//}
//
//const (
//	PlatformTypeLinux   uint16 = 1
//	PlatformTypeWindows uint16 = 2
//	PlatformTypeMacOS   uint16 = 3
//	PlatformTypeWeb     uint16 = 4
//	PlatformTypeAndroid uint16 = 5
//	PlatformTypeIOS     uint16 = 6
//	PlatformTypeApplet  uint16 = 7
//	//PlatformTypeTvAnd   uint16 = 8
//	//PlatformTypeTvIOS   uint16 = 9
//)
//
//const (
//	AreaTypeWord      uint = 1 // 全球 (默认英文，泛指海外)
//	AreaTypeChinaLand uint = 2 // 中国大陆 (简体中文)
//	AreaTypeChinaHMT  uint = 3 // 中国港澳台 (繁体中文)
//	AreaTypeEurope    uint = 4 // 欧洲 (默认英文，GDPR)
//	// 后面可能会划分更多的区域
//)
//
//const (
//	SocialLinkTypeEmail     uint16 = 1
//	SocialLinkTypePhone     uint16 = 2
//	SocialLinkTypeQQ        uint16 = 3
//	SocialLinkTypeWeChat    uint16 = 4
//	SocialLinkTypeWeibo     uint16 = 5
//	SocialLinkTypeFacebook  uint16 = 6
//	SocialLinkTypeTwitter   uint16 = 7
//	SocialLinkTypeTelegram  uint16 = 8
//	SocialLinkTypeDiscord   uint16 = 9
//	SocialLinkTypeInstagram uint16 = 10
//	SocialLinkTypeYouTube   uint16 = 11
//	SocialLinkTypeTikTok    uint16 = 12
//)
//
//func isPlatformTypeOk(platformType uint16) bool {
//	switch platformType {
//	case PlatformTypeLinux,
//		PlatformTypeWindows,
//		PlatformTypeMacOS,
//		PlatformTypeWeb,
//		PlatformTypeAndroid,
//		PlatformTypeIOS,
//		PlatformTypeApplet:
//		//PlatformTypeTvAnd,
//		//PlatformTypeTvIOS:
//		return true
//	}
//	return false
//}
//
//func isAreaTypeOk(areaType uint) bool {
//	switch areaType {
//	case AreaTypeWord,
//		AreaTypeChinaLand,
//		AreaTypeChinaHMT,
//		AreaTypeEurope:
//		return true
//	}
//	return false
//}
//
//func isSocialLinkTypeOk(socialLinkType uint16) bool {
//	switch socialLinkType {
//	case SocialLinkTypeEmail,
//		SocialLinkTypePhone,
//		SocialLinkTypeQQ,
//		SocialLinkTypeWeChat,
//		SocialLinkTypeWeibo,
//		SocialLinkTypeFacebook,
//		SocialLinkTypeTwitter,
//		SocialLinkTypeTelegram,
//		SocialLinkTypeDiscord,
//		SocialLinkTypeInstagram,
//		SocialLinkTypeYouTube,
//		SocialLinkTypeTikTok:
//		return true
//	}
//	return false
//}
//
//var platformInfos = map[uint16]string{
//	PlatformTypeLinux:   "Linux",
//	PlatformTypeWindows: "Windows",
//	PlatformTypeMacOS:   "MacOS",
//	PlatformTypeWeb:     "Web",
//	PlatformTypeAndroid: "Android",
//	PlatformTypeIOS:     "IOS",
//	PlatformTypeApplet:  "Applet",
//	//PlatformTypeTvAnd: "TvAnd",
//	//PlatformTypeTvIOS: "TvIOS",
//}
//
//var areaInfos = map[uint]string{
//	AreaTypeWord:      "Word",
//	AreaTypeChinaLand: "ChinaLand",
//	AreaTypeChinaHMT:  "ChinaHMT",
//	AreaTypeEurope:    "Europe",
//}
//
//var socialLinkInfos = map[uint16]string{
//	SocialLinkTypeEmail:     "Email",
//	SocialLinkTypePhone:     "Phone",
//	SocialLinkTypeQQ:        "QQ",
//	SocialLinkTypeWeChat:    "WeChat",
//	SocialLinkTypeWeibo:     "Weibo",
//	SocialLinkTypeFacebook:  "Facebook",
//	SocialLinkTypeTwitter:   "Twitter",
//	SocialLinkTypeTelegram:  "Telegram",
//	SocialLinkTypeDiscord:   "Discord",
//	SocialLinkTypeInstagram: "Instagram",
//	SocialLinkTypeYouTube:   "YouTube",
//	SocialLinkTypeTikTok:    "TikTok",
//}
//
//func platformName(platformType uint16) string {
//	if v, ok := platformInfos[platformType]; ok {
//		return v
//	}
//	return ""
//}
//
//func areaName(areaType uint) string {
//	if v, ok := areaInfos[areaType]; ok {
//		return v
//	}
//	return ""
//}
//
//func socialLinkName(socialLinkType uint16) string {
//	if v, ok := socialLinkInfos[socialLinkType]; ok {
//		return v
//	}
//	return ""
//}
