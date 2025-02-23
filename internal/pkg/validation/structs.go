package validation

import (
	accountModel "katydid-mp-user/internal/account/model"
	clientModel "katydid-mp-user/internal/client/model"
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/valid"
	"regexp"
)

const (
	patUsername = `^[a-zA-Z0-9_]{6,16}$`             // 只允许字母、数字和下划线
	patPassword = "^[a-zA-Z0-9_]{6,16}$"             // 只允许字母、数字和下划线
	patPhone    = `^1[3-9]\d{9}$`                    // 11位手机号
	patEmail    = `^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$` // 邮箱地址
)

var Structs = []interface{}{
	&model.Base{},
	// account
	&accountModel.Account{},
	// client
	&clientModel.Client{},
}

// TODO:GG 国际化

var NilTips = map[string]string{
	"Phone": "没有手机号",
}

var FiledValidators = map[string]valid.ValidatorFieldFunc{
	"username": func(value interface{}) (string, bool) {
		if str, ok := value.(string); ok {
			match, _ := regexp.MatchString(patUsername, str)
			return "用户名格式不正确", match
		}
		return "用户名格式不正确", false
	},
	"password": func(value interface{}) (string, bool) {
		if str, ok := value.(string); ok {
			match, _ := regexp.MatchString(patPassword, str)
			return "密码格式不正确", match
		}
		return "密码格式不正确", false
	},
	"phone": func(value interface{}) (string, bool) {
		if str, ok := value.(string); ok {
			match, _ := regexp.MatchString(patPhone, str)
			return "手机号格式不正确", match
		}
		return "手机号格式不正确", false
	},
	"email": func(value interface{}) (string, bool) {
		if str, ok := value.(string); ok {
			match, _ := regexp.MatchString(patEmail, str)
			return "邮箱格式不正确", match
		}
		return "邮箱格式不正确", false
	},
}

var GroupValidators = map[string]valid.ValidatorGroupFunc{
	"pwd-confirm": func(values map[string]interface{}) (string, bool) {
		pwd := values["Password"].(string)
		rePwd := values["RePassword"].(string)
		return "两次密码不一致", pwd == rePwd
	},
}
