package service

import (
	"fmt"
	"katydid-mp-user/internal/api/account/model"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/pkg/valid"
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
			Add *struct {
				OwnKind     model.VerifyOwn
				OwnID       *uint64
				Number      string
				AuthKind    model.AuthKind
				AuthID      *uint64
				Apply       model.VerifyApply
				ExpireS     *int64
				MaxSends    *int
				MaxAttempts *int
			}
			Del *struct {
			}
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
			}
		}
		//db    *db.Account
		//cache *cache.Account

		//IsOwnerAuthEnable func(ownKind model.TokenOwn, ownID uint64, kind model.AuthKind) (bool, *errs.CodeErrs)
		//GetMaxNumByOwner  func(ownKind model.TokenOwn, ownID uint64) (int, *errs.CodeErrs)
	}
)

func (v *Verify) Add() (*model.Verify, *errs.CodeErrs) {
	param := v.Params.Add
	verify := model.NewVerify(
		param.OwnKind, param.OwnID, param.Number,
		param.AuthKind, param.AuthID, param.Apply,
	)
	valid.Check(verify, valid.SceneInsert) // TODO:GG 完善
	body := ""
	switch verify.AuthKind {
	case model.AuthKindPhone:
	case model.AuthKindEmail:
		body = utils.Random(VerifyCodeLength)
	default:
		body = fmt.Sprintf("kind: %s", strconv.Itoa(int(verify.AuthKind)))
	}
	verify.SetBody(&body)
	verify.SetExpireSec(param.ExpireS)
	verify.SetMaxSends(param.MaxSends)
	verify.SetMaxAttempts(param.MaxAttempts)

	// TODO:GG DB
	log.Debug("DB_添加验证", log.FAny("verify", verify))
	return verify, nil
}

func (v *Verify) Del() *errs.CodeErrs {
	// TODO:GG DB删除
	return nil
}

func (v *Verify) AfterSend() (*model.Verify, *errs.CodeErrs) {
	param := v.Params.OnSendOk
	// TODO:GG DB获取
	verify := &model.Verify{}
	log.Debug("DB_获取验证", log.FAny("verify", verify))

	verify.Status = model.VerifyStatusPending
	if param.PendingAt != nil && *param.PendingAt > 0 {
		verify.PendingAt = param.PendingAt
	} else {
		now := time.Now().Unix()
		verify.PendingAt = &now
	}
	verify.VerifiedAt = nil // reset
	verify.Attempts = 0     // reset

	// TODO:GG DB update
	log.Debug("DB_修改验证", log.FAny("verify", verify))

	return verify, nil
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

	case model.AuthKindBiometric:
		// TODO:GG 生物特征
	case model.AuthKindThirdParty:
		// TODO:GG 三方
	default:
		return false, errs.Match2("验证：不支持的类型！")
	}

	data := auth.Extra
	if v.Kind == model.AuthKindPhone || v.Kind == model.AuthKindEmail {
		if v.State == model.VerityStateInit {
			return false, error2.MatchByMessage("验证：还没开始呢！")
		} else if v.State == model.VerityStateReject {
			if v.Attempts >= VerifyMaxAttempts {
				return false, error2.MatchByMessage("验证：超过最大次数了！")
			}
		}
		if time.Now().UnixMilli() > v.ExpiresAt {
			log.Debug("认证超时",
				log.Int64("pending", v.PendingAt),
				log.Int64("now", time.Now().UnixMilli()),
				log.Int64("超出s", (time.Now().UnixMilli()-v.PendingAt)/1000),
			)
			return false, error2.MatchByMessage("验证：已经过期了！")
		}
		needCode := v.GetCode()
		if dataCode, ok := data["code"]; !ok {
			return false, error2.MatchByMessage("验证：没有验证码！")
		} else if (dataCode == nil) || len(dataCode.(string)) <= 0 {
			return false, error2.MatchByMessage("验证：验证码为空！")
		} else if needCode != dataCode.(string) {
			log.Debug("认证失败",
				log.String("needCode", needCode),
				log.String("requestCode", dataCode.(string)),
			)
			v.State = model.VerityStateReject
			v.VerifiedAt = time.Now().UnixMilli()
			v.Attempts++
			// TODO:GG DB
			return false, error2.MatchByMessage("验证：验证码错误！")
		}
		log.Debug("认证成功",
			log.String("needCode", needCode),
			log.String("requestCode", data["code"].(string)),
		)
		if v.State != model.VerityStateSuccess {
			v.State = model.VerityStateSuccess
			v.VerifiedAt = time.Now().UnixMilli()
			// TODO:GG DB
		} else {
			log.Warn("重复认证成功",
				log.String("needCode", needCode),
				log.String("requestCode", data["code"].(string)),
			)
		}
		return true, nil
	}
	return false, errs.Match(nil) // "验证：不支持的类型！"
}
