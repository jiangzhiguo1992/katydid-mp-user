package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/utils"
	"time"
)

const (
	VerityStateInit    int8 = 1 // 初始状态
	VerityStatePending int8 = 2 // 等待验证
	VerityStateSuccess int8 = 3 // 验证成功
	VerityStateReject  int8 = 4 // 验证失败
)

// VerifyInfo 验证内容
type VerifyInfo struct {
	*model.Base
	ClientId  uint64 `json:"clientId"`  // 客户端Id
	AccountId uint64 `json:"accountId"` // 账户Id
	Kind      int16  `json:"kind"`      // 平台 (手机号/邮箱/...)

	State       int8  `json:"state"`       // 验证状态
	PendingAt   int64 `json:"pendingAt"`   // 等待时间
	VerityAt    int64 `json:"verityAt"`    // 验证时间
	ExpireAt    int64 `json:"expireAt"`    // 过期时间
	VerifyTimes int8  `json:"verifyTimes"` // 验证次数

	// TODO:GG ip, device, location
}

func NewVerifyInfo(clientId, accountId uint64, kind int16, expireAt int64) *VerifyInfo {
	verify := &VerifyInfo{
		Base:        model.NewBaseEmpty(),
		ClientId:    clientId,
		AccountId:   accountId,
		Kind:        kind,
		State:       VerityStateInit,
		PendingAt:   -1,
		VerityAt:    -1,
		ExpireAt:    expireAt,
		VerifyTimes: 0,
	}
	verify.GenerateCode()
	return verify
}

// SetBody 设置验证内容 (注意language)
func (v *VerifyInfo) SetBody(data *string) {
	if (data != nil) && (len(*data) > 0) {
		v.Extra["body"] = *data
	} else {
		delete(v.Extra, "body")
	}
}

func (v *VerifyInfo) GetBody() string {
	if v.Extra["body"] == nil {
		return ""
	}
	return v.Extra["body"].(string)
}

// SetCode 设置验证码
func (v *VerifyInfo) SetCode(data *string) {
	if (data != nil) && (len(*data) > 0) {
		v.Extra["code"] = *data
	} else {
		delete(v.Extra, "code")
	}
}

func (v *VerifyInfo) GetCode() string {
	if v.Extra["code"] == nil {
		return ""
	}
	return v.Extra["code"].(string)
}

// GenerateCode 生成验证码
func (v *VerifyInfo) GenerateCode() {
	code := utils.Random(6)
	v.SetCode(&code)
	log.Debug("验证：生成验证码", log.String("code", code))
}

func (v *VerifyInfo) IsExpired() bool {
	return time.Now().UnixMilli() < v.ExpireAt
}

func (v *VerifyInfo) SetPending() {
	v.PendingAt = time.Now().UnixMilli()
	v.State = VerityStatePending
	log.Debug("验证：开始等待")
}

func (v *VerifyInfo) SetSuccess() {
	v.State = VerityStateSuccess
	v.VerityAt = time.Now().UnixMilli()
	log.Debug("验证：成功")
}

func (v *VerifyInfo) SetReject() {
	v.State = VerityStateReject
	v.VerityAt = time.Now().UnixMilli()
	v.VerifyTimes++
	log.Debug("验证：失败")
}
