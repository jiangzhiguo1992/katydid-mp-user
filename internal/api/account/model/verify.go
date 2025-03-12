package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/data"
	"katydid-mp-user/pkg/valid"
	"reflect"
	"time"
)

type (
	// Verify 验证内容
	Verify struct {
		*model.Base
		OwnKind  TokenOwn    `json:"ownType" validate:"required,own-check"`     // 验证平台 (组织/应用)
		AuthKind AuthKind    `json:"authKind" validate:"required,auth-check"`   // 认证类型 (手机号/邮箱/...)
		Apply    VerifyApply `json:"applyKind" validate:"required,apply-check"` // 申请类型 (注册/登录/修改密码/...)
		Target   string      `json:"target" validate:"required"`                // 标识，用户名/手机号/邮箱/生物特征/第三方平台

		OwnID  *uint64 `json:"ownId"`  // 认证拥有者Id (组织/应用)
		AuthId *uint64 `json:"authId"` // 认证Id

		PendingAt  *int64 `json:"pendingAt"`  // 等待时间(发送成功时间)
		VerifiedAt *int64 `json:"verifiedAt"` // 验证时间
		Attempts   int    `json:"attempts"`   // 验证次数
	}

	// VerifyApply 申请类型
	VerifyApply int16
)

func NewVerifyEmpty() *Verify {
	return &Verify{Base: model.NewBaseEmpty()}
}

func NewVerify(
	ownKind TokenOwn, authKind AuthKind, apply VerifyApply, target string,
	ownID *uint64, authID *uint64,
) *Verify {
	return &Verify{
		Base:    model.NewBase(make(data.KSMap)),
		OwnKind: ownKind, AuthKind: authKind, Apply: apply, Target: target,
		AuthId: authID, OwnID: ownID,
		PendingAt: nil, VerifiedAt: nil, Attempts: 0,
	}
}

func (v *Verify) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			"own-check": func(value reflect.Value, param string) bool {
				val := value.Interface().(TokenOwn)
				switch val {
				case TokenOwnOrg,
					TokenOwnApp,
					TokenOwnClient:
					return true
				default:
					return false
				}
			},
			"auth-check": func(value reflect.Value, param string) bool {
				val := value.Interface().(AuthKind)
				switch val {
				case AuthKindPhone,
					AuthKindEmail,
					AuthKindBiometric,
					AuthKindThirdParty:
					return true
				default:
					return false
				}
			},
			"apply-check": func(value reflect.Value, param string) bool {
				val := value.Interface().(VerifyApply)
				switch val {
				case VerifyApplyUnregister,
					VerifyApplyRegister,
					VerifyApplyLogin,
					VerifyApplyResetPwd,
					VerifyApplyChangePhone,
					VerifyApplyChangeEmail,
					VerifyApplyChangeBio,
					VerifyApplyChangeThird:
					return true
				default:
					return false
				}
			},
		},
	}
}

func (v *Verify) ValidExtraRules() (data.KSMap, valid.ExtraValidRules) {
	return v.Extra, valid.ExtraValidRules{
		valid.SceneSave: map[valid.Tag]valid.ExtraValidRuleInfo{
			// 验证内容
			verifyExtraKeyBody: {
				Field: verifyExtraKeyBody,
				ValidFn: func(value any) bool {
					val, ok := value.(string)
					if !ok {
						return false
					}
					return len(val) <= 0
				},
			},
		},
	}
}

func (v *Verify) ValidStructRules(scene valid.Scene, fn valid.FuncReportError) {
	switch scene {
	case valid.SceneAll:
		targetOk := false
		switch v.AuthKind {
		case AuthKindPhone:
			// 检测v.Number是否符合手机号格式
			_, _, targetOk = valid.IsPhone(v.Target)
		case AuthKindEmail:
			_, targetOk = valid.IsEmail(v.Target)
		default:
			targetOk = false
		}
		if !targetOk {
			fn(v.Target, "Target", "", "")
		}
	}
}

func (v *Verify) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"ownType":   {"required_own_type_err", false, nil},
					"authKind":  {"required_auth_kind_err", false, nil},
					"applyKind": {"required_apply_kind_err", false, nil},
					"target":    {"required_target_err", false, nil},
				},
				valid.TagFormat: {
					"CreateAt": {"format_create_at_err", false, nil},
					"DeleteAt": {"format_delete_at_err", false, nil},
					"DeleteBy": {"format_delete_by_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"own-check":        {"own_check_err", false, nil},
				"auth-check":       {"auth_check_err", false, nil},
				"apply-check":      {"apply_check_err", false, nil},
				verifyExtraKeyBody: {"format_body_err", false, nil},
			},
		},
	}
}

const (
	VerifyStatusPending model.Status = 1 // 等待验证
	VerifyStatusReject  model.Status = 2 // 验证失败
	VerifyStatusSuccess model.Status = 3 // 验证成功

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
		return 5 // 默认最大尝试次数
	}
	return attempts
}
