package db

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/pkg/db"
	"katydid-mp-user/pkg/errs"
)

type (
	// Auth 认证仓储
	Token struct {
		*db.Base
	}
)

func NewToken() *Token {
	return &Token{
		Base: db.NewBase(),
	}
}

func (dbs *Token) Insert(bean model.Token) *errs.CodeErrs {
	// TODO:GG 雪花ID 10+9 or 分布ID 5+14
	//dbs.W.Create(&Token{})
	return nil
}

func (dbs *Token) Delete(id uint64, deleteBy *uint64) *errs.CodeErrs {
	return nil
}

func (dbs *Token) Update(bean model.Token) *errs.CodeErrs {
	return nil
}

func (dbs *Token) Select(bean model.Token) (*model.Token, *errs.CodeErrs) {
	return nil, nil
}

func (dbs *Token) Selects(bean model.Token) ([]*model.Token, *errs.CodeErrs) {
	return nil, nil
}
