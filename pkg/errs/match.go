package errs

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
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

	patternCache *lru.Cache[string, string] // 缓存已匹配的错误信息和对应的locId
	codeCache    *lru.Cache[string, int]    // 缓存locId和对应的code
}

// InitMatch 初始化错误匹配器
func InitMatch(cacheSize int, codes map[int][]string, patterns map[string]string, onWarn func(string)) {
	once.Do(func() {
		// 创建LRU缓存
		if cacheSize <= 0 {
			cacheSize = 1000 // 使用合理的默认值
		}
		patternCache, err := lru.New[string, string](cacheSize)
		if err != nil {
			panic(fmt.Errorf("■ ■ errs ■ ■ 初始化patternCache失败: %w", err))
		}
		codeCache, err := lru.New[string, int](cacheSize)
		if err != nil {
			panic(fmt.Errorf("■ ■ errs ■ ■ 初始化codeCache失败: %w", err))
		}

		matcher = &Matcher{
			codeLocIds:   codes,
			msgPatterns:  patterns,
			onWarn:       onWarn,
			patternCache: patternCache,
			codeCache:    codeCache,
		}

		// 初始化空映射
		if matcher.codeLocIds == nil {
			matcher.codeLocIds = make(map[int][]string)
		}
		if matcher.msgPatterns == nil {
			matcher.msgPatterns = make(map[string]string)
		}

		// 预热缓存：将locId到code的映射预先计算并缓存
		for code, locIds := range matcher.codeLocIds {
			for _, locId := range locIds {
				if locId != "" {
					matcher.codeCache.Add(locId, code)
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
	locId, found := matcher.findLocId(errMsg)

	// 再从codeLocIds里找code
	code, hasCode := matcher.findCode(locId)

	if hasCode {
		return New(err).WithCode(code).AppendLocale(locId, nil).Wash()
	} else if !found && matcher.onWarn != nil {
		matcher.onWarn(fmt.Sprintf("■ ■ Err ■ ■错误匹配err失败, msgID: %s", locId))
	}

	// 未匹配到，返回通用错误，err不返回msg
	return New(err).Wash()
}

// MatchMsg 匹配错误消息
func MatchMsg(msg string) *Error {
	if msg == "" {
		return nil
	} else if matcher == nil {
		return New(nil).WithMsg(msg).Wash()
	}

	// 先从patterns里找locId
	locId, found := matcher.findLocId(msg)

	// 再从codeLocIds里找code
	code, hasCode := matcher.findCode(locId)

	if hasCode {
		return New(nil).WithCode(code).WithMsg(msg).AppendLocale(locId, nil).Wash()
	} else if !found && matcher.onWarn != nil {
		matcher.onWarn(fmt.Sprintf("■ ■ Err ■ ■ 错误匹配msg失败, msgID: %s", locId))
	}

	// 未匹配到，返回带原始消息的错误，可以返回msg
	return New(nil).WithMsg(msg).Wash()
}

// findLocId 寻找匹配的本地化ID
func (m *Matcher) findLocId(msg string) (string, bool) {
	// 先检查缓存
	if cachedLocId, ok := m.patternCache.Get(msg); ok {
		return cachedLocId, true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 从patterns中查找匹配的本地化ID
	for pattern, msgID := range m.msgPatterns {
		if pattern != "" && strings.Contains(msg, pattern) {
			// 缓存结果
			m.patternCache.Add(msg, msgID)
			return msgID, true
		}
	}
	return msg, false
}

// findCode 根据本地化ID查找错误码
func (m *Matcher) findCode(locId string) (int, bool) {
	// 先检查缓存
	if cachedCode, ok := m.codeCache.Get(locId); ok {
		return cachedCode, true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 在codeLocIds中查找匹配的code
	for code, locIds := range m.codeLocIds {
		for _, id := range locIds {
			if id == locId {
				// 缓存结果
				m.codeCache.Add(locId, code)
				return code, true
			}
		}
	}
	return 0, false
}
