package service

import (
	"fmt"
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/api/auth/repo/db"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/errs"
	"strconv"
)

type (
	// Account 账号服务
	Account struct {
		*service.Base

		dbs     *db.Account
		dbsAuth *db.Auth

		//cache *cache.Account
	}
)

func NewAccount(
	db *db.Account, //cache *cache.Account,
) *Account {
	return &Account{
		Base: service.NewBase(nil),
		dbs:  db, // cache: cache,
	}
}

// Register 注册账号 TODO:GG 上层已经验证了verify + 检查ownId是否存在
// TODO:GG 上层生成token并返回 (即使Token生成失败也不要紧，已经注册成功了，直接登录即可)
func (svc *Account) Register(param *model.Account) *errs.CodeErrs {
	// auths检查 (有且仅有一个)
	iAuth := param.FirstAuth()
	if iAuth == nil {
		return errs.Match2("注册时，认证不能为空")
	}

	// 清洗数据
	entity := param.Wash()
	iAuth = iAuth.Wash()

	// number生成
	err := svc.generateNumber(entity)
	if err != nil {
		return err
	}

	// nickname检查
	err = svc.checkNickname(entity)
	if err != nil {
		return err
	}

	// auth检查
	err = svc.checkAuth(entity, iAuth)
	if err != nil {
		return err
	}
	return nil
}

// UnRegister 注销账号
func (svc *Account) UnRegister(exist *model.Account) *errs.CodeErrs {
	// TODO:GG 需要解绑各种auths吗？
	// TODO:GG 解绑了的话，就找不回来了！除非有password
	// TODO:GG 不解绑的话，auth就不能注册新账号
	return nil
}

// Login 登录账号
func (svc *Account) Login(param *model.Account) *errs.CodeErrs {
	limit := svc.GetLimitAccount(int16(param.OwnKind), param.OwnID)
	_ = limit.AuthLogins
	_ = limit.AuthRequires
	_ = limit.UserIDCardRequire
	_ = limit.UserInfoRequire
	_ = limit.UserBioRequire

	_ = svc.checkActionLogin(param)

	if param.Status == model.AccountStatusLocked {
		// TODO:GG 解锁
	}

	_ = param.GetAuthKinds()

	return nil
}

//case AuthKindPassword:
// TODO:GG 在auth里检查?
// TODO: 比较 hashedPassword 和 a.Password
// 实际场景中应该使用安全的密码哈希比较
// 例如: hashedPassword := HashPassword(cred.Password, salt)
//salt, _ := a.GetPasswordSalt()

// checkNickname 检查昵称
func (svc *Account) checkNickname(entity *model.Account) *errs.CodeErrs {
	limit := svc.GetLimitAccount(int16(entity.OwnKind), entity.OwnID)
	if !limit.NicknameRequire {
		entity.Nickname = nil
		return nil
	} else if entity.Nickname == nil {
		return errs.Match2("昵称不能为空")
	}

	// 查重
	if !limit.NicknameUnique {
		return nil
	}
	exist, err := svc.dbs.Select(entity) // TODO:GG 根据 OwnKind + OwnId + nickname 查找account
	if err != nil {
		return err
	} else if exist != nil {
		return errs.Match2("昵称已存在")
	}
	return nil
}

// generateNumber 生成账号标识
func (svc *Account) generateNumber(entity *model.Account) *errs.CodeErrs {
	// TODO:GG 生成账号标识
	return nil
}

// checkAuth 检查认证
func (svc *Account) checkAuth(entity *model.Account, iAuth model.IAuth) *errs.CodeErrs {
	authKind := iAuth.GetKind()

	// 检查是否是必要的AuthKind
	if !svc.isAuthKindRequire(entity, authKind) {
		return errs.Match2(fmt.Sprintf("不是必须的认证方式 kind: %svc", strconv.Itoa(int(authKind))))
	}

	// 查重，固定成1了，同own下，account和auth是一对一的关系
	exist, err := svc.dbs.Select(entity) // TODO:GG 根据 ownKind + ownID + authKind + target 查找accounts
	if err != nil {
		return err
	} else if exist != nil {
		if !exist.IsUnRegister() || !exist.CanRegister() {
			return errs.Match2("账号已存在")
		}
	}

	// 根据authKind来检查
	switch authKind {
	case model.AuthKindPassword:
		// 查重，AuthKindPassword的username只能是owner里唯一的
		// 不检查TokenShares了，只有share里有的，这里也可以注册
		existAuth, err := svc.dbsAuth.Select(iAuth) // TODO:GG 根据 OwnKind + OwnId + Username 查找auth
		if err != nil {
			return err
		} else if existAuth != nil {
			return errs.Match2("用户名已存在")
		}

		// 添加/重新注册账号
		if exist == nil {
			err = svc.dbs.Insert(entity)
			exist = entity
		} else {
			err = svc.dbs.Update(exist)
		}
		if err != nil {
			return err
		}

		// 添加关联的auth
		err = svc.dbsAuth.Insert(iAuth) // TODO:GG 这里是多对多表的修改
		if err != nil {
			return err
		}

		// 修改实体类绑定
		if exist.AddAuth(iAuth) {
			// 修改account状态
			err = svc.dbs.Update(exist) // TODO:GG 更新账号状态
			if err != nil {
				return err
			}
		}

	case model.AuthKindCellphone,
		model.AuthKindEmail:
		// TODO:GG 上层进行过auth的verify认证了 (如果limit.VerifyRegister=true的话)
		// 查重，相同的auth只能有一个(全局),pwd除外
		existAuth, err := svc.dbsAuth.Select(iAuth) // TODO:GG 根据 number/email/openID/... 查找auth
		if err != nil {
			return err
		} else if (existAuth != nil) && !existAuth.IsEnabled() {
			return errs.Match2("认证不可用")
		}

		// 添加/重新注册账号
		if exist == nil {
			err = svc.dbs.Insert(entity)
			exist = entity
		} else {
			err = svc.dbs.Update(exist)
		}
		if err != nil {
			return err
		}

		// auth的是否首次注册
		if existAuth == nil {
			// 添加关联的auth
			err = svc.dbsAuth.Insert(iAuth) // TODO:GG 这里是多对多表的修改
			existAuth = iAuth
		} else {
			// 更新关联auth
			err = svc.dbsAuth.Update(existAuth) // TODO:GG 更新accountID 这里是多对多表的修改
		}
		if err != nil {
			return err
		}

		// 修改实体类绑定
		if exist.AddAuth(existAuth) {
			// 修改account状态
			err = svc.dbs.Update(exist) // TODO:GG 更新账号状态
			if err != nil {
				return err
			}
		}

	default:
		return errs.Match2(fmt.Sprintf("不支持的认证方式 kind: %svc", strconv.Itoa(int(authKind))))
	}
	return nil
}

// isAuthKindRequire 检查是否是必要的AuthKind
func (svc *Account) isAuthKindRequire(param *model.Account, authKind model.AuthKind) bool {
	limit := svc.GetLimitAccount(int16(param.OwnKind), param.OwnID)
	if len(limit.AuthRequires) <= 0 {
		return true // 没有则都可以
	}
	for _, groupKind := range limit.AuthRequires {
		for _, enableKind := range groupKind {
			if authKind == model.AuthKind(enableKind) {
				return true
			}
		}
	}
	return false
}

// isAuthKindEnable 检查是否是支持的AuthKind
func (svc *Account) isAuthKindEnable(param *model.Account, authKind model.AuthKind) bool {
	limit := svc.GetLimitAccount(int16(param.OwnKind), param.OwnID)
	for _, enableKind := range limit.AuthEnables {
		if authKind == model.AuthKind(enableKind) {
			return true
		}
	}
	return false
}

// isAuthKindLogin 检查是否是登录的AuthKind
func (svc *Account) isAuthKindLogin(param *model.Account, authKind model.AuthKind) bool {
	limit := svc.GetLimitAccount(int16(param.OwnKind), param.OwnID)
	for _, loginKind := range limit.AuthLogins {
		if authKind == model.AuthKind(loginKind) {
			return true
		}
	}
	return false
}

// checkActionLogin 检查登录行为合法性
func (svc *Account) checkActionLogin(exist *model.Account) *errs.CodeErrs {
	if exist.CanLogin() {
		return nil
	}
	if exist.IsUnRegister() {
		if exist.CanRegister() {
			return errs.Match2("账号不存在")
		} else {
			return errs.Match2("账号被拉黑")
		}
	} else {
		return errs.Match2("账号暂时被锁定")
	}
}
