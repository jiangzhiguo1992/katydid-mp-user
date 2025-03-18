package service

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/api/auth/repo/db"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/errs"
)

type (
	// Auth 认证服务
	Auth struct {
		*service.Base

		db        *db.Auth
		dbAccount *db.Account
	}
)

// Add 添加认证
func (svc *Auth) Add(param model.IAuth) *errs.CodeErrs {
	// 查重
	exist, err := svc.db.Selects(param) // TODO:GG 根据 AuthKind + Target 查找Auth
	if err != nil {
		return err
	} else if exist != nil {
		return errs.Match2("认证：已存在！")
	}

	// 记录+清洗数据
	accounts := param.GetAccAccounts()
	for _, owns := range accounts {
		for _, acc := range owns {
			param.DelAccount(acc)
		}
	}
	userID := param.GetUserID()
	param.SetUserID(nil)

	// 数据库添加 (后面的关联是多对多，需要先insertAuth)
	err = svc.db.Insert(param)
	if err != nil {
		return err
	}

	// 检查状态
	err = svc.checkStatus(param)
	if err != nil {
		return err
	}

	// 绑定账号
	for _, owns := range accounts {
		for _, acc := range owns {
			err = svc.BindAccountID(param, acc.ID, true)
			if err != nil {
				return err
			}
		}
	}

	// 绑定用户
	if userID != nil {
		err = svc.BindUserID(param, *userID, true)
	}
	return err
}

// BindAccountID 绑定账号ID
func (svc *Auth) BindAccountID(param model.IAuth, accountID uint64, force bool) *errs.CodeErrs {
	account, err := svc.dbAccount.Select(nil) // TODO:GG 根据ID查找account
	if err != nil {
		return err
	} else if account == nil {
		return errs.Match2("账号不存在")
	}
	return svc.BindAccount(param, account, force)
}

// BindAccount 绑定账号 TODO:GG 外部需要 account.AddAuth(a)
func (svc *Auth) BindAccount(param model.IAuth, account *model.Account, force bool) *errs.CodeErrs {
	// 检查账号是否已绑定
	oldBind := param.GetAccount(account.OwnKind, account.OwnID)
	if (oldBind != nil) && !force {
		return errs.Match2("已绑定账号")
	}

	// 更新auth下的account关联 (多对多表修改)
	// TODO:GG 更新外表accountID，新旧都会改，需要在这里更新吗？
	// TODO:GG 被绑定的account也需要修改auth，auth只同时bind一个(当前own下)账号
	param.SetAccount(account)

	// 更新auth的状态
	if param.IsActive() && !param.IsBind() {
		param.SetStatus(model.AuthStatusBind)
		err := svc.db.Update(param) // TODO:GG 更新status
		if err != nil {
			return err
		}
	}
	return nil
}

// UnbindAccountID 解绑账号ID
func (svc *Auth) UnbindAccountID(param model.IAuth, accountID uint64) *errs.CodeErrs {
	account, err := svc.dbAccount.Select(nil) // TODO:GG 根据ID查找account
	if err != nil {
		return err
	} else if account == nil {
		return errs.Match2("账号不存在")
	}
	return svc.UnbindAccount(param, account)
}

// UnbindAccount 解绑账号 TODO:GG 外部需要 account.DelAuth(a)
func (svc *Auth) UnbindAccount(param model.IAuth, account *model.Account) *errs.CodeErrs {
	// 检查账号是否已绑定
	oldBind := param.GetAccount(account.OwnKind, account.OwnID)
	if oldBind == nil {
		return errs.Match2("未绑定账号")
	}

	// 更新auth下的account关联 (多对多表修改)
	// TODO:GG 更新外表accountID，新旧都会改，需要在这里更新吗？
	// TODO:GG 被绑定的account也需要修改auth，auth只同时bind一个(当前own下)账号
	param.DelAccount(account)

	// 更新auth的状态
	if param.IsBind() && len(param.GetAccAccounts()) == 0 {
		param.SetStatus(model.AuthStatusActive)
		err := svc.db.Update(param) // TODO:GG 更新status
		if err != nil {
			return err
		}
	}
	return nil
}

// BindUserID 绑定用户ID
func (svc *Auth) BindUserID(param model.IAuth, userID uint64, force bool) *errs.CodeErrs {
	user, err := svc.dbAccount.Select(nil) // TODO:GG 根据ID查找user
	if err != nil {
		return errs.Match2("用户不存在")
	} else if user == nil {
		return errs.Match2("用户不存在")
	}
	return svc.BindUser(param, user, force)
}

// BindUser 绑定用户
func (svc *Auth) BindUser(param model.IAuth, user any, force bool) *errs.CodeErrs {
	limit := svc.GetLimitAuth(int16(account.OwnKind), account.OwnID)

	// TODO:GG
	return nil
}

func (svc *Auth) checkStatus(param model.IAuth) *errs.CodeErrs {

}

//case AuthKindPassword:
// TODO:GG 在auth里检查?
// TODO: 比较 hashedPassword 和 a.Password
// 实际场景中应该使用安全的密码哈希比较
// 例如: hashedPassword := HashPassword(cred.Password, salt)
//salt, _ := a.GetPasswordSalt()
