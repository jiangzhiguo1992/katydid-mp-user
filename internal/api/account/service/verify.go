package service

import (
	"fmt"
	"katydid-mp-user/internal/api/account/model"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/utils"
	"strconv"
	"time"
)

const (
	VerifyCodeLength = 6 // 验证码长度 // TODO:GG 移动到configs/clients
)

type (
	// Verify 验证码服务
	Verify struct {
		*service.Base

		Params struct {
			OnSendOk *struct {
				VerifyID  uint64
				PendingAt *int64
				ExpireSec int64
			}
			OnSendFail *struct {
				VerifyID uint64
				TryTimes int
			}
			Valid *struct {
				OwnKind  model.VerifyOwn
				OwnID    *uint64
				Number   string
				AuthKind model.AuthKind
				AuthID   *uint64
				Apply    model.VerifyApply
				Body     string
			}
		}
		//db    *db.Account
		//cache *cache.Account

		//IsOwnerAuthEnable func(ownKind model.TokenOwn, ownID uint64, kind model.AuthKind) (bool, *errs.CodeErrs)
		//GetMaxNumByOwner  func(ownKind model.TokenOwn, ownID uint64) (int, *errs.CodeErrs)
	}
)

func NewVerify(
// db *db.Account, cache *cache.Account,
// isOwnerAuthEnable func(ownKind model.TokenOwn, ownID uint64, kind model.AuthKind) (bool, *errs.CodeErrs),
// getMaxNumByOwner func(ownKind model.TokenOwn, ownID uint64) (int, *errs.CodeErrs),
) *Verify {
	return &Verify{
		Base: service.NewBase(nil),
		//db:   db, cache: cache,
		//IsOwnerAuthEnable: isOwnerAuthEnable,
		//GetMaxNumByOwner:  getMaxNumByOwner,
	}
}

func (v *Verify) Add(verify *model.Verify) (*model.Verify, *errs.CodeErrs) {
	body := ""
	switch verify.AuthKind {
	case model.AuthKindPhone:
	case model.AuthKindEmail:
		body = utils.Random(VerifyCodeLength)
	default:
		body = fmt.Sprintf("kind: %s", strconv.Itoa(int(verify.AuthKind)))
	}
	verify.SetBody(&body)

	// TODO:GG DB
	log.Debug("DB_添加验证", log.FAny("verify", verify))
	return verify, nil
}

func (v *Verify) Del(ID uint64) *errs.CodeErrs {
	// TODO:GG DB删除
	return nil
}

func (v *Verify) OnSendOk(ID uint64, pendingAt *int64) (*model.Verify, *errs.CodeErrs) {
	// TODO:GG DB获取
	verify := model.NewVerifyEmpty()
	log.Debug("DB_获取验证", log.FAny("verify", verify))

	verify.Status = model.VerifyStatusPending
	nowUnix := time.Now().Unix()
	if (pendingAt != nil) && (*pendingAt > 0) && (*pendingAt <= nowUnix) {
		verify.PendingAt = pendingAt
	} else {
		verify.PendingAt = &nowUnix
	}
	verify.VerifiedAt = nil // reset
	verify.Attempts = 0     // reset

	// TODO:GG DB update
	log.Debug("DB_修改验证", log.FAny("verify", verify))

	return verify, nil
}

func (v *Verify) OnSendFail() (*model.Verify, *errs.CodeErrs) {
	return nil, nil
}

func (v *Verify) Valid() (bool, *errs.CodeErrs) {
	param := v.Params.Valid

	// TODO:GG DB获取
	verify := &model.Verify{}
	log.Debug("DB_获取验证", log.FAny("verify", verify))

	if !verify.CanValid() {
		return false, errs.Match2("失效的验证码")
	}

	switch param.AuthKind {
	case model.AuthKindPhone:
	case model.AuthKindEmail:

		body, ok := verify.GetBody()
		if !ok {
			return false, errs.Match2("验证：没有验证码！")
		}
		verify.Attempts++
		if body != param.Body {
			log.Debug("认证失败",
				log.FString("needCode", body),
				log.FString("requestCode", param.Body),
			)
			verify.Status = model.VerifyStatusReject
			log.Debug("DB_更新", log.FAny("verify", verify))
		} else if verify.Status == model.VerifyStatusSuccess {
			log.Warn("重复认证失败",
				log.FString("needCode", body),
				log.FString("requestCode", param.Body),
			)
		} else {
			verify.Status = model.VerifyStatusSuccess
			now := time.Now().Unix()
			verify.VerifiedAt = &now
			log.Debug("DB_更新", log.FAny("verify", verify))
		}
	case model.AuthKindBiometric:
		// TODO:GG 生物特征
	case model.AuthKindThirdParty:
		// TODO:GG 三方
	default:
		return false, errs.Match2("验证：不支持的类型！")
	}
	return false, nil
}
