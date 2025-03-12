package db

import (
	"katydid-mp-user/internal/pkg/db"
	"katydid-mp-user/pkg/errs"
)

type (
	// Account 账号仓储
	Account struct {
		*db.Base
	}
)

func NewAccount() *Account {
	return &Account{
		Base: db.NewBase(),
	}
}

func (a *Account) Insert() {
	// TODO:GG 雪花ID 10+9
	// TODO:GG 分布ID 5+14
	//a.W.Create(&Account{})
}

func (a *Account) Delete() {
}

func (a *Account) Update() {
}

func (a *Account) Select() {
}

func (a *Account) SelectCount() (int, *errs.CodeErrs) {
	return 1, nil
}
