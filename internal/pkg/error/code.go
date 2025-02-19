package error

import (
	"errors"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/log"
	"strings"
)

const (
	CodeUnknown = 0
	CodeDB      = 1000

	MsgIdDBPkDuplicated   = "err_db_pk_duplicated"
	MsgIdDBAddNil         = "err_db_add_nil"
	MsgIdDBDelNil         = "err_db_del_nil"
	MsgIdDBUpdNil         = "err_db_upd_nil"
	MsgIdDBQueNil         = "err_db_que_nil"
	MsgIdDBFieldNil       = "err_db_field_nil"
	MsgIdDBFieldLarge     = "err_db_field_large"
	MsgIdDBFieldShort     = "err_db_field_short"
	MsgIdDBFieldMax       = "err_db_field_max"
	MsgIdDBFieldMin       = "err_db_field_min"
	MsgIdDBFieldRange     = "err_db_field_range"
	MsgIdDBFieldUnDefined = "err_db_field_undefined"
	MsgIdDBQueParams      = "err_db_que_params"
	MsgIdDBQueNone        = "err_db_que_none"
	MsgIdDBQueForeignNone = "err_db_que_foreign_none"
)

var (
	// 错误信息映射
	codeMsgIds = map[int][]string{
		CodeDB: {
			MsgIdDBPkDuplicated,
			MsgIdDBAddNil,
			MsgIdDBDelNil,
			MsgIdDBUpdNil,
			MsgIdDBQueNil,
			MsgIdDBFieldNil,
			MsgIdDBFieldLarge,
			MsgIdDBFieldShort,
			MsgIdDBFieldMax,
			MsgIdDBFieldMin,
			MsgIdDBFieldRange,
			MsgIdDBFieldUnDefined,
			MsgIdDBQueParams,
			MsgIdDBQueNone,
			MsgIdDBQueForeignNone,
		},
	}

	// 错误模式匹配
	errorPatterns = map[string]string{
		"duplicate key value violates unique constraint": MsgIdDBPkDuplicated,
	}
)

// MatchErrorMsgId 通过错误码匹配错误
func MatchErrorMsgId(msgId string) *err.CodeError {
	for code, msgIds := range codeMsgIds {
		for _, id := range msgIds {
			if id == msgId {
				return err.NewCodeError(errors.New(id)).WithCode(code)
			}
		}
	}
	log.Error("没有匹配的错误MsgId:", log.String("msgId", msgId))
	return nil
}

// MatchErrorPattern 通过错误信息模式匹配错误
func MatchErrorPattern(e error) *err.CodeError {
	if e == nil {
		return nil
	}
	errMsg := e.Error()
	for pattern, msgId := range errorPatterns {
		if strings.Contains(errMsg, pattern) {
			if codeErr := MatchErrorMsgId(msgId); codeErr != nil {
				return codeErr
			}
			log.Error("没有匹配的错误Msg:", log.Err(e))
			return err.NewCodeError(e).WithCode(CodeUnknown)
		}
	}
	return err.NewCodeError(e).WithCode(CodeUnknown)
}

func MatchErrorMessage(msg string) *err.CodeError {
	return MatchErrorPattern(errors.New(msg))
}
