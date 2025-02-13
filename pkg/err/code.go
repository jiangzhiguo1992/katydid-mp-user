package err

import (
	"errors"
	"go.uber.org/zap"
	"katydid-mp-user/pkg/log"
	"strings"
)

const (
	// TODO:GG 这些还有必要吗？一般的code都是提醒前端做某些操作？

	CodeDBAddNil         = 1002
	CodeDBDelNil         = 1003
	CodeDBUpdNil         = 1004
	CodeDBQueNil         = 1005
	CodeDBFieldNil       = 1006
	CodeDBFieldLarge     = 1007 // 长
	CodeDBFieldShort     = 1008 // 短
	CodeDBFieldMax       = 1009 // 数量大
	CodeDBFieldMin       = 1010 // 数量小
	CodeDBFieldRange     = 1011
	CodeDBFieldUnDefined = 1012
	CodeDBPkDuplicated   = 1001
	CodeDBQueParams      = 1013
	CodeDBQueNone        = 1014
	CodeDBQueForeignNone = 1015
)

// TODO:GG 是不是可以在newerr的时候直接把locales塞进去，不用匹配?
var codeLocales = map[int]string{
	CodeDBAddNil:         "err_db_add_nil",
	CodeDBDelNil:         "err_db_del_nil",
	CodeDBUpdNil:         "err_db_upd_nil",
	CodeDBQueNil:         "err_db_que_nil",
	CodeDBFieldNil:       "err_db_field_nil",
	CodeDBFieldLarge:     "err_db_field_large",
	CodeDBFieldShort:     "err_db_field_short",
	CodeDBFieldMax:       "err_db_field_max",
	CodeDBFieldMin:       "err_db_field_min",
	CodeDBFieldRange:     "err_db_field_range",
	CodeDBFieldUnDefined: "err_db_field_undefined",
	CodeDBPkDuplicated:   "err_db_pk_duplicated",
	CodeDBQueParams:      "err_db_que_params",
	CodeDBQueNone:        "err_db_que_none",
	CodeDBQueForeignNone: "err_db_que_foreign_none",
}

var errorCodes = map[string]int{
	"duplicate key value violates unique constraint": CodeDBPkDuplicated,
}

func MatchErrorByCode(code int) *CodeError {
	if message, ok := codeLocales[code]; ok {
		return &CodeError{
			Code: code,
			Err:  errors.New(message),
		}
	}
	log.Warn("没有匹配的错误Code:", zap.Int("code", code))
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
			log.Warn("没有匹配的错误Msg:", zap.Error(err))
			return &CodeError{
				Code: code,
				Err:  err,
			}
		}
	}
	return NewCodeError(err)
}
