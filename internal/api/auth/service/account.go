package service

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/api/auth/repo/cache"
	"katydid-mp-user/internal/api/auth/repo/db"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/log"
)

type (
	// Account 账号服务
	Account struct {
		*service.Base

		db    *db.Account
		cache *cache.Account

		IsOwnerAuthEnable func(ownKind model.TokenOwn, ownID uint64, kind model.AuthKind) (bool, *errs.CodeErrs)
		GetMaxNumByOwner  func(ownKind model.TokenOwn, ownID uint64) (int, *errs.CodeErrs)
	}
)

func NewAccount(
	db *db.Account, cache *cache.Account,
	isOwnerAuthEnable func(ownKind model.TokenOwn, ownID uint64, kind model.AuthKind) (bool, *errs.CodeErrs),
	getMaxNumByOwner func(ownKind model.TokenOwn, ownID uint64) (int, *errs.CodeErrs),
) *Account {
	return &Account{
		Base: service.NewBase(nil),
		db:   db, cache: cache,
		IsOwnerAuthEnable: isOwnerAuthEnable,
		GetMaxNumByOwner:  getMaxNumByOwner,
	}
}

func (as *Account) AddAccount(
	ownKind model.TokenOwn, ownID uint64,
	userID *uint64, nickname string,
	auth model.IAuth,
) (*model.Account, *errs.CodeErrs) {
	// TODO:GG 检查client的limit

	// TODO:GG 是否可认证和必须认证的方式，由org/client决定
	// TODO:GG 存id？还是存带extra里？

	// TODO:GG 需要检查user最大数量
	if userID != nil {
		maxCount, err := as.GetMaxNumByOwner(ownKind, ownID)
		if err != nil {
			return nil, err
		}
		exitsNum, err := as.db.SelectCount()
		if err != nil {
			return nil, err
		}
		if (maxCount >= 0) && (exitsNum >= maxCount) {
			return nil, errs.Match2("账号数量超过限制")
		}
	}

	// 检查可否认证方式
	enable, err := as.IsOwnerAuthEnable(ownKind, ownID, auth.GetKind())
	if err != nil {
		return nil, err
	} else if !enable {
		return nil, errs.Match2("认证方式不可用")
	}

	add := model.NewAccount(userID, nickname)

	log.Debug("DB_添加账号", log.FAny("account", add))

	return add, nil
}

//func (as *AccountService) DelAccount(instance *model.Account) *errs.CodeErrs {
//
//}
//
//func (as *AccountService) UpdAccountNickName(instance *model.Account) *errs.CodeErrs {
//
//}
//
//func (as *AccountService) UpdateAccountTokens(instance *model.Account) *errs.CodeErrs {
//
//}
//
//func (as *AccountService) UpdAccountAuthAdd(instance *model.Account, auth *model.Auth) *errs.CodeErrs {
//
//}
//
//func (as *AccountService) UpdAccountAuthDel(instance *model.Account, auth *model.Auth) *errs.CodeErrs {
//
//}
//
//func (as *AccountService) UpdAccountAuthUpd(instance *model.Account, auth *model.Auth) *errs.CodeErrs {
//
//}
//
//func (as *AccountService) GetAccount(id uint64) (*model.Account, *errs.CodeErrs) {
//
//}
