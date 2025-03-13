package service

import (
	"fmt"
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/api/auth/repo/db"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/pkg/num"
	"strconv"
	"time"
)

type (
	// Verify 验证码服务
	Verify struct {
		*service.Base
		db *db.Verify
		//cache *cache.Verify
	}
)

func NewVerify(
	db *db.Verify, // cache *cache.Account,
) *Verify {
	return &Verify{
		Base: service.NewBase(nil),
		db:   db, // cache: cache,
	}
}

func (v *Verify) Add(verify *model.Verify) *errs.CodeErrs {
	body := ""
	switch verify.AuthKind {
	case model.AuthKindPhone, model.AuthKindEmail:
		body = num.Random(6) // TODO:GG 外部传config?
	default:
		return errs.Match2(fmt.Sprintf("不支持的验证类型 kind: %s", strconv.Itoa(int(verify.AuthKind))))
	}
	verify.SetBody(&body)

	// TODO:GG pendingAt之内的数量(发送成功的)
	temp := time.Now().Unix() - verify.GetPerSaves()
	verify.PendingAt = &temp
	results, err := v.db.Selects(verify)
	if err != nil {
		return err
	} else if (results != nil) && (len(results) > verify.GetMaxSaves()) {
		return errs.Match2("验证码数量超过限制")
	}

	// TODO:GG 检查authId和ownId是否存在

	verify.PendingAt = nil // reset
	verify.ValidAt = nil   // reset
	verify.ValidTimes = 0  // reset
	return v.db.Insert(verify)
}

func (v *Verify) Del(ID uint64) *errs.CodeErrs {
	// TODO:GG DB删除
	return nil
}

func (v *Verify) OnSendOk(verify *model.Verify) *errs.CodeErrs {
	verify.Status = model.VerifyStatusPending
	nowUnix := time.Now().Unix()
	if (verify.PendingAt == nil) || (*verify.PendingAt > nowUnix) {
		verify.PendingAt = &nowUnix
	}
	verify.ValidAt = nil  // reset
	verify.ValidTimes = 0 // reset
	return v.db.Update(verify)
}

func (v *Verify) OnSendFail(verify *model.Verify) *errs.CodeErrs {
	return nil
}

func (v *Verify) Valid(verify *model.Verify) (bool, *errs.CodeErrs) {
	body, ok := verify.GetBody()
	if !ok {
		return false, errs.Match2("验证：没有验证码！")
	}
	store, err := v.db.Select(verify)
	if err != nil {
		return false, err
	}
	if !store.CanValid() {
		return false, errs.Match2("失效的验证码")
	}
	realBody, ok := store.GetBody()
	if !ok {
		return false, errs.Match2("验证：没有验证码！")
	}
	validOk := false
	switch store.AuthKind {
	case model.AuthKindPhone, model.AuthKindEmail:
		validOk = realBody == body
	case model.AuthKindBiometric:
		// TODO:GG 生物特征
	case model.AuthKindThirdParty:
		// TODO:GG 三方
	default:
		return false, errs.Match2("验证：不支持的类型！")
	}
	if validOk {
		store.Status = model.VerifyStatusSuccess
		log.Debug("认证成功",
			log.FString("needCode", realBody),
			log.FString("requestCode", body),
		)
	} else {
		store.Status = model.VerifyStatusReject
		log.Debug("认证失败",
			log.FString("needCode", realBody),
			log.FString("requestCode", body),
		)
	}
	now := time.Now().Unix()
	store.ValidAt = &now
	store.ValidTimes++
	err = v.db.Update(store)
	return validOk, err
}
