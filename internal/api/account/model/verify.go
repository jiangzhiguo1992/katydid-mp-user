package model

import (
	"katydid-mp-user/internal/pkg/model"
)

const (
	VerifyStatusInit    model.Status = 0 // 初始状态
	VerifyStatusPending model.Status = 1 // 等待验证
	VerifyStatusSuccess model.Status = 2 // 验证成功
	VerifyStatusReject  model.Status = 3 // 验证失败
)

type VerifyKind uint16

const (
	VerifyKindPwd   VerifyKind = 1 // 密码
	VerifyKindPhone VerifyKind = 2 // 短信
	VerifyKindEmail VerifyKind = 3 // 邮箱
	VerifyKindBio   VerifyKind = 4 // 生物特征
	VerifyKindThird VerifyKind = 5 // 第三方平台
)

type VerifyApply int16

const (
	VerifyApplyUnregister  VerifyApply = -1 // 注销
	VerifyApplyRegister    VerifyApply = 1  // 注册
	VerifyApplyLogin       VerifyApply = 2  // 登录
	VerifyApplyResetPwd    VerifyApply = 3  // 重置密码
	VerifyApplyChangePhone VerifyApply = 4  // 修改手机号
	VerifyApplyChangeEmail VerifyApply = 5  // 修改邮箱
	VerifyApplyChangeBio   VerifyApply = 6  // 修改生物特征
	VerifyApplyChangeThird VerifyApply = 7  // 修改第三方平台
)

// VerifyOwn 认证拥有者
type VerifyOwn int8

const (
	VerifyOwnOrg    VerifyOwn = 1 // 组织类型
	VerifyOwnApp    VerifyOwn = 2 // 应用类型
	VerifyOwnClient VerifyOwn = 3 // 应用类型
)

// VerifyInfo 验证内容
type VerifyInfo struct {
	*model.Base
	AccountId uint64    `json:"accountId"` // 账户Id
	OwnKind   VerifyOwn `json:"ownType"`   // 认证方式 (手机号/邮箱/...)
	OwnID     uint64    `json:"ownId"`     // 认证拥有者Id (组织/应用)

	AuthKind  AuthKind    `json:"kind"`      // 平台 (手机号/邮箱/...)
	ApplyKind VerifyApply `json:"applyKind"` // 申请类型 (注册/登录/修改密码/...)

	State      int8  `json:"state"`      // 验证状态
	PendingAt  int64 `json:"pendingAt"`  // 等待时间
	VerifiedAt int64 `json:"verifiedAt"` // 验证时间
	ExpiresAt  int64 `json:"expiresAt"`  // 过期时间
	Attempts   int8  `json:"attempts"`   // 验证次数

	// TODO:GG ip, device, location
}

func NewVerifyInfoEmpty() *VerifyInfo {
	return &VerifyInfo{Base: model.NewBaseEmpty()}
}

func NewVerifyInfoDef(clientId, accountId uint64, kind int16) *VerifyInfo {
	return &VerifyInfo{
		Base:       model.NewBaseEmpty(),
		ClientId:   clientId,
		AccountId:  accountId,
		Kind:       kind,
		State:      VerityStateInit,
		PendingAt:  -1,
		VerifiedAt: -1,
		ExpiresAt:  -1,
		Attempts:   0,
	}
}

// WithBody 设置验证内容 (注意language)
func (v *VerifyInfo) WithBody(data *string) map[string]any {
	datas := v.Extra
	if (data != nil) && (len(*data) > 0) {
		datas["body"] = *data
	} else {
		delete(datas, "body")
	}
	return datas
}

func (v *VerifyInfo) GetBody() string {
	if v.Extra["body"] == nil {
		return ""
	}
	return v.Extra["body"].(string)
}

// WithCode 设置验证码
func (v *VerifyInfo) WithCode(data *string) map[string]any {
	datas := v.Extra
	if (data != nil) && (len(*data) > 0) {
		datas["code"] = *data
	} else {
		delete(datas, "code")
	}
	return datas
}

func (v *VerifyInfo) GetCode() string {
	if v.Extra["code"] == nil {
		return ""
	}
	return v.Extra["code"].(string)
}
