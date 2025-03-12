package handler

import (
	"katydid-mp-user/internal/api/account/model"
	"katydid-mp-user/internal/api/account/service"
	"katydid-mp-user/internal/pkg/handler"
	"time"
)

type Verify struct {
	*handler.Base
	service *service.Verify
}

func NewVerify(
// db *db.Account, cache *cache.Account,
) *Verify {
	return &Verify{
		Base: handler.NewBase(nil),
		service: service.NewVerify(
			nil,
		),
	}
}

func (v *Verify) Post() {
	verify := model.NewVerifyEmpty()
	err := v.RequestBind(verify, true)
	if err != nil {
		v.Response400("绑定失败", err)
		return
	}

	// TODO:GG 哪里传进来比较好呢 (client.Extra)
	//verify.SetExpireSec(param.ExpireS)
	//verify.SetMaxSends(param.MaxSends)

	// 添加记录
	err = v.service.Add(verify)
	if err != nil {
		v.Response400("添加验证码失败", err)
		return
	}

	//verifyExtraKeyPerSends  = "perSends"  // 发送时间范围 TODO:GG 上层实现
	//verifyExtraKeyMaxSends  = "maxSends"  // 最大发送次数 TODO:GG 上层实现
	sendOk := true // TODO:GG rpc发送验证码/Oauth2/等等...
	sendAt := time.Now().Unix()
	if sendOk {
		verify.PendingAt = &sendAt // TODO:GG 上层返回
		err = v.service.OnSendOk(verify)
	} else {
		err = v.service.OnSendFail(verify)
	}
	if err != nil {
		v.Response400("发送验证码失败", err)
		return
	}
	v.Response200(verify)
}

func (v *Verify) Del() {

}

func (v *Verify) Put() {

}

func (v *Verify) Get() {

}
