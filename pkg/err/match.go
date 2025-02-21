package err

import (
	"errors"
	"log/slog"
	"strings"
)

var (
	codeMsgIds  = map[int][]string{}
	msgPatterns = map[string]string{}
)

func Init(codes map[int][]string, patterns map[string]string) {
	codeMsgIds = codes
	msgPatterns = patterns
}

// MatchByMsgId 通过错误码匹配错误
func MatchByMsgId(msgId string) *CodeErrors {
	for code, msgIds := range codeMsgIds {
		for _, id := range msgIds {
			if id == msgId {
				return NewCodeError(errors.New(id)).WithCode(code)
			}
		}
	}
	slog.Error("没有匹配的错误MsgId:", msgId)
	return nil
}

// MatchByError 通过错误信息模式匹配错误
func MatchByError(e error) *CodeErrors {
	if e == nil {
		return nil
	}
	errMsg := e.Error()
	for pattern, msgId := range msgPatterns {
		if strings.Contains(errMsg, pattern) {
			if codeErr := MatchByMsgId(msgId); codeErr != nil {
				return codeErr
			}
			return NewCodeError(e)
		}
	}
	return NewCodeError(e)
}

func MatchByMessage(msg string) *CodeErrors {
	return MatchByError(errors.New(msg))
}
