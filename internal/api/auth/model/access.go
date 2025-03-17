package model

import "katydid-mp-user/internal/pkg/model"

type (
	// Access 访问记录
	Access struct {
		*model.Base

		Kind AccessKind

		OwnKind   OwnKind
		OwnID     uint64
		DeviceID  string
		AccountID uint64

		UserID *uint64
		RoleID *uint64

		// ip
		// time
		// TODO:GG isLogin

		// TODO:GG 不信任的设备登录是，需要refresh JWTToken? client定

		// TODO:GG 小游戏和单机，都是不需要account的，只有device？

		// TODO:GG ip, device, location
		// 		UserId     int64  `json:"userId"`     // --
		//		DeviceId   string `json:"deviceId"`   // 设备标识，用于身份验证
		//		DeviceName string `json:"deviceName"` // 设备名称+型号，用于兼容设备
		//		Market     string `json:"market"`     // 渠道，用于统计下载来源
		//		Language   string `json:"language"`   // 语言
		//		Platform   string `json:"platform"`   // WeChat，IOS，Android，用于统计平台
		//		OsVersion  string `json:"osVersion"`  // weChat/android/ios版本，用于统计兼容版本
		//		AppVersion int    `json:"appVersion"` // 软件versionCode，用于统计升级率，409 低版本升级
	}

	// AccessKind 访问类型
	AccessKind int8
)

const (
	EntryKindLogin  AccessKind = 1 // 登录
	EntryKindStart  AccessKind = 2 // 启动(重新打开app)
	EntryKindWake   AccessKind = 3 // 唤醒(后台切前台)
	EntryKindAccess AccessKind = 4 // 访问(api)
	EntryKindLogout AccessKind = 5 // 登出
)
