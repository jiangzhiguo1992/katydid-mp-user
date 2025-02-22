package err

import (
	"errors"
	"fmt"
	"strings"
)

var (
	codeMsgIds  = map[int][]string{}
	msgPatterns = map[string]string{}
	onErr       = func(string) {}
)

func Init(codes map[int][]string, patterns map[string]string, onError func(string)) {
	codeMsgIds = codes
	msgPatterns = patterns
	onErr = onError
}

// MatchByErrMsgId 通过错误码匹配错误
func MatchByErrMsgId(msgId string) *CodeErrors {
	for code, msgIds := range codeMsgIds {
		for _, mId := range msgIds {
			if mId == msgId {
				return NewCodeError(errors.New(mId)).WithCode(code)
			}
		}
	}
	return NewCodeError(errors.New(msgId))
}

// MatchByError 通过错误信息模式匹配错误
func MatchByError(e error) *CodeErrors {
	if e == nil {
		return nil
	}
	errMsg := e.Error()
	for pattern, msgId := range msgPatterns {
		if strings.Contains(errMsg, pattern) {
			if codeErr := MatchByErrMsgId(msgId); codeErr != nil {
				return codeErr
			}
			if onErr != nil {
				onErr(fmt.Sprintf("■ ■ Err ■ ■ match pattern no code: %s", pattern))
			}
			return NewCodeError(e)
		}
	}
	return NewCodeError(e)
}

func MatchByMessage(msg string) *CodeErrors {
	return MatchByError(errors.New(msg))
}
