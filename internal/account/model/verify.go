package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/utils"
	"time"
)

const (
	VerifyKindPwd   int16 = 1 // 密码
	VerifyKindPhone int16 = 2 // 短信
	VerifyKindEmail int16 = 3 // 邮箱
	VerifyKindBio   int16 = 4 // 生物特征
	VerifyKindThird int16 = 5 // 第三方平台

	VerityStateInit    int8 = 1 // 初始状态
	VerityStatePending int8 = 2 // 等待验证
	VerityStateSuccess int8 = 3 // 验证成功
	VerityStateReject  int8 = 4 // 验证失败

)

// Verifier 验证器接口
type Verifier interface {
	Generate(count int) string
	GetExpiresAt(intervalMin int64) int64
	IsExpired() bool
	IsOverMax(max int8) bool
	Validate(input string) bool
}

// CodeVerifier 验证码验证器
type CodeVerifier struct {
	*VerifyInfo
}

// Generate 生成验证码
func (v *CodeVerifier) Generate(count int) string {
	return utils.Random(count)
}

func (v *CodeVerifier) GetExpiresAt(intervalMin int64) int64 {
	return time.Now().UnixMilli() + intervalMin*60*1000
}

func (v *CodeVerifier) IsExpired() bool {
	return time.Now().UnixMilli() > v.ExpiresAt
}

func (v *CodeVerifier) IsOverMax(max int8) bool {
	return v.Attempts >= max
}

func (v *CodeVerifier) Validate(input string) bool {
	return v.GetCode() == input
}

// VerifyInfo 验证内容
type VerifyInfo struct {
	*model.Base
	ClientId  uint64 `json:"clientId"`  // 客户端Id
	AccountId uint64 `json:"accountId"` // 账户Id
	Kind      int16  `json:"kind"`      // 平台 (手机号/邮箱/...)

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

func NewVerifyInfoDef(clientId, accountId uint64, kind int16, expireAt int64) *VerifyInfo {
	verify := &VerifyInfo{
		Base:       model.NewBaseEmpty(),
		ClientId:   clientId,
		AccountId:  accountId,
		Kind:       kind,
		State:      VerityStateInit,
		PendingAt:  -1,
		VerifiedAt: -1,
		ExpiresAt:  expireAt,
		Attempts:   0,
	}
	return verify
}

// WithBody 设置验证内容 (注意language)
func (v *VerifyInfo) WithBody(data *string) map[string]interface{} {
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
func (v *VerifyInfo) WithCode(data *string) map[string]interface{} {
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
