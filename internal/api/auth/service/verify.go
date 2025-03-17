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
		db     *db.Verify
		dbAuth *db.Auth
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

func (svc *Verify) Add(param *model.Verify) *errs.CodeErrs {

	// TODO:GG 检查ownId是否存在

	// 检查auth是否存在
	existAuth, err := svc.dbAuth.Select(nil) // TODO:GG 根据 AuthKind + Target 查找Auth
	if err != nil {
		return err
	} else if existAuth == nil {
		return errs.Match2(fmt.Sprintf("没有找到对应的auth %d", param.AuthKind))
	}

	// 生成验证码
	err = svc.generateBody(param)
	if err != nil {
		return err
	}

	return svc.addVerify(param)
}

func (svc *Verify) Del(param *model.Verify) *errs.CodeErrs {
	// TODO:GG DB删除
	return nil
}

func (svc *Verify) OnSendOk(param *model.Verify) *errs.CodeErrs {
	// 不检查ownerID和authID了
	param.Status = model.VerifyStatusPending
	nowUnix := time.Now().Unix()
	if (param.SendAt == nil) || (*param.SendAt > nowUnix) {
		param.SendAt = &nowUnix
	}
	param.ValidAt = nil  // reset TODO:GG DB里做?
	param.ValidTimes = 0 // reset TODO:GG DB里做?
	return svc.db.Update(param)
}

func (svc *Verify) OnSendFail(verify *model.Verify) *errs.CodeErrs {
	// 不检查ownerID和authID了
	return nil
}

func (svc *Verify) Valid(param *model.Verify) *errs.CodeErrs {
	// 检查传进来的参数
	body, ok := param.GetBody()
	if !ok {
		return errs.Match2("验证：没有验证码！")
	} else if len(body) <= 0 {
		return errs.Match2("验证：验证码不能为空！")
	}

	// TODO:GG 检查ownId是否存在

	// 检查auth是否存在
	existAuth, err := svc.dbAuth.Select(nil) // TODO:GG 根据 AuthKind + Target 查找Auth
	if err != nil {
		return err
	} else if existAuth == nil {
		return errs.Match2(fmt.Sprintf("没有找到对应的auth %d", param.AuthKind))
	}

	// 检查已存在验证码的合法性
	exist, err := svc.checkExist(param)
	if err != nil {
		return err
	}
	existBody, ok := exist.GetBody()

	// 验证内容体
	validOk := exist.Valid(body)

	// 更新验证状态
	if !validOk {
		exist.Status = model.VerifyStatusReject
		log.Debug("认证失败",
			log.FString("needCode", existBody),
			log.FString("requestCode", body),
		)
		return errs.Match2("验证码错误")
	}

	exist.Status = model.VerifyStatusSuccess
	log.Debug("认证成功",
		log.FString("needCode", existBody),
		log.FString("requestCode", body),
	)
	now := time.Now().Unix()
	exist.ValidAt = &now
	exist.ValidTimes++
	return svc.db.Update(exist)
}

// generateBody 生成验证码
func (svc *Verify) generateBody(param *model.Verify) *errs.CodeErrs {
	limit := svc.GetLimitVerify(int16(param.OwnKind), param.OwnID)
	bodyLen := limit.BodyLen[int16(param.AuthKind)]
	body := ""
	switch param.AuthKind {
	case model.AuthKindCellphone: // TODO:GG 也可以是链接？
		body = num.Random(bodyLen)
	case model.AuthKindEmail: // TODO:GG 也可以是链接？
		body = num.Random(bodyLen)
	default:
		return errs.Match2(fmt.Sprintf("不支持的验证类型 kind: %svc", strconv.Itoa(int(param.AuthKind))))
	}
	param.SetBody(&body)
	return nil
}

// addVerify 添加验证码
func (svc *Verify) addVerify(param *model.Verify) *errs.CodeErrs {
	limit := svc.GetLimitVerify(int16(param.OwnKind), param.OwnID)

	// 检查添加间隔时间
	exist, err := svc.db.Select(param) // TODO:GG 根据 OwnKind + OwnID + AuthKind + Apply + Target 查找最近的
	if err != nil {
		return err
	} else if exist != nil {
		// 计算间隔时间，不能小于InsertInterval
		interval := exist.CreateAt - time.Now().UnixMilli()
		if interval < (limit.InsertInterval * 1000) {
			return errs.Match2(fmt.Sprintf("添加间隔时间不能小于 %d", limit.InsertInterval))
		}
	}

	// 检查添加次数
	_ = param.CreateAt - (limit.InsertDuration * 1000)
	count, err := svc.db.SelectCount(param) // TODO:GG 根据 OwnKind + OwnID + AuthKind + Apply + Target + startAt(上面) 查找最近的
	if err != nil {
		return err
	} else if count >= limit.InsertMaxTimes {
		return errs.Match2(fmt.Sprintf("添加次数不能超过 %d", limit.InsertMaxTimes))
	}

	// 添加数据库
	return svc.db.Insert(param)
}

// checkExist 检查验证码是否存在
func (svc *Verify) checkExist(param *model.Verify) (*model.Verify, *errs.CodeErrs) {
	exist, err := svc.db.Select(param) // TODO:GG 根据 OwnKind + OwnID + AuthKind + Apply + Target 查找最近的
	if err != nil {
		return nil, err
	} else if exist == nil {
		return nil, errs.Match2(fmt.Sprintf("没有找到对应的验证码 %d", param.AuthKind))
	}
	limit := svc.GetLimitVerify(int16(exist.OwnKind), exist.OwnID)
	if !exist.CanValid(limit.Expires, limit.VerifyMaxTimes) {
		return nil, errs.Match2("失效的验证码")
	}
	existBody, ok := exist.GetBody()
	if !ok {
		return nil, errs.Match2("验证：没有验证码！")
	} else if len(existBody) <= 0 {
		return nil, errs.Match2("验证：验证码不能为空！")
	}
	return exist, nil
}
