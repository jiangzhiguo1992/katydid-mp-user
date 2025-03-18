package db

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/pkg/db"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/log"
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

func (sql *Account) Insert(bean *model.Account) *errs.CodeErrs {
	// TODO:GG 雪花ID 10+9 or 分布ID 5+14
	//sql.W.Create(&Account{})
	log.Debug("DB_添加账号", log.FAny("account", bean))
	return nil
}

func (sql *Account) Delete(id uint64, deleteBy *uint64) *errs.CodeErrs {
	return nil
}

func (sql *Account) Update(bean *model.Account) *errs.CodeErrs {
	return nil
}

func (sql *Account) Select(bean *model.Account) (*model.Account, *errs.CodeErrs) {
	return nil, nil
}

func (sql *Account) Selects(bean *model.Account) ([]*model.Account, *errs.CodeErrs) {
	return nil, nil
}

func (sql *Account) SelectCount(bean *model.Account) (int, *errs.CodeErrs) {
	return 1, nil
}
