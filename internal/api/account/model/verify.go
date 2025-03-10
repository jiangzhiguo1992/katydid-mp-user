package model

import (
	"katydid-mp-user/internal/pkg/model"
)

type (
	// VerifyInfo 验证内容
	VerifyInfo struct {
		*model.Base

		AccountId uint64    `json:"accountId"` // 账户Id
		OwnKind   VerifyOwn `json:"ownType"`   // 验证平台 (组织/应用)
		OwnID     uint64    `json:"ownId"`     // 认证拥有者Id (组织/应用)

		AuthKind AuthKind    `json:"kind"`      // 认证类型 (手机号/邮箱/...)
		Apply    VerifyApply `json:"applyKind"` // 申请类型 (注册/登录/修改密码/...)

		PendingAt  *int64 `json:"pendingAt"`  // 等待时间
		VerifiedAt *int64 `json:"verifiedAt"` // 验证时间
		ExpiresAt  *int64 `json:"expiresAt"`  // 过期时间
		Attempts   int    `json:"attempts"`   // 验证次数
	}

	// VerifyOwn 认证拥有者
	VerifyOwn int8

	// VerifyApply 申请类型
	VerifyApply int16
)

func NewVerifyInfoEmpty() *VerifyInfo {
	return &VerifyInfo{Base: model.NewBaseEmpty()}
}

func NewVerifyInfoDef(
	accountId uint64, ownKind VerifyOwn, ownId uint64,
	authKind AuthKind, apply VerifyApply,
) *VerifyInfo {
	return &VerifyInfo{
		Base:      model.NewBaseEmpty(),
		AccountId: accountId, OwnKind: ownKind, OwnID: ownId,
		AuthKind: authKind, Apply: apply,
		PendingAt: nil, VerifiedAt: nil, ExpiresAt: nil, Attempts: 0,
	}
}

const (
	VerifyStatusInit    model.Status = 0 // 初始状态
	VerifyStatusPending model.Status = 1 // 等待验证
	VerifyStatusReject  model.Status = 2 // 验证失败
	VerifyStatusSuccess model.Status = 3 // 验证成功
)

const (
	VerifyOwnOrg    VerifyOwn = 1 // 组织类型
	VerifyOwnApp    VerifyOwn = 2 // 应用类型
	VerifyOwnClient VerifyOwn = 3 // 应用类型
)

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

const (
	verifyExtraKeyBody = "body" // 验证内容
)

func (v *VerifyInfo) SetBody(body *string) {
	v.Extra.SetString(verifyExtraKeyBody, body)
}

func (v *VerifyInfo) GetBody() (string, bool) {
	return v.Extra.GetString(verifyExtraKeyBody)
}
