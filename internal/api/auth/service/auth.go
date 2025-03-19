package service

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/api/auth/repo/storage"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/errs"
)

type (
	// Auth 认证服务
	Auth struct {
		*service.Base

		dbs        *storage.Auth
		dbsAccount *storage.Account
		dbsVerify  *storage.Verify
	}
)

// BindAccounts 添加认证并绑定账号
func (svc *Auth) BindAccounts(param model.IAuth) *errs.CodeErrs {
	// 记录+清洗数据
	accounts := param.GetAccAccounts() // TODO:GG token里的的account(exist)填充到auth里
	entity := param.Wash()

	// 历史记录
	exist, err := svc.dbs.Select(param) // TODO:GG 根据 AuthKind + Target 查找Auth
	if err != nil {
		return err
	} else if exist == nil {
		// 数据库添加 (后面的关联是多对多，需要先insertAuth)
		entity.SetStatus(model.AuthStatusActive) // TODO:GG 上层需要验证verify
		err = svc.dbs.Insert(entity)
		if err != nil {
			return err
		}
		exist = entity
	} else {
		// 检查状态
		err = svc.tryStatusActive(exist)
		if err != nil {
			return err
		}
	}

	// 检查状态
	if !exist.IsEnabled() {
		return errs.Match2("认证不可用")
	}

	// 绑定账号
	if len(accounts) < 0 {
		return errs.Match2("账号不能为空")
	}
	for _, owns := range accounts {
		for _, acc := range owns {
			err = svc.bindAccount(exist, acc)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// BindAccount 绑定账号
func (svc *Auth) bindAccount(exist model.IAuth, account *model.Account) *errs.CodeErrs {
	// 检查账号是否已绑定 TODO:GG 如果不是强一致性，则查多对多表
	oldBindAccount := exist.GetAccount(account.OwnKind, account.OwnID)
	if oldBindAccount != nil {
		limit := svc.GetLimitAccount(int16(account.OwnKind), account.OwnID)
		for _, authKind := range limit.AuthLogins {
			if int16(exist.GetKind()) == authKind {
				return errs.Match2("此认证已绑定账号，请先上对应的账号解绑")
			}
		}
	}

	// 更新auth下的account关联 (多对多表修改)
	// TODO:GG 更新外表accountID，新旧都会改，需要在这里更新吗？
	// TODO:GG 被绑定的account也需要修改auth，auth只同时bind一个(当前own下)账号

	// 修改实体类绑定
	if oldBindAccount != nil {
		delete(oldBindAccount.Auths, exist.GetKind())
	}
	if oldBindAuth := account.Auths[exist.GetKind()]; oldBindAuth != nil {
		delete(account.Auths, oldBindAuth.GetKind())
	}
	if account.AddAuth(exist) {
		// 修改account状态
		err := svc.dbsAccount.Update(account) // TODO:GG 更新账号状态
		if err != nil {
			return err
		}
	}

	// 更新auth的状态
	return svc.tryStatusBind(exist)
}

// UnbindAccount 解绑账号
func (svc *Auth) UnbindAccount(param model.IAuth, account *model.Account) *errs.CodeErrs {
	exist, err := svc.dbs.Select(param) // TODO:GG 上层需要验证verify
	if err != nil {
		return err
	} else if exist == nil {
		return errs.Match2("认证不存在")
	} else if !exist.IsEnabled() {
		return errs.Match2("认证不可用")
	}

	// 检查账号是否已绑定 TODO:GG 如果不是强一致性，则查多对多表
	if exist.GetAccount(account.OwnKind, account.OwnID) == nil {
		return errs.Match2("未绑定账号")
	}

	isLoginAUth := false
	limit := svc.GetLimitAccount(int16(account.OwnKind), account.OwnID)
	for _, loginKind := range limit.AuthLogins {
		if int16(exist.GetKind()) == loginKind {
			isLoginAUth = true
			break
		}
	}
	if isLoginAUth {
		existLoginAuth := 0
		for _, loginKind := range limit.AuthLogins {
			if account.HasAuthKind(model.AuthKind(loginKind)) {
				existLoginAuth++
			}
		}
		if existLoginAuth <= 1 {
			return errs.Match2("最后一个登录认证不能解绑，只能更换绑定或者注销")
		}
	}

	// 更新auth下的account关联 (多对多表修改)
	// TODO:GG 更新外表accountID，新旧都会改，需要在这里更新吗？
	// TODO:GG 被绑定的account也需要修改auth，auth只同时bind一个(当前own下)账号(limit固定1)

	// 修改实体类绑定
	if account.DelAuth(exist) {
		// 修改account状态
		err := svc.dbsAccount.Update(account) // TODO:GG 更新账号状态
		if err != nil {
			return err
		}
	}

	// 更新auth的状态
	return svc.tryStatusActive(exist)
}

// ResetPassword 修改密码
func (svc *Auth) ResetPassword(param *model.AuthPassword) *errs.CodeErrs {
	// TODO:GG 修改密码
	return nil
}

// tryStatusActive 尝试修改成激活状态
func (svc *Auth) tryStatusActive(exist model.IAuth) *errs.CodeErrs {
	// 检查是否满足更新条件
	update := false
	if exist.IsEnabled() && !exist.IsActive() {
		existVerify, err := svc.dbsVerify.Select(nil) // TODO:GG 根据 StatusSuccess + AuthKind + Target 查找最近的
		if err != nil {
			return err
		} else if existVerify == nil {
			return nil
		}
		update = true
	} else if exist.IsBind() && len(exist.GetAccAccounts()) == 0 {
		// TODO:GG 除非GetAccAccounts具有一致性(TX事务)，否则要在多对多的表里查accounts
		update = true
	}
	// active不会回溯，除非拉黑

	// 更新auth的状态
	if update {
		exist.SetStatus(model.AuthStatusActive)
		err := svc.dbs.Update(exist) // TODO:GG 更新status
		if err != nil {
			return err
		}
	}
	return nil
}

// tryStatusBind 尝试修改成绑定状态
func (svc *Auth) tryStatusBind(exist model.IAuth) *errs.CodeErrs {
	// 先检查是否active
	err := svc.tryStatusActive(exist)
	if err != nil {
		return err
	}

	// 检查是否满足更新条件
	update := false
	if exist.IsActive() && !exist.IsBind() && (len(exist.GetAccAccounts()) > 0) {
		// TODO:GG 除非GetAccAccounts具有一致性(TX事务)，否则要在多对多的表里查accounts
		update = true
	}

	// 更新auth的状态
	if update {
		exist.SetStatus(model.AuthStatusBind)
		err = svc.dbs.Update(exist) // TODO:GG 更新status
		if err != nil {
			return err
		}
	}
	return nil
}
