package model

import (
	"katydid-mp-user/internal/pkg/model"
	"time"
)

// Version 应用版本
type Version struct {
	*model.Base
	ClientId uint64 `json:"clientId" validate:"required"` // 客户端平台id
	Market   uint   `json:"market" validate:"required"`   // 市场/渠道/投放
	Code     uint   `json:"code" validate:"required"`     // 版本号

	Enable    bool  `json:"enable"`    // 是否可用 (是否可下载+可使用+是否对用户可见?)
	Force     bool  `json:"force"`     // 是否强制升级
	BuildAt   int64 `json:"buildAt"`   // 构建时间 (一般是上传时间)
	PublishAt int64 `json:"publishAt"` // 发布时间 (审核通过后发布时间，是否对用户可见?)

	AppSecret string `json:"appSecret"` // app密钥 (终端使用，版本更替，确定后不可改) TODO:GG 不会返回给客户端，在网关/代理层处理?
}

func NewVersionEmpty() *Version {
	return &Version{
		Base: model.NewBaseEmpty(),
	}
}

func NewVersionDefault(
	clientId uint64, market uint, code uint,
	enable bool, force bool, buildAt, publishAt int64,
	appSecret string,
) *Version {
	return &Version{
		Base:     model.NewBaseEmpty(),
		ClientId: clientId, Market: market, Code: code,
		Enable: enable, Force: force, BuildAt: buildAt, PublishAt: publishAt,
		AppSecret: appSecret,
	}
}

const (
	VersionMarketLinuxOfficial uint = 1
	VersionMarketLinuxSteam    uint = 2

	VersionMarketWindowsOfficial uint = 1
	VersionMarketWindowsSteam    uint = 2

	VersionMarketMacOsOfficial uint = 1
	VersionMarketMacOsAppStore uint = 2
	VersionMarketMacOsSteam    uint = 3

	VersionMarketWebOfficial uint = 1
	VersionMarketWebChrome   uint = 2
	VersionMarketWebSafari   uint = 3
	VersionMarketWebFirefox  uint = 4
	VersionMarketWebEdge     uint = 5
	VersionMarketWebOpera    uint = 6
	VersionMarketWebIE       uint = 7
	VersionMarketWeb360      uint = 8
	VersionMarketWebQQ       uint = 9
	VersionMarketWebHuoHu    uint = 10
	VersionMarketWebLieBao   uint = 11
	VersionMarketWebSouGou   uint = 12

	VersionMarketAndroidOfficial   uint = 1
	VersionMarketAndroidGooglePlay uint = 2
	VersionMarketAndroidTapTap     uint = 3
	VersionMarketAndroidHuawei     uint = 4
	VersionMarketAndroidXiaomi     uint = 5
	VersionMarketAndroidOppo       uint = 6
	VersionMarketAndroidVivo       uint = 7
	VersionMarketAndroidMeizu      uint = 8
	VersionMarketAndroidOnePlus    uint = 9
	VersionMarketAndroidSamsung    uint = 10
	VersionMarketAndroidLenovo     uint = 11
	VersionMarketAndroidSony       uint = 12
	VersionMarketAndroidLG         uint = 13
	VersionMarketAndroidHTC        uint = 14
	VersionMarketAndroidMotorola   uint = 15
	VersionMarketAndroidNokia      uint = 16
	VersionMarketAndroidTencent    uint = 30
	VersionMarketAndroidBaidu      uint = 31
	VersionMarketAndroid360        uint = 32

	VersionMarketIOSOfficial uint = 1
	VersionMarketIOSAppStore uint = 2

	VersionMarketAppletOfficial    uint = 1
	VersionMarketAppletWeChat      uint = 2
	VersionMarketAppletQQ          uint = 3
	VersionMarketAppletDouYin      uint = 4
	VersionMarketAppletKuaiShou    uint = 5
	VersionMarketAppletXiaoHongShu uint = 6
	VersionMarketAppletBaidu       uint = 7
	VersionMarketAppletZhiFuBao    uint = 8
	VersionMarketAppletTaoBao      uint = 9
	VersionMarketAppletJingDong    uint = 10
	VersionMarketAppletDingDing    uint = 11

	//ItchIo     uint = 2
	//KongReGate uint = 3
	//IndieDb    uint = 4

	VersionMarketAdsMin uint = 1000 // 大于这个的都是投量广告
)

func (v *Version) IsBuild() bool {
	return v.BuildAt > time.Now().UnixMilli()
}

func (v *Version) IsPublish() bool {
	return v.PublishAt > time.Now().UnixMilli()
}

const (
	versionExtraKeyName       = "name"       // 版本名
	versionExtraKeyUrl        = "url"        // 下载地址
	versionExtraKeyLog        = "log"        // 更新日志
	versionExtraKeySize       = "size"       // 安装包大小
	versionExtraKeyCompact    = "compact"    // 兼容信息
	versionExtraKeyIconUrl    = "iconUrl"    // icon地址
	versionExtraKeyImgUrls    = "imgUrls"    // 版本介绍图片地址
	versionExtraKeyVideoUrls  = "videoUrls"  // 版本介绍视频地址
	versionExtraKeyMarketName = "marketName" // 设置广告渠道名称 (广告投放)
)

func (v *Version) SetName(name *string) {
	v.Extra.SetString(versionExtraKeyName, name)
}

func (v *Version) GetName() string {
	data, _ := v.Extra.GetString(versionExtraKeyName)
	return data
}

func (v *Version) SetUrl(url *string) {
	v.Extra.SetString(versionExtraKeyUrl, url)
}

func (v *Version) GetUrl() string {
	data, _ := v.Extra.GetString(versionExtraKeyUrl)
	return data
}

func (v *Version) SetLog(log *string) {
	v.Extra.SetString(versionExtraKeyLog, log)
}

func (v *Version) GetLog() string {
	data, _ := v.Extra.GetString(versionExtraKeyLog)
	return data
}

func (v *Version) SetSize(size *int64) {
	v.Extra.SetInt64(versionExtraKeySize, size)
}

func (v *Version) GetSize() int64 {
	data, _ := v.Extra.GetInt64(versionExtraKeySize)
	return data
}

func (v *Version) SetCompact(compact *string) {
	v.Extra.SetString(versionExtraKeyCompact, compact)
}

func (v *Version) GetCompact() string {
	data, _ := v.Extra.GetString(versionExtraKeyCompact)
	return data
}

func (v *Version) SetIconUrl(iconUrl *string) {
	v.Extra.SetString(versionExtraKeyIconUrl, iconUrl)
}

func (v *Version) GetIconUrl() string {
	data, _ := v.Extra.GetString(versionExtraKeyIconUrl)
	return data
}

func (v *Version) SetImgUrls(imgUrls *[]string) {
	v.Extra.SetStringSlice(versionExtraKeyImgUrls, imgUrls)
}

func (v *Version) GetImgUrls() []string {
	data, _ := v.Extra.GetStringSlice(versionExtraKeyImgUrls)
	return data
}

func (v *Version) SetVideoUrls(videoUrls *[]string) {
	v.Extra.SetStringSlice(versionExtraKeyVideoUrls, videoUrls)
}

func (v *Version) GetVideoUrls() []string {
	data, _ := v.Extra.GetStringSlice(versionExtraKeyVideoUrls)
	return data
}

func (v *Version) SetMarketName(marketName *string) {
	if v.Market < VersionMarketAdsMin {
		return
	}
	v.Extra.SetString(versionExtraKeyMarketName, marketName)
}

func (v *Version) GetMarketName(platform uint) string {
	if v.Market < VersionMarketAdsMin {
		return GetPlatformMarketName(platform, v.Market)
	}
	data, _ := v.Extra.GetString(versionExtraKeyMarketName)
	return data
}

var platformMarketInfos = map[uint]map[uint]string{
	ClientPlatformLinux: {
		VersionMarketLinuxOfficial: "Linux_官网",
		VersionMarketLinuxSteam:    "Linux_Steam",
	},
	ClientPlatformWindows: {
		VersionMarketWindowsOfficial: "Windows_官网",
		VersionMarketWindowsSteam:    "Windows_Steam",
	},
	ClientPlatformMacOS: {
		VersionMarketMacOsOfficial: "MacOS_官网",
		VersionMarketMacOsAppStore: "MacOS_应用商店",
		VersionMarketMacOsSteam:    "MacOS_Steam",
	},
	ClientPlatformWeb: {
		VersionMarketWebOfficial: "Web_官网",
		VersionMarketWebChrome:   "Web_Chrome",
		VersionMarketWebSafari:   "Web_Safari",
		VersionMarketWebFirefox:  "Web_Firefox",
		VersionMarketWebEdge:     "Web_Edge",
		VersionMarketWebOpera:    "Web_Opera",
		VersionMarketWebIE:       "Web_IE",
		VersionMarketWeb360:      "Web_360",
		VersionMarketWebQQ:       "Web_QQ",
		VersionMarketWebHuoHu:    "Web_火狐",
		VersionMarketWebLieBao:   "Web_猎豹",
		VersionMarketWebSouGou:   "Web_搜狗",
	},
	ClientPlatformAndroid: {
		VersionMarketAndroidOfficial:   "Android_官网",
		VersionMarketAndroidGooglePlay: "Android_谷歌",
		VersionMarketAndroidTapTap:     "Android_TapTap",
		VersionMarketAndroidHuawei:     "Android_华为",
		VersionMarketAndroidXiaomi:     "Android_小米",
		VersionMarketAndroidOppo:       "Android_oppo",
		VersionMarketAndroidVivo:       "Android_vivo",
		VersionMarketAndroidMeizu:      "Android_魅族",
		VersionMarketAndroidOnePlus:    "Android_一加",
		VersionMarketAndroidSamsung:    "Android_三星",
		VersionMarketAndroidLenovo:     "Android_联想",
		VersionMarketAndroidSony:       "Android_索尼",
		VersionMarketAndroidLG:         "Android_LG",
		VersionMarketAndroidHTC:        "Android_HTC",
		VersionMarketAndroidMotorola:   "Android_摩托罗拉",
		VersionMarketAndroidNokia:      "Android_诺基亚",
		VersionMarketAndroidTencent:    "Android_腾讯",
		VersionMarketAndroidBaidu:      "Android_百度",
		VersionMarketAndroid360:        "Android_360",
	},
	ClientPlatformIOS: {
		VersionMarketIOSOfficial: "IOS_官网",
		VersionMarketIOSAppStore: "IOS_AppStore",
	},
	ClientPlatformApplet: {
		VersionMarketAppletOfficial:    "Applet_官网",
		VersionMarketAppletWeChat:      "Applet_微信",
		VersionMarketAppletQQ:          "Applet_QQ",
		VersionMarketAppletDouYin:      "Applet_抖音",
		VersionMarketAppletKuaiShou:    "Applet_快手",
		VersionMarketAppletXiaoHongShu: "Applet_小红书",
		VersionMarketAppletBaidu:       "Applet_百度",
		VersionMarketAppletZhiFuBao:    "Applet_支付宝",
		VersionMarketAppletTaoBao:      "Applet_淘宝",
		VersionMarketAppletJingDong:    "Applet_京东",
		VersionMarketAppletDingDing:    "Applet_钉钉",
	},
}

func IsPlatformMarketOk(platform, market uint) bool {
	if _, ok := platformMarketInfos[platform]; !ok {
		return false
	}
	if market < VersionMarketAdsMin {
		if _, ok := platformMarketInfos[platform][market]; !ok {
			return false
		}
	}
	return true
}

func GetPlatformMarketName(platform, market uint) string {
	if _, ok := platformMarketInfos[platform]; !ok {
		return ""
	}
	if market < VersionMarketAdsMin {
		if _, ok := platformMarketInfos[platform][market]; !ok {
			return ""
		}
		return platformMarketInfos[platform][market]
	}
	return ""
}
