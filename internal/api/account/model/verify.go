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
		OwnKind  TokenOwn    `json:"ownKind" validate:"required,own-check"`   // 验证平台 (组织/应用)
		AuthKind AuthKind    `json:"authKind" validate:"required,auth-check"` // 认证类型 (手机号/邮箱/...)
		Apply    VerifyApply `json:"apply" validate:"required,apply-check"`   // 申请类型 (注册/登录/修改密码/...)
		Target   []string    `json:"target" validate:"required"`              // 标识，用户名/手机号/邮箱/生物特征/第三方平台

		OwnID  *uint64 `json:"ownId"`  // 认证拥有者Id (组织/应用)
		AuthId *uint64 `json:"authId"` // 认证Id

		PendingAt  *int64 `json:"pendingAt"`  // 等待时间(发送成功时间)
		VerifiedAt *int64 `json:"verifiedAt"` // 验证时间
		ValidTimes int    `json:"validTimes"` // 验证次数
	}

	// VerifyApply 申请类型
	VerifyApply int16
)

func NewVerifyEmpty() *Verify {
	return &Verify{Base: model.NewBaseEmpty()}
}

func NewVerify(
	ownKind TokenOwn, authKind AuthKind, apply VerifyApply, target []string,
	ownID *uint64, authID *uint64,
) *Verify {
	return &Verify{
		Base:    model.NewBase(make(data.KSMap)),
		OwnKind: ownKind, AuthKind: authKind, Apply: apply, Target: target,
		AuthId: authID, OwnID: ownID,
		PendingAt: nil, VerifiedAt: nil, ValidTimes: 0,
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
			if len(v.Target) == 1 {
				_, _, targetOk = valid.IsPhone(v.Target[0])
			} else if len(v.Target) == 1 {
				_, _, targetOk = valid.IsPhoneNumber(v.Target[0], v.Target[1])
			}
		case AuthKindEmail:
			if len(v.Target) == 1 {
				_, targetOk = valid.IsEmail(v.Target[0])
			}
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
					"ownKind":  {"required_own_type_err", false, nil},
					"authKind": {"required_auth_kind_err", false, nil},
					"apply":    {"required_apply_kind_err", false, nil},
					"target":   {"required_target_err", false, nil},
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
	now := time.Now().Unix()
	expireSec := v.GetExpireSec()
	return (now - *v.PendingAt) >= expireSec
}

// IsVerified 检查是否已验证成功
func (v *Verify) IsVerified() bool {
	return v.Status == VerifyStatusSuccess && v.VerifiedAt != nil
}

// CanValid 检查是否可以验证 TODO:GG 上层实现times++
func (v *Verify) CanValid() bool {
	if v.Status < VerifyStatusPending || v.Status >= VerifyStatusSuccess {
		return false
	} else if v.IsExpired() {
		return false
	} else if v.ValidTimes >= v.GetMaxTimes() {
		return false
	}
	return true
}

const (
	verifyExtraKeyBody      = "body"      // 验证内容
	verifyExtraKeyExpireSec = "expireSec" // 过期时间S
	verifyExtraKeyPerSends  = "perSends"  // 发送时间范围 TODO:GG 上层实现
	verifyExtraKeyMaxSends  = "maxSends"  // 最大发送次数 TODO:GG 上层实现
	verifyExtraKeyPerSaves  = "perSaves"  // 保存时间范围 TODO:GG 上层实现
	verifyExtraKeyMaxSaves  = "maxSaves"  // 最大保存次数 TODO:GG 上层实现
	verifyExtraKeyMaxTimes  = "maxTimes"  // 最大验证次数
)

func (v *Verify) SetBody(body *string) {
	v.Extra.SetString(verifyExtraKeyBody, body)
}

func (v *Verify) GetBody() (string, bool) {
	return v.Extra.GetString(verifyExtraKeyBody)
}

func (v *Verify) GetExpireSec() int64 {
	val, ok := v.Extra.GetInt64(verifyExtraKeyExpireSec)
	if !ok || val <= 0 {
		return 60 * 5 // 默认过期时间为5分
	}
	return val
}

func (v *Verify) SetExpireSec(expireSec *int64) {
	v.Extra.SetInt64(verifyExtraKeyExpireSec, expireSec)
}

func (v *Verify) SetPerSends(perSends *int64) {
	v.Extra.SetInt64(verifyExtraKeyPerSends, perSends)
}

func (v *Verify) GetPerSends() int64 {
	val, ok := v.Extra.GetInt64(verifyExtraKeyPerSends)
	if !ok || val <= 0 {
		return 60 // 默认发送时间范围为60秒
	}
	return val
}

func (v *Verify) SetMaxSends(sends *int) {
	v.Extra.SetInt(verifyExtraKeyMaxSends, sends)
}

func (v *Verify) GetMaxSends() int {
	val, ok := v.Extra.GetInt(verifyExtraKeyMaxSends)
	if !ok || val <= 0 {
		return 3 // 默认最大发送次数(60s)
	}
	return val
}

func (v *Verify) SetPerSaves(perSaves *int64) {
	v.Extra.SetInt64(verifyExtraKeyPerSaves, perSaves)
}

func (v *Verify) GetPerSaves() int64 {
	val, ok := v.Extra.GetInt64(verifyExtraKeyPerSaves)
	if !ok || val <= 0 {
		return 60 * 60 // 默认保存时间范围为1小时
	}
	return val
}

func (v *Verify) SetMaxSaves(saves *int) {
	v.Extra.SetInt(verifyExtraKeyMaxSaves, saves)
}

func (v *Verify) GetMaxSaves() int {
	val, ok := v.Extra.GetInt(verifyExtraKeyMaxSaves)
	if !ok || val <= 0 {
		return 5 // 默认最大保存次数(1h)
	}
	return val
}

func (v *Verify) SetMaxTimes(attempts *int) {
	v.Extra.SetInt(verifyExtraKeyMaxTimes, attempts)
}

func (v *Verify) GetMaxTimes() int {
	val, ok := v.Extra.GetInt(verifyExtraKeyMaxTimes)
	if !ok || val <= 0 {
		return 10 // 默认最大尝试次数(60s)
	}
	return val
}
