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
		OwnKind  OwnKind     `json:"ownKind" validate:"required,own-range"`   // 验证平台 (组织/应用)
		OwnID    uint64      `json:"ownId"`                                   // 认证拥有者Id (组织/应用)
		AuthKind AuthKind    `json:"authKind" validate:"required,auth-range"` // 认证类型 (手机号/邮箱/...)
		Apply    VerifyApply `json:"apply" validate:"required,apply-range"`   // 申请类型 (注册/登录/修改密码/...)
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

func NewVerify(
	ownKind OwnKind, ownID uint64, authKind AuthKind, apply VerifyApply, target []string,
) *Verify {
	base := model.NewBase(make(data.KSMap))
	base.Status = VerifyStatusInit
	return &Verify{
		Base:    base,
		OwnKind: ownKind, OwnID: ownID, AuthKind: authKind, Apply: apply, Target: target,
		SendAt: nil, ValidAt: nil, ValidTimes: 0,
	}
}

func (v *Verify) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 所属类型
			"own-range": func(value reflect.Value, param string) bool {
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
			"auth-range": func(value reflect.Value, param string) bool {
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
			"apply-range": func(value reflect.Value, param string) bool {
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
			fn(v.Target, "Target", "target-format", "")
		}
	}
}

func (v *Verify) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"OwnKind":  {"required_own_kind_err", false, nil},
					"AuthKind": {"required_auth_kind_err", false, nil},
					"Apply":    {"required_apply_kind_err", false, nil},
					"Target":   {"required_target_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"own-range":        {"own_range_err", false, nil},
				"auth-range":       {"auth_range_err", false, nil},
				"apply-range":      {"apply_range_err", false, nil},
				"target-format":    {"target_format_err", false, nil},
				verifyExtraKeyBody: {"body_format_err", false, nil},
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
