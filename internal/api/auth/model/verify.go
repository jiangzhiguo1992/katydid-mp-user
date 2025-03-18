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
		OwnKind  OwnKind     `json:"ownKind" validate:"required,range-own"`   // 验证平台 (组织/应用)
		OwnID    uint64      `json:"ownId"`                                   // 认证拥有者Id (组织/应用)
		AuthKind AuthKind    `json:"authKind" validate:"required,range-auth"` // 认证类型 (手机号/邮箱/...)
		Apply    VerifyApply `json:"apply" validate:"required,range-apply"`   // 申请类型 (注册/登录/修改密码/...)
		Target   []string    `json:"target" validate:"required"`              // 标识，用户名/手机号/邮箱/生物特征/第三方平台 // TODO:GG DB存{"cellphone:86_12345678901"}/{"email:address@domain"}

		SendAt     *int64 `json:"sendAt"`     // 发送时间(发送成功时间)
		ValidAt    *int64 `json:"validAt"`    // 验证时间
		ValidTimes int    `json:"validTimes"` // 验证次数
	}

	// VerifyApply 申请类型
	VerifyApply int16
)

func NewVerifyEmpty() *Verify {
	return &Verify{Base: model.NewBaseEmpty()}
}

func (v *Verify) Wash() *Verify {
	v.Base = model.NewBase(make(data.KSMap))
	v.Status = VerifyStatusInit
	v.SendAt = nil
	v.ValidAt = nil
	v.ValidTimes = 0
	return v
}

func (v *Verify) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 所属类型
			"range-own": func(value reflect.Value, param string) bool {
				val := value.Interface().(OwnKind)
				switch val {
				case OwnKindOrg,
					OwnKindRole,
					OwnKindApp,
					OwnKindClient,
					OwnKindUser:
					return true
				default:
					return false
				}
			},
			// 认证类型
			"range-auth": func(value reflect.Value, param string) bool {
				val := value.Interface().(AuthKind)
				switch val {
				case AuthKindCellphone,
					AuthKindEmail,
					AuthKindBioFace,
					AuthKindBioFinger,
					AuthKindBioVoice,
					AuthKindBioIris,
					AuthKindThirdGoogle,
					AuthKindThirdApple,
					AuthKindThirdWechat,
					AuthKindThirdQQ,
					AuthKindThirdIns,
					AuthKindThirdFB:
					return true
				default:
					return false
				}
			},
			// 验证类型
			"range-apply": func(value reflect.Value, param string) bool {
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
					return len(val) > 0
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
		case AuthKindCellphone:
			if len(v.Target) == 2 {
				_, _, targetOk = valid.IsPhoneNumber(v.Target[0], v.Target[1])
			}
		case AuthKindEmail:
			if len(v.Target) == 2 {
				if valid.IsEmailUsername(v.Target[0]) {
					targetOk = valid.IsEmailDomain(v.Target[1])
				}
			}
		default:
			targetOk = false
		}
		if !targetOk {
			// 接受验证的目标
			fn(v.Target, "Target", valid.TagFormat, "")
		}
	}
}

func (v *Verify) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"OwnKind":  {"required_verify_own_kind_err", false, nil},
					"AuthKind": {"required_verify_auth_kind_err", false, nil},
					"Apply":    {"required_verify_apply_kind_err", false, nil},
					"Target":   {"required_verify_target_err", false, nil},
				},
				valid.TagFormat: {
					"Target": {"format_verify_target_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"range-own":        {"range_verify_own_err", false, nil},
				"range-auth":       {"range_verify_auth_err", false, nil},
				"range-apply":      {"range_verify_apply_err", false, nil},
				verifyExtraKeyBody: {"format_verify_body_err", false, nil},
			},
		},
	}
}

const (
	VerifyStatusInit    model.Status = 0 // 初始状态
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
func (v *Verify) IsExpired(expireSec int64) bool {
	if v.SendAt == nil {
		return false
	}
	now := time.Now().Unix()
	return (now - *v.SendAt) >= expireSec
}

// IsVerified 检查是否已验证成功
func (v *Verify) IsVerified() bool {
	return v.Status == VerifyStatusSuccess && v.ValidAt != nil
}

func (v *Verify) SetPending() {
	v.Status = VerifyStatusPending
	nowUnix := time.Now().Unix()
	if (v.SendAt == nil) || (*v.SendAt > nowUnix) {
		v.SendAt = &nowUnix
	}
	v.ValidAt = nil  // reset
	v.ValidTimes = 0 // reset
}

func (v *Verify) SetSuccess() {
	v.Status = VerifyStatusSuccess
	nowUnix := time.Now().Unix()
	v.ValidAt = &nowUnix
	v.ValidTimes++
}

func (v *Verify) SetReject() {
	v.Status = VerifyStatusReject
	nowUnix := time.Now().Unix()
	v.ValidAt = &nowUnix
	v.ValidTimes++
}

// CanValid 检查是否可以验证
func (v *Verify) CanValid(expireSec int64, maxValidTimes int) bool {
	if v.Status < VerifyStatusPending || v.Status >= VerifyStatusSuccess {
		return false
	} else if v.IsExpired(expireSec) {
		return false
	} else if v.ValidTimes >= maxValidTimes {
		return false
	}
	return true
}

// Valid 验证
func (v *Verify) Valid(body string) bool {
	exist, ok := v.GetBody()
	if !ok {
		return false
	}
	validOk := false
	switch v.AuthKind {
	case AuthKindCellphone:
		validOk = exist == body
	case AuthKindEmail: // TODO:GG 要是链接呢?
		validOk = exist == body
	default:
		return false
	}
	return validOk
}

const (
	verifyExtraKeyBody = "body" // 验证内容
)

func (v *Verify) SetBody(body *string) {
	v.Extra.SetString(verifyExtraKeyBody, body)
}

func (v *Verify) GetBody() (string, bool) {
	return v.Extra.GetString(verifyExtraKeyBody)
}
