package service

import (
	"katydid-mp-user/internal/api/account/model"
	"katydid-mp-user/pkg/err"
)

func AddAccount(userId uint64) (*model.Account, *err.CodeErrs) {

	// TODO:GG 检查client的limit

	account := model.NewAccountDef(userId)
	return account, nil
}
