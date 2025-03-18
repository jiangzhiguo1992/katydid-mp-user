package db

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/pkg/db"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/log"
)

type (
	// Verify 验证码仓储
	Verify struct {
		*db.Base
	}
)

func NewVerify() *Verify {
	return &Verify{
		Base: db.NewBase(),
	}
}

func (dbs *Verify) Insert(bean *model.Verify) *errs.CodeErrs {
	// TODO:GG 雪花ID 10+9 or 分布ID 5+14
	//d.W.Create(&Verify{})
	log.Debug("DB_添加验证", log.FAny("verify", bean))
	return nil
}

func (dbs *Verify) Delete(id uint64, deleteBy *uint64) *errs.CodeErrs {
	log.Debug("DB_删除验证", log.FAny("verify", id))
	return nil
}

func (dbs *Verify) Update(bean *model.Verify) *errs.CodeErrs {
	log.Debug("DB_修改验证", log.FAny("verify", bean))
	return nil
}

func (dbs *Verify) Select(bean *model.Verify) (*model.Verify, *errs.CodeErrs) {
	log.Debug("DB_获取验证", log.FAny("verify", bean))
	return bean, nil
}

func (dbs *Verify) Selects(bean *model.Verify) ([]*model.Verify, *errs.CodeErrs) {
	return nil, nil
}

func (dbs *Verify) SelectCount(bean *model.Verify) (int, *errs.CodeErrs) {
	return 0, nil
}
