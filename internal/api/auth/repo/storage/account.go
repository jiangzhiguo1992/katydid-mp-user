package storage

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/pkg/storage"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/log"
)

type (
	// Account 账号仓储
	Account struct {
		*storage.Base
	}
)

func NewAccount() *Account {
	return &Account{
		Base: storage.NewBase(),
	}
}

func (sto *Account) Insert(bean *model.Account) *errs.CodeErrs {
	// TODO:GG 雪花ID 10+9 or 分布ID 5+14
	//sto.W.Create(&Account{})
	log.Debug("DB_添加账号", log.FAny("account", bean))
	return nil
}

func (sto *Account) Delete(id uint64, deleteBy *uint64) *errs.CodeErrs {
	return nil
}

func (sto *Account) Update(bean *model.Account) *errs.CodeErrs {
	return nil
}

func (sto *Account) Select(bean *model.Account) (*model.Account, *errs.CodeErrs) {
	return nil, nil
}

func (sto *Account) Selects(bean *model.Account) ([]*model.Account, *errs.CodeErrs) {
	return nil, nil
}

func (sto *Account) SelectCount(bean *model.Account) (int, *errs.CodeErrs) {
	return 1, nil
}
