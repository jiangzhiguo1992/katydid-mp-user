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
// getMaxNumByOwner func(ownKind model.TokenOwn, ownID uint64) (int, *errs.CodeErrs),
// isOwnerAuthEnable func(ownKind model.TokenOwn, ownID uint64, kind model.AuthKind) (bool, *errs.CodeErrs),
) *Verify {
	return &Verify{
		Base:    handler.NewBase(nil),
		service: service.NewVerify(
		//db, cache, isOwnerAuthEnable, getMaxNumByOwner,
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
	// TODO:GG 哪里传进来比较好呢
	//verify.SetExpireSec(param.ExpireS)
	//verify.SetMaxSends(param.MaxSends)
	//verify.SetMaxAttempts(param.MaxAttempts)

	// 添加记录
	verify, err = v.service.Add(verify)
	if err != nil {
		v.Response400("添加验证码失败", err)
		return
	}

	// TODO:GG 发送验证码/Oauth2/等等...
	sendOkAt := time.Now().Unix()

	// 发送成功
	verify, err = v.service.OnSendOk(verify.ID, &sendOkAt)
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
