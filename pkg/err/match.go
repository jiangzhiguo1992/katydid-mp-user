package err

import (
	"fmt"
	"strings"
	"sync"
)

var (
	matcher *Matcher
	once    sync.Once
)

// Matcher 错误匹配器
type Matcher struct {
	codeLocIds map[int][]string
	patterns   map[string]string
	onError    func(string)
}

// Init 初始化错误匹配器
func Init(codes map[int][]string, patterns map[string]string, onError func(string)) {
	once.Do(func() {
		matcher = &Matcher{
			codeLocIds: codes,
			patterns:   patterns,
			onError:    onError,
		}
	})
}

// Match 匹配错误
func Match(err error) *CodeErrs {
	if err == nil {
		return nil
	}
	errMsg := err.Error()

	// 先从patterns里找locId
	var matchLocId string
	for pattern, msgID := range matcher.patterns {
		if strings.Contains(errMsg, pattern) {
			matchLocId = msgID
			break
		}
	}

	// 未匹配到，返回通用错误
	common := false
	if len(matchLocId) <= 0 {
		common = true
		matchLocId = errMsg
	}

	// 再从codeLocIds里找code
	if len(matchLocId) > 0 {
		for code, locIds := range matcher.codeLocIds {
			for _, locId := range locIds {
				if locId == matchLocId {
					return New(err).WithCode(code).WithLocalize(locId, nil)
				}
			}
		}
		if !common && (matcher.onError != nil) {
			matcher.onError(fmt.Sprintf("■ ■ Err ■ ■ match pattern no code: %s", matchLocId))
		}
	}

	// 未匹配到，返回通用错误
	return New(err)
}
