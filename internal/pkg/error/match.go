package error

import (
	"errors"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/log"
	"strings"
)

// MatchErrorMsgId 通过错误码匹配错误
func MatchErrorMsgId(msgId string) *err.CodeErrors {
	for code, msgIds := range codeMsgIds {
		for _, id := range msgIds {
			if id == msgId {
				return err.NewCodeErrors(errors.New(id)).WithCode(code)
			}
		}
	}
	log.Error("没有匹配的错误MsgId:", log.String("msgId", msgId))
	return nil
}

// MatchErrorPattern 通过错误信息模式匹配错误
func MatchErrorPattern(e error) *err.CodeErrors {
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
			return err.NewCodeErrors(e).WithCode(CodeUnknown)
		}
	}
	return err.NewCodeErrors(e).WithCode(CodeUnknown)
}

func MatchErrorMessage(msg string) *err.CodeErrors {
	return MatchErrorPattern(errors.New(msg))
}
