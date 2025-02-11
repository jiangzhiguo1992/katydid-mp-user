package err

import (
	"errors"
	"go.uber.org/zap"
	"katydid-mp-user/pkg/log"
	"strings"
)

const (
	ErrorCodeDBInsNil        = 1002
	ErrorCodeDBSelNil        = 1003
	ErrorCodeDBUpdNil        = 1004
	ErrorCodeDBDelNil        = 1005
	ErrorCodeDBFieldNil      = 1006
	ErrorCodeDBFieldLarge    = 1007 // 长
	ErrorCodeDBFieldShort    = 1008 // 短
	ErrorCodeDBFieldMax      = 1009 // 数量大
	ErrorCodeDBFieldMin      = 1010 // 数量小
	ErrorCodeDBFieldRange    = 1011
	ErrorCodeDBFieldUnDef    = 1012
	ErrorCodeDBDupPk         = 1001
	ErrorCodeDBQueryParams   = 1013
	ErrorCodeDBNoFind        = 1014
	ErrorCodeDBForeignNoFind = 1015
)

// TODO:GG 国际化
var errorMessages = map[int]string{
	ErrorCodeDBInsNil:        "数据库_插入对象为空",
	ErrorCodeDBSelNil:        "数据库_查询对象为空",
	ErrorCodeDBUpdNil:        "数据库_更新对象为空",
	ErrorCodeDBDelNil:        "数据库_删除对象为空",
	ErrorCodeDBFieldNil:      "数据库_字段为空",
	ErrorCodeDBFieldLarge:    "数据库_字段过长",
	ErrorCodeDBFieldShort:    "数据库_字段过短",
	ErrorCodeDBFieldMax:      "数据库_字段数量过多",
	ErrorCodeDBFieldMin:      "数据库_字段数量过少",
	ErrorCodeDBFieldRange:    "数据库_字段范围错误",
	ErrorCodeDBFieldUnDef:    "数据库_字段未定义",
	ErrorCodeDBDupPk:         "数据库_唯一约束冲突",
	ErrorCodeDBQueryParams:   "数据库_查询参数错误",
	ErrorCodeDBNoFind:        "数据库_未找到数据",
	ErrorCodeDBForeignNoFind: "数据库_外键未找到数据",
}

var errorCodes = map[string]int{
	"duplicate key value violates unique constraint": ErrorCodeDBDupPk,
}

func MatchErrorByCode(code int) *CodeError {
	if message, ok := errorMessages[code]; ok {
		return &CodeError{
			Code: code,
			Err:  errors.New(message),
		}
	}
	log.Warn("MatchErrorByCode 没有匹配的错误Code:", zap.Int("code", code))
	return nil
}

func MatchErrorByMsg(msg string) *CodeError {
	return MatchErrorByErr(errors.New(msg))
}

func MatchErrorByErr(err error) *CodeError {
	if err == nil {
		return nil
	}
	for msg, code := range errorCodes {
		if strings.Contains(err.Error(), msg) {
			if errorCode := MatchErrorByCode(code); errorCode != nil {
				return errorCode
			}
			log.Warn("MatchErrorByErr 没有匹配的错误Msg:", zap.Error(err))
			return &CodeError{
				Code: code,
				Err:  err,
			}
		}
	}
	return NewCodeError(err)
}
