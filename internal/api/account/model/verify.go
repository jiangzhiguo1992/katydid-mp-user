package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/utils"
	"time"
)

type (
	// Verify 验证内容
	Verify struct {
		*model.Base
		OwnKind VerifyOwn `json:"ownType"` // 验证平台 (组织/应用)
		OwnID   *uint64   `json:"ownId"`   // 认证拥有者Id (组织/应用)
		Number  string    `json:"number"`  // 标识，用户名/手机号/邮箱/生物特征/第三方平台

		AuthKind AuthKind    `json:"kind"`      // 认证类型 (手机号/邮箱/...)
		AuthId   *uint64     `json:"authId"`    // 认证Id
		Apply    VerifyApply `json:"applyKind"` // 申请类型 (注册/登录/修改密码/...)

		PendingAt  *int64 `json:"pendingAt"`  // 等待时间(发送成功时间)
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
	ownKind VerifyOwn, ownID *uint64, number string,
	authKind AuthKind, authID *uint64, apply VerifyApply,
) *Verify {
	return &Verify{
		Base:    model.NewBase(make(utils.KSMap)),
		OwnKind: ownKind, OwnID: ownID, Number: number,
		AuthKind: authKind, AuthId: authID, Apply: apply,
		PendingAt: nil, VerifiedAt: nil, Attempts: 0,
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
	if v.PendingAt == nil {
		return false
	}
	expireSec := v.GetExpireSec()
	now := time.Now().Unix()
	return (now - *v.PendingAt) >= expireSec
}

// IsVerified 检查是否已验证成功
func (v *Verify) IsVerified() bool {
	return v.Status == VerifyStatusSuccess && v.VerifiedAt != nil
}

func (v *Verify) CanValid() bool {
	if v.Status < VerifyStatusPending || v.Status >= VerifyStatusSuccess {
		return false
	} else if v.IsExpired() {
		return false
	} else if v.Attempts >= v.GetMaxAttempts() {
		return false
	}
	return true
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
	verifyExtraKeyExpireSec   = "expireSec"   // 过期时间S
	verifyExtraKeyMaxSends    = "maxSends"    // 最大发送次数
	verifyExtraKeyMaxAttempts = "maxAttempts" // 最大尝试次数
)

func (v *Verify) SetBody(body *string) {
	v.Extra.SetString(verifyExtraKeyBody, body)
}

func (v *Verify) GetBody() (string, bool) {
	return v.Extra.GetString(verifyExtraKeyBody)
}

func (v *Verify) GetExpireSec() int64 {
	expireSec, ok := v.Extra.GetInt64(verifyExtraKeyExpireSec)
	if !ok || expireSec <= 0 {
		return 60 * 5 // 默认过期时间为5分
	}
	return expireSec
}

func (v *Verify) SetExpireSec(expireSec *int64) {
	v.Extra.SetInt64(verifyExtraKeyExpireSec, expireSec)
}

// SetMaxSends 设置最大发送次数
func (v *Verify) SetMaxSends(sends *int) {
	v.Extra.SetInt(verifyExtraKeyMaxSends, sends)
}

// GetMaxSends 获取最大发送次数
func (v *Verify) GetMaxSends() int {
	sends, ok := v.Extra.GetInt(verifyExtraKeyMaxSends)
	if !ok || sends <= 0 {
		return 3 // 默认最大发送次数
	}
	return sends
}

// SetMaxAttempts 设置最大尝试次数
func (v *Verify) SetMaxAttempts(attempts *int) {
	v.Extra.SetInt(verifyExtraKeyMaxAttempts, attempts)
}

// GetMaxAttempts 获取最大尝试次数，默认为5
func (v *Verify) GetMaxAttempts() int {
	attempts, ok := v.Extra.GetInt(verifyExtraKeyMaxAttempts)
	if !ok || attempts <= 0 {
		return 10 // 默认最大尝试次数
	}
	return attempts
}
