package model

import (
	"katydid-mp-user/internal/pkg/model"
	"time"
)

type (
	// Verify 验证内容
	Verify struct {
		*model.Base

		AuthId  uint64    `json:"authId"`  // 认证Id
		OwnKind VerifyOwn `json:"ownType"` // 验证平台 (组织/应用)
		OwnID   uint64    `json:"ownId"`   // 认证拥有者Id (组织/应用)

		AuthKind AuthKind    `json:"kind"`      // 认证类型 (手机号/邮箱/...)
		Apply    VerifyApply `json:"applyKind"` // 申请类型 (注册/登录/修改密码/...)

		PendingAt  *int64 `json:"pendingAt"`  // 等待时间
		ExpiresAt  *int64 `json:"expiresAt"`  // 过期时间
		VerifiedAt *int64 `json:"verifiedAt"` // 验证时间
		Attempts   int    `json:"attempts"`   // 验证次数
	}

	// VerifyOwn 验证平台
	VerifyOwn int8

	// VerifyApply 申请类型
	VerifyApply int16
)

func NewVerifyEmpty() *Verify {
	return &Verify{Base: model.NewBaseEmpty()}
}

func NewVerify(
	authId uint64, ownKind VerifyOwn, ownId uint64,
	authKind AuthKind, apply VerifyApply,
) *Verify {
	return &Verify{
		Base:   model.NewBaseEmpty(),
		AuthId: authId, OwnKind: ownKind, OwnID: ownId,
		AuthKind: authKind, Apply: apply,
		PendingAt: nil, ExpiresAt: nil, VerifiedAt: nil, Attempts: 0,
	}
}

const (
	VerifyStatusInit    model.Status = 0 // 初始状态
	VerifyStatusPending model.Status = 1 // 等待验证
	VerifyStatusReject  model.Status = 2 // 验证失败
	VerifyStatusSuccess model.Status = 3 // 验证成功

	VerifyOwnOrg    VerifyOwn = 1 // 组织类型
	VerifyOwnApp    VerifyOwn = 2 // 应用类型
	VerifyOwnClient VerifyOwn = 3 // 应用类型

	VerifyApplyUnregister  VerifyApply = -1 // 注销
	VerifyApplyRegister    VerifyApply = 1  // 注册
	VerifyApplyLogin       VerifyApply = 2  // 登录
	VerifyApplyResetPwd    VerifyApply = 3  // 重置密码
	VerifyApplyChangePhone VerifyApply = 4  // 修改手机号
	VerifyApplyChangeEmail VerifyApply = 5  // 修改邮箱
	VerifyApplyChangeBio   VerifyApply = 6  // 修改生物特征
	VerifyApplyChangeThird VerifyApply = 7  // 修改第三方平台
)

// IsExpired 检查验证是否已过期
func (v *Verify) IsExpired() bool {
	now := time.Now().Unix()
	return (v.ExpiresAt != nil) && (*v.ExpiresAt <= now)
}

// IsPending 检查是否处于等待验证状态
func (v *Verify) IsPending() bool {
	return v.Status == VerifyStatusPending && !v.IsExpired()
}

// IsVerified 检查是否已验证成功
func (v *Verify) IsVerified() bool {
	return v.Status == VerifyStatusSuccess && v.VerifiedAt != nil
}

// CanResend 检查是否可以重新发送验证码
func (v *Verify) CanResend(cooldownS int64) bool {
	if v.PendingAt == nil {
		return true
	}
	now := time.Now().Unix()
	return (now - *v.PendingAt) >= cooldownS
}

// RemainingAttempts 获取剩余尝试次数
func (v *Verify) RemainingAttempts() int {
	maxAttempts := v.GetMaxAttempts()
	if v.Attempts >= maxAttempts {
		return 0
	}
	return maxAttempts - v.Attempts
}

const (
	verifyExtraKeyBody        = "body"        // 验证内容
	verifyExtraKeyMaxAttempts = "maxAttempts" // 最大尝试次数
	verifyExtraKeyTrySends    = "trySends"    // 尝试发送次数
)

func (v *Verify) SetBody(body *string) {
	v.Extra.SetString(verifyExtraKeyBody, body)
}

func (v *Verify) GetBody() (string, bool) {
	return v.Extra.GetString(verifyExtraKeyBody)
}

// SetMaxAttempts 设置最大尝试次数
func (v *Verify) SetMaxAttempts(attempts *int) {
	v.Extra.SetInt(verifyExtraKeyMaxAttempts, attempts)
}

// GetMaxAttempts 获取最大尝试次数，默认为5
func (v *Verify) GetMaxAttempts() int {
	attempts, ok := v.Extra.GetInt(verifyExtraKeyMaxAttempts)
	if !ok || attempts <= 0 {
		return 5 // 默认最大尝试次数
	}
	return attempts
}

// SetTrySends 设置尝试发送次数
func (v *Verify) SetTrySends(sends *int) {
	v.Extra.SetInt(verifyExtraKeyTrySends, sends)
}

// GetTrySends 获取尝试发送次数
func (v *Verify) GetTrySends() int {
	sends, ok := v.Extra.GetInt(verifyExtraKeyTrySends)
	if !ok || sends <= 0 {
		return 1 // 默认尝试发送次数
	}
	return sends
}
