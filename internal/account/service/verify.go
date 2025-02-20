package service

import (
	"katydid-mp-user/internal/account/model"
	error2 "katydid-mp-user/internal/pkg/error"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/log"
	"time"
)

const (
	VerifyExpire   = 5 * 60 // 验证码过期时间(5分钟) // TODO:GG 移动到configs/clients
	VerifyMaxTimes = 5      // 验证码最大次数(5次) // TODO:GG 移动到configs/clients
)

func AddVerify(clientId, accountId uint64, kind int16) (*model.VerifyInfo, *err.MultiCodeError) {
	// TODO:GG 需要检查enable
	verifyInfo := model.NewVerifyInfo(clientId, accountId, kind, time.Now().UnixMilli()+VerifyExpire*1000)
	// TODO:GG DB
	return verifyInfo, nil
}

func DelVerify(instance *model.Account) *err.MultiCodeError {
	return nil
}

func SendVerify(v *model.VerifyInfo, data map[string]interface{}) *err.CodeError {
	// TODO:GG 发送
	return nil
}

func CheckVerify(v *model.VerifyInfo, data map[string]interface{}) (bool, *err.CodeError) {
	if v.Kind == model.AuthKindPwd {
		// TODO:GG 账号密码
	} else if v.Kind == model.AuthKindPhone || v.Kind == model.AuthKindEmail {
		if v.State == model.VerityStateInit {
			return false, error2.MatchErrorMessage("验证：还没开始呢！")
		} else if v.State == model.VerityStateReject {
			if v.VerifyTimes >= VerifyMaxTimes {
				return false, error2.MatchErrorMessage("验证：超过最大次数了！")
			}
		}
		if v.IsExpired() {
			log.Debug("认证超时",
				log.Int64("pending", v.PendingAt),
				log.Int64("now", time.Now().UnixMilli()),
				log.Int64("超出s", (time.Now().UnixMilli()-v.PendingAt)/1000),
			)
			return false, error2.MatchErrorMessage("验证：已经过期了！")
		}
		needCode := v.GetCode()
		if dataCode, ok := data["code"]; !ok {
			return false, error2.MatchErrorMessage("验证：没有验证码！")
		} else if (dataCode == nil) || len(dataCode.(string)) <= 0 {
			return false, error2.MatchErrorMessage("验证：验证码为空！")
		} else if needCode != dataCode.(string) {
			log.Debug("认证失败",
				log.String("needCode", needCode),
				log.String("requestCode", dataCode.(string)),
			)
			v.SetReject()
			// TODO:GG DB
			return false, error2.MatchErrorMessage("验证：验证码错误！")
		}
		log.Debug("认证成功",
			log.String("needCode", needCode),
			log.String("requestCode", data["code"].(string)),
		)
		if v.State != model.VerityStateSuccess {
			v.SetSuccess()
			// TODO:GG DB
		} else {
			log.Warn("重复认证成功",
				log.String("needCode", needCode),
				log.String("requestCode", data["code"].(string)),
			)
		}
		return true, nil
	} else if v.Kind == model.AuthKindBio {
		// TODO:GG 生物特征
	} else if v.Kind == model.AuthKindThird {
		// TODO:GG 第三方平台
	}
	return false, error2.MatchErrorMessage("验证：不支持的类型！")
}
