package service

import (
	"fmt"
	"katydid-mp-user/internal/api/account/model"
	"katydid-mp-user/internal/api/account/repo/db"
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

func (v *Verify) Add(entity *model.Verify) *errs.CodeErrs {
	body := ""
	switch entity.AuthKind {
	case model.AuthKindPhone:
	case model.AuthKindEmail:
		body = num.Random(6) // TODO:GG 外部传config?
	default:
		return errs.Match2(fmt.Sprintf("不支持的验证类型 kind: %s", strconv.Itoa(int(entity.AuthKind))))
	}
	entity.SetBody(&body)

	// TODO:GG pendingAt之内的数量(发送成功的)
	temp := time.Now().Unix() - entity.GetPerSaves()
	entity.PendingAt = &temp
	results, err := v.db.Selects(entity)
	if err != nil {
		return err
	} else if (results != nil) && (len(results) > entity.GetMaxSaves()) {
		return errs.Match2("验证码数量超过限制")
	}

	// TODO:GG 检查authId和ownId是否存在

	entity.PendingAt = nil // reset
	entity.ValidAt = nil   // reset
	entity.ValidTimes = 0  // reset
	return v.db.Insert(entity)
}

func (v *Verify) Del(ID uint64) *errs.CodeErrs {
	// TODO:GG DB删除
	return nil
}

func (v *Verify) OnSendOk(entity *model.Verify) *errs.CodeErrs {
	entity.Status = model.VerifyStatusPending
	nowUnix := time.Now().Unix()
	if (entity.PendingAt == nil) || (*entity.PendingAt > nowUnix) {
		entity.PendingAt = &nowUnix
	}
	entity.ValidAt = nil  // reset
	entity.ValidTimes = 0 // reset
	return v.db.Update(entity)
}

func (v *Verify) OnSendFail(verify *model.Verify) *errs.CodeErrs {
	return nil
}

func (v *Verify) Valid(entity *model.Verify) (bool, *errs.CodeErrs) {
	body, ok := entity.GetBody()
	if !ok {
		return false, errs.Match2("验证：没有验证码！")
	}
	verify, err := v.db.Select(entity)
	if err != nil {
		return false, err
	}
	if !verify.CanValid() {
		return false, errs.Match2("失效的验证码")
	}
	realBody, ok := verify.GetBody()
	if !ok {
		return false, errs.Match2("验证：没有验证码！")
	}
	validOk := false
	switch verify.AuthKind {
	case model.AuthKindPhone:
	case model.AuthKindEmail:
		validOk = realBody == body
	case model.AuthKindBiometric:
		// TODO:GG 生物特征
	case model.AuthKindThirdParty:
		// TODO:GG 三方
	default:
		return false, errs.Match2("验证：不支持的类型！")
	}
	if validOk {
		verify.Status = model.VerifyStatusSuccess
		log.Debug("认证成功",
			log.FString("needCode", realBody),
			log.FString("requestCode", body),
		)
	} else {
		verify.Status = model.VerifyStatusReject
		log.Debug("认证失败",
			log.FString("needCode", realBody),
			log.FString("requestCode", body),
		)
	}
	now := time.Now().Unix()
	verify.ValidAt = &now
	verify.ValidTimes++
	err = v.db.Update(verify)
	return validOk, err
}
