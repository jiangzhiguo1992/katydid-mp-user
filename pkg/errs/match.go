package errs

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
	codeLocIds  map[int][]string  // 错误码到本地化ID列表的映射
	msgPatterns map[string]string // 错误模式到本地化ID的映射
	onWarn      func(string)      // 警告回调函数
	mu          sync.RWMutex      // 用于保护 codeLocIds 和 msgPatterns 的并发读取

	patternCache sync.Map // 缓存已匹配的错误信息和对应的locId
	codeCache    sync.Map // 缓存locId和对应的code
}

// InitMatch 初始化错误匹配器
func InitMatch(codes map[int][]string, patterns map[string]string, onWarn func(string)) {
	once.Do(func() {
		if codes == nil {
			codes = make(map[int][]string)
		}
		if patterns == nil {
			patterns = make(map[string]string)
		}

		matcher = &Matcher{
			codeLocIds:  codes,
			msgPatterns: patterns,
			onWarn:      onWarn,
		}

		// 预热缓存：将locId到code的映射预先计算并缓存
		for code, locIds := range codes {
			for _, locId := range locIds {
				if locId != "" {
					matcher.codeCache.Store(locId, code)
				}
			}
		}
	})
}

// MatchErr 匹配Error
func MatchErr(err error) *Error {
	if err == nil {
		return nil
	}

	// 如果已经是自定义错误，直接返回
	var e *Error
	if errors.As(err, &e) {
		return e
	}

	errMsg := err.Error()
	if errMsg == "" || matcher == nil {
		return New(err).Wash()
	}

	// 先从patterns里找locId
	locId, ok1 := matcher.findLocId(errMsg)

	// 再从codeLocIds里找code
	code, ok2 := matcher.findCode(locId)

	if ok2 {
		return New(err).WithCode(code).AppendLocale(locId, nil).Wash()
	} else if !ok1 && matcher.onWarn != nil {
		matcher.onWarn(fmt.Sprintf("■ ■ Err ■ ■错误匹配err失败, msgID: %s", locId))
	}

	// 未匹配到，返回通用错误，err不返回msg
	return New(err).Wash()
}

// MatchMsg 匹配错误消息
func MatchMsg(msg string) *Error {
	if msg == "" || matcher == nil {
		return nil
	}

	// 先从patterns里找locId
	locId, ok1 := matcher.findLocId(msg)

	// 再从codeLocIds里找code
	code, ok2 := matcher.findCode(locId)

	if ok2 {
		return New(nil).WithCode(code).WithMsg(msg).AppendLocale(locId, nil).Wash()
	} else if !ok1 && matcher.onWarn != nil {
		matcher.onWarn(fmt.Sprintf("■ ■ Err ■ ■ 错误匹配msg失败, msgID: %s", locId))
	}

	// 未匹配到，返回通用错误，可以返回msg
	return New(nil).AppendLocale(msg, nil).Wash()
}

// findLocId 寻找匹配的本地化ID
func (m *Matcher) findLocId(msg string) (string, bool) {
	// 先检查缓存
	if cachedLocId, ok := m.patternCache.Load(msg); ok {
		return cachedLocId.(string), true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 从patterns中查找匹配的本地化ID
	for pattern, msgID := range m.msgPatterns {
		if pattern != "" && strings.Contains(msg, pattern) {
			// 缓存结果
			m.patternCache.Store(msg, msgID)
			return msgID, true
		}
	}
	return msg, false
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
