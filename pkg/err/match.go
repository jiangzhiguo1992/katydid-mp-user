package err

import (
	"errors"
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
	mu         sync.RWMutex

	patternCache sync.Map // 缓存已匹配的错误信息和对应的locId
	codeCache    sync.Map // 缓存locId和对应的code
	//commonErrorCache sync.Map // 缓存常见错误消息的结果
}

// Init 初始化错误匹配器
func Init(codes map[int][]string, patterns map[string]string, onError func(string)) {
	once.Do(func() {
		matcher = &Matcher{
			codeLocIds: codes,
			patterns:   patterns,
			onError:    onError,
		}

		// 预热缓存：将locId到code的映射预先计算并缓存
		for code, locIds := range codes {
			for _, locId := range locIds {
				matcher.codeCache.Store(locId, code)
			}
		}
	})
}

// Match 匹配错误
func Match(err error) *CodeErrs {
	if err == nil {
		return nil
	} else if matcher == nil {
		return New(err).Real()
	} else if len(err.Error()) <= 0 {
		return New(err).Real()
	}

	// 如果已经是自定义错误，直接返回
	var e *CodeErrs
	if errors.As(err, &e) {
		return e
	}

	// 先从patterns里找locId
	locId, ok1 := matcher.findLocId(err.Error())

	// 再从codeLocIds里找code
	code, ok2 := matcher.findCode(locId)

	if ok2 {
		return New(err).WithCode(code).WrapLocalize(locId, nil, nil).Real()
	} else if !ok1 && matcher.onError != nil {
		matcher.onError(fmt.Sprintf("■ ■ Err ■ ■ match pattern no code: %s", locId))
	}
	// 未匹配到，返回通用错误
	return New(err).Real()
}

func Match2(msg string) *CodeErrs {
	if msg == "" {
		return nil
	}

	// 先从patterns里找locId
	locId, ok1 := matcher.findLocId(msg)

	// 再从codeLocIds里找code
	code, ok2 := matcher.findCode(locId)

	if ok2 {
		return New().WithCode(code).WrapLocalize(locId, nil, nil).Real()
	} else if !ok1 && matcher.onError != nil {
		matcher.onError(fmt.Sprintf("■ ■ Err ■ ■ matchMsg pattern no code: %s", locId))
	}
	// 未匹配到，返回通用错误
	return New().WrapLocalize(msg, nil, nil).Real()
}

// findLocId 寻找匹配的本地化ID
func (m *Matcher) findLocId(errMsg string) (string, bool) {
	// 先检查缓存
	if cachedLocId, ok := m.patternCache.Load(errMsg); ok {
		return cachedLocId.(string), true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 从patterns中查找匹配的本地化ID
	for pattern, msgID := range m.patterns {
		if strings.Contains(errMsg, pattern) {
			// 缓存结果
			m.patternCache.Store(errMsg, msgID)
			return msgID, true
		}
	}

	return errMsg, false
}

// findCode 根据本地化ID查找错误码
func (m *Matcher) findCode(locId string) (int, bool) {
	// 先检查缓存
	if cachedCode, ok := m.codeCache.Load(locId); ok {
		return cachedCode.(int), true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 在codeLocIds中查找匹配的code
	for code, locIds := range m.codeLocIds {
		for _, id := range locIds {
			if id == locId {
				// 缓存结果
				m.codeCache.Store(locId, code)
				return code, true
			}
		}
	}

	return 0, false
}

//// createCodeError 创建错误对象
//func (m *Matcher) createCodeError(origErr error, locId string, isCommon bool, hasCode bool, code int) *CodeErrs {
//	// 对于常见错误模式，使用缓存减少对象创建
//	if isCommon {
//		cacheKey := fmt.Sprintf("%s:%d", locId, code)
//		if cached, ok := m.commonErrorCache.Load(cacheKey); ok {
//			// 返回缓存的副本以避免并发修改
//			cachedErr := cached.(*CodeErrs)
//			return New(origErr).WithCode(cachedErr.code).WrapLocalize(cachedErr.localizes[0].localeId, nil, nil)
//		}
//	}
//
//	codeErr := New(origErr).WrapLocalize(locId, nil, nil)
//	if hasCode {
//		_ = codeErr.WithCode(code)
//	}
//
//	// 对于常见错误，缓存模式
//	if isCommon {
//		cacheKey := fmt.Sprintf("%s:%d", locId, code)
//		m.commonErrorCache.Store(cacheKey, codeErr)
//	}
//
//	return codeErr.Real()
//}
