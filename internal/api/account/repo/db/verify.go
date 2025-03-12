package db

import (
	"katydid-mp-user/internal/api/account/model"
	"katydid-mp-user/internal/pkg/db"
	"katydid-mp-user/pkg/errs"
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

func (v *Verify) Insert(body *model.Verify) *errs.CodeErrs {
	// TODO:GG 雪花ID 10+9
	// TODO:GG 分布ID 5+14
	//v.W.Create(&Verify{})
	return nil
}

func (v *Verify) Delete(id uint64, deleteBy *uint64) *errs.CodeErrs {
	return nil
}

func (v *Verify) Update(body *model.Verify) *errs.CodeErrs {
	return nil
}

func (v *Verify) Select(body *model.Verify) (*model.Verify, *errs.CodeErrs) {
	return nil, nil
}

func (v *Verify) Selects(body *model.Verify) ([]*model.Verify, *errs.CodeErrs) {
	return nil, nil
}
