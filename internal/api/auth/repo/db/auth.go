package db

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/pkg/db"
	"katydid-mp-user/pkg/errs"
)

type (
	// Auth 认证仓储
	Auth struct {
		*db.Base
	}
)

func NewAuth() *Auth {
	return &Auth{
		Base: db.NewBase(),
	}
}

func (sql *Auth) Insert(bean model.IAuth) *errs.CodeErrs {
	// TODO:GG 雪花ID 10+9 or 分布ID 5+14
	//sql.W.Create(&Auth{})
	return nil
}

func (sql *Auth) Delete(id uint64, deleteBy *uint64) *errs.CodeErrs {
	return nil
}

func (sql *Auth) Update(bean model.IAuth) *errs.CodeErrs {
	return nil
}

func (sql *Auth) Select(bean model.IAuth) (model.IAuth, *errs.CodeErrs) {
	return nil, nil
}

func (sql *Auth) Selects(bean model.IAuth) ([]model.IAuth, *errs.CodeErrs) {
	return nil, nil
}

func (sql *Auth) SelectCount(bean model.IAuth) (int, *errs.CodeErrs) {
	return 1, nil
}

func (sql *Auth) Login() {
	// TODO:GG 登录的时候，也是先看有没有当前的account，有就校验密码/验证码，没有就校验share的(激活其他平台)

}
