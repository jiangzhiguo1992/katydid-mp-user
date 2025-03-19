package storage

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/pkg/storage"
	"katydid-mp-user/pkg/errs"
)

type (
	// Auth 认证仓储
	Auth struct {
		*storage.Base
	}
)

func NewAuth() *Auth {
	return &Auth{
		Base: storage.NewBase(),
	}
}

func (sto *Auth) Insert(bean model.IAuth) *errs.CodeErrs {
	// TODO:GG 雪花ID 10+9 or 分布ID 5+14
	//sto.W.Create(&Auth{})
	return nil
}

func (sto *Auth) Delete(id uint64, deleteBy *uint64) *errs.CodeErrs {
	return nil
}

func (sto *Auth) Update(bean model.IAuth) *errs.CodeErrs {
	return nil
}

func (sto *Auth) Select(bean model.IAuth) (model.IAuth, *errs.CodeErrs) {
	return nil, nil
}

func (sto *Auth) Selects(bean model.IAuth) ([]model.IAuth, *errs.CodeErrs) {
	return nil, nil
}

func (sto *Auth) SelectCount(bean model.IAuth) (int, *errs.CodeErrs) {
	return 1, nil
}

func (sto *Auth) Login() {
	// TODO:GG 登录的时候，也是先看有没有当前的account，有就校验密码/验证码，没有就校验share的(激活其他平台)

}
