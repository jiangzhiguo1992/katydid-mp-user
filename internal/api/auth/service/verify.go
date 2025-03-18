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

// Add 添加验证码
func (svc *Verify) Add(param *model.Verify) *errs.CodeErrs {
	entity := param.Wash()

	// TODO:GG 检查ownId是否存在

	// 生成验证码
	err := svc.generateBody(entity)
	if err != nil {
		return err
	}
	return svc.addWithCheck(entity)
}

func (svc *Verify) Del(param *model.Verify) *errs.CodeErrs {
	// TODO:GG DB删除
	return nil
}

// OnSendOk 发送验证码成功
func (svc *Verify) OnSendOk(exist *model.Verify) *errs.CodeErrs {
	// 不检查ownerID了
	exist.SetPending()
	return svc.db.Update(exist)
}

// OnSendFail 发送验证码失败
func (svc *Verify) OnSendFail(exist *model.Verify) *errs.CodeErrs {
	// 不检查ownerID了
	return nil
}

// Valid 验证验证码
func (svc *Verify) Valid(param *model.Verify) *errs.CodeErrs {
	// TODO:GG 检查ownId是否存在

	// 检查传进来的参数
	body, ok := param.GetBody()
	if !ok {
		return errs.Match2("验证：没有验证码！")
	} else if len(body) <= 0 {
		return errs.Match2("验证：验证码不能为空！")
	}

	// 检查已存在验证码的合法性
	exist, err := svc.checkExist(param)
	if err != nil {
		return err
	}
	existBody, ok := exist.GetBody()

	// 验证内容体
	validOk := exist.Valid(body)

	// 更新验证结果
	if validOk {
		exist.SetSuccess()
	} else {
		exist.SetReject()
	}
	err = svc.db.Update(exist)
	if err != nil {
		return err
	}
	if !validOk {
		log.Debug("认证失败",
			log.FString("needCode", existBody),
			log.FString("requestCode", body),
		)
		return errs.Match2("验证码错误")
	}

	// 检查auth是否存在
	existAuth, err := svc.dbAuth.Select(nil) // TODO:GG 根据 AuthKind + Target 查找Auth
	if (err != nil) || (existAuth == nil) {
		return nil
	}
	// 更新auth的状态
	if existAuth.TryActive() {
		_ = svc.dbAuth.Update(nil) // TODO:GG 更新auth的status
	}
	return nil
}

// generateBody 生成验证码
func (svc *Verify) generateBody(entity *model.Verify) *errs.CodeErrs {
	limit := svc.GetLimitVerify(int16(entity.OwnKind), entity.OwnID)
	bodyLen := limit.BodyLen[int16(entity.AuthKind)]

	body := ""
	switch entity.AuthKind {
	case model.AuthKindCellphone:
		body = num.Random(bodyLen)
	case model.AuthKindEmail: // TODO:GG 也可以是链接？
		body = num.Random(bodyLen)
	default:
		return errs.Match2(fmt.Sprintf("不支持的验证类型 kind: %svc", strconv.Itoa(int(entity.AuthKind))))
	}
	entity.SetBody(&body)
	return nil
}

// addWithCheck 添加验证码
func (svc *Verify) addWithCheck(entity *model.Verify) *errs.CodeErrs {
	limit := svc.GetLimitVerify(int16(entity.OwnKind), entity.OwnID)

	// 检查添加间隔时间
	exist, err := svc.db.Select(entity) // TODO:GG 根据 OwnKind + OwnID + AuthKind + Apply + Target 查找最近的
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
	_ = time.Now().UnixMilli() - (limit.InsertDuration * 1000)
	count, err := svc.db.SelectCount(entity) // TODO:GG 根据 OwnKind + OwnID + AuthKind + Apply + Target + time(上面) 查找最近的
	if err != nil {
		return err
	} else if count >= limit.InsertMaxTimes {
		return errs.Match2(fmt.Sprintf("添加次数不能超过 %d", limit.InsertMaxTimes))
	}

	// 添加数据库
	return svc.db.Insert(entity)
}

// checkExist 检查验证码是否存在
func (svc *Verify) checkExist(param *model.Verify) (*model.Verify, *errs.CodeErrs) {
	// 查找验证码
	exist, err := svc.db.Select(param) // TODO:GG 根据 OwnKind + OwnID + AuthKind + Apply + Target 查找最近的
	if err != nil {
		return nil, err
	} else if exist == nil {
		return nil, errs.Match2(fmt.Sprintf("没有找到对应的验证码 %d", param.AuthKind))
	}

	// 检查验证码是否有效
	limit := svc.GetLimitVerify(int16(exist.OwnKind), exist.OwnID)
	if !exist.CanValid(limit.Expires, limit.VerifyMaxTimes) {
		return nil, errs.Match2("失效的验证码")
	}

	// 检查验证内容
	existBody, ok := exist.GetBody()
	if !ok {
		return nil, errs.Match2("验证：没有验证码！")
	} else if len(existBody) <= 0 {
		return nil, errs.Match2("验证：验证码不能为空！")
	}
	return exist, nil
}
