package service

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/data"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/log"
)

type (
	// Auth 认证服务
	Auth struct {
		*service.Base
	}
)

func (a *Auth) Add(auth model.IAuth) (model.IAuth, *errs.CodeErrs) {
	added := auth
	switch auth.GetKind() {
	case model.AuthKindPassword:
		if _, ok := auth.(*model.AuthPassword); !ok {
			return nil, errs.Match2("认证：密码类型错误！")
		}
		added = auth.(*model.AuthPassword)
		log.Debug("DB_添加认证1", log.FAny("auth", added))
	case model.AuthKindPhone:
		if _, ok := auth.(*model.AuthPhone); !ok {
			return nil, errs.Match2("认证：手机号类型错误！")
		}
		added = auth.(*model.AuthPhone)
		log.Debug("DB_添加认证2", log.FAny("auth", added))
	case model.AuthKindEmail:
		if _, ok := auth.(*model.AuthEmail); !ok {
			return nil, errs.Match2("认证：邮箱类型错误！")
		}
		added = auth.(*model.AuthEmail)
		log.Debug("DB_添加认证3", log.FAny("auth", added))
	case model.AuthKindBiometric:
		if _, ok := auth.(*model.AuthBiometric); !ok {
			return nil, errs.Match2("认证：生物特征类型错误！")
		}
		added = auth.(*model.AuthBiometric)
		log.Debug("DB_添加认证4", log.FAny("auth", added))
	case model.AuthKindThirdParty:
		if _, ok := auth.(*model.AuthThirdParty); !ok {
			return nil, errs.Match2("认证：第三方类型错误！")
		}
		added = auth.(*model.AuthThirdParty)
		log.Debug("DB_添加认证5", log.FAny("auth", added))
	default:
		return nil, errs.Match2("认证：未知类型错误！")
	}
	return added, nil
}

func (a *Auth) Del(instance *model.Auth) *errs.CodeErrs {
	return nil
}

func (a *Auth) Upd(instance *model.Auth) *errs.CodeErrs {
	return nil
}

func (a *Auth) Get(kind uint16, maps data.KSMap) (model.IAuth, *errs.CodeErrs) {
	//	switch kind {
	//	case model.AuthKindPwd:
	//		if username, ok := maps.GetString("username"); ok {
	//			// TODO:GG DB获取
	//			return &model.UsernamePwd{
	//				Auth:     &model.Auth{},
	//				Username: username,
	//			}, nil
	//		}
	//		return nil, errs.MatchByMessage("认证：没有账号！")
	//	case model.AuthKindPhone:
	//		if areaCode, ok := maps.GetString("areaCode"); ok {
	//			if number, okk := maps.GetString("number"); okk {
	//				// TODO:GG DB获取
	//				return &model.PhoneNumber{
	//					Auth:     &model.Auth{},
	//					AreaCode: areaCode,
	//					Number:   number,
	//				}, nil
	//			} else {
	//				return nil, errs.MatchByMessage("认证：没有手机号！")
	//			}
	//		}
	//		return nil, errs.MatchByMessage("认证：没有区号！")
	//	case model.AuthKindEmail:
	//		if username, ok := maps.GetString("username"); ok {
	//			if domain, okk := maps.GetString("domain"); okk {
	//				// TODO:GG DB获取
	//				return &model.EmailAddress{
	//					Auth:     &model.Auth{},
	//					Username: username,
	//					Domain:   domain,
	//				}, nil
	//			} else {
	//				return nil, errs.MatchByMessage("认证：没有域名！")
	//			}
	//		}
	//		return nil, errs.MatchByMessage("认证：没有用户名！")
	//	case model.AuthKindBio:
	//		return nil, errs.NewErr(errs.MatchByMessage("生物特征认证未实现！"))
	//	case model.AuthKindThird:
	//		return nil, errs.NewErr(errs.MatchByMessage("第三方认证未实现！"))
	//	}
	//	return nil, errs.NewErr(errs.MatchByMessage("未知认证类型！"))
	//}
	//
	//func AuthCheck(instance model.IAuth, kind int16, maps utils.KSMap) *errs.CodeErrs {
	//	switch kind {
	//	case model.VerifyKindPwd:
	//		auth := (instance).(*model.UsernamePwd)
	//		// 账号不是必选项
	//		if username, ok := maps.GetString("username"); ok {
	//			if username != auth.Username {
	//				return errs.NewErr(errs.MatchByMessage("认证：账号错误！"))
	//			}
	//		}
	//		// 密码是必选项
	//		password, ok := maps.GetString("password")
	//		if !ok {
	//			return errs.NewErr(errs.MatchByMessage("认证：没有密码！"))
	//		} else if password != auth.Password {
	//			return errs.NewErr(errs.MatchByMessage("认证：密码错误！"))
	//		}
	//	case model.VerifyKindPhone:
	//		auth := (instance).(*model.PhoneNumber)
	//		code, ok := maps.GetString("code")
	//		if !ok {
	//			return errs.NewErr(errs.MatchByMessage("认证：没有验证码！"))
	//		}
	//		clientId, ok := maps.GetUint64("clientId")
	//		if !ok {
	//			return errs.NewErr(errs.MatchByMessage("认证：未知客户端！"))
	//		}
	//		verify := model.NewVerifyInfoDef(clientId, auth.AccountId, kind)
	//		if verify.GetCode() != code {
	//			return errs.NewErr(errs.MatchByMessage("认证：验证码错误！"))
	//		}
	//	case model.VerifyKindEmail:
	//
	//	case model.VerifyKindBio:
	//
	//	case model.VerifyKindThird:
	//
	//	default:
	//
	//	}
	return nil, nil
}
