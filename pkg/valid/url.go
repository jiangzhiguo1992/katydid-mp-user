package valid

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	// 预编译正则表达式，提高性能
	urlRangeRegex = regexp.MustCompile(`\{(\d+)-(\d+)\}`)
	urlParamRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	urlBrackets   = regexp.MustCompile(`\[(.*?)\]`)

	// 使用sync.Pool缓存编译后的正则表达式，减少重复编译
	urlRegexPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]*regexp.Regexp)
		},
	}
	urlRegexPoolMutex sync.RWMutex
)

const (
	maxRecursionDepth  = 50   // 递归深度限制
	maxSegmentLength   = 2000 // 路径段长度限制
	maxRegexLength     = 500  // 正则表达式长度限制
	maxQuestionMarks   = 100  // 问号数量限制
	maxBrackets        = 10   // 方括号数量限制
	maxBracketContent  = 100  // 字符集合内容长度限制
	maxWildcards       = 10   // 通配符数量限制
	maxRangeValue      = 50   // 范围上限值
	maxRangeDifference = 20   // 范围差值限制
)

// MatchURLPath 检查传入的路径是否匹配指定的模式
// 支持以下通配符规则:
//   - * : 匹配单个路径段中的任意字符
//   - ** : 匹配零个或多个路径段
//   - ? : 匹配单个字符
//   - {name} : 命名参数，匹配任意单个路径段
//   - {name:pattern} : 带正则模式的命名参数，如 {id:[0-9]+} 只匹配数字
//   - [...] : 字符集合，如 [abc] 匹配 a、b 或 c
//   - /prefix* : 前缀匹配，如 /api* 匹配所有以 /api 开头的路径
//   - *.suffix : 后缀匹配，如 *.js 匹配所有以 .js 结尾的路径
//
// 例如:
//   - /user/* 匹配 /user/login, /user/profile, 但不匹配 /user 或 /user/login/details
//   - /api/** 匹配 /api, /api/v1, /api/v1/users 等任意层级
//   - /user/?id 匹配 /user/1id, /user/aid 等
//   - /user/{id} 匹配 /user/123, /user/abc 等任意单段
//   - /user/{id:[0-9]+} 只匹配 /user/123, 不匹配 /user/abc
//   - /files/[abc] 匹配 /files/a, /files/b, /files/c
//   - /api* 匹配 /api, /api/v1, /apidocs 等任意以 /api 开头的路径
//   - *.js 匹配 /script.js, /js/app.js 等任意以 .js 结尾的路径
func MatchURLPath(path string, pattern string) bool {
	// 处理空路径或模式
	if path == "" || pattern == "" {
		return path == pattern
	}

	// 标准化输入，移除多余的空白
	path = strings.TrimSpace(path)
	pattern = strings.TrimSpace(pattern)

	// 检查路径和模式的长度
	if len(path) > maxSegmentLength*10 || len(pattern) > maxSegmentLength*10 {
		return false
	}

	// 优化：特殊情况处理
	if pattern == "**" || pattern == "/**" {
		return true // 匹配所有路径
	}

	// 处理前缀匹配（优先于正则匹配，提高性能）
	if strings.HasSuffix(pattern, "*") && pattern != "*" && len(pattern) > 1 {
		prefix := pattern[:len(pattern)-1]
		if !strings.ContainsAny(prefix, "*?{[") {
			return strings.HasPrefix(path, prefix)
		}
	}

	// 处理后缀匹配
	if strings.HasPrefix(pattern, "*") && pattern != "*" && len(pattern) > 1 {
		suffix := pattern[1:]
		if !strings.ContainsAny(suffix, "*?{[") {
			return strings.HasSuffix(path, suffix)
		}
	}

	// 将路径和模式按斜杠分割成段
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")

	// 限制路径段数量
	if len(pathSegments) > 100 || len(patternSegments) > 100 {
		return false
	}

	// 创建一个新的参数映射表，避免并发问题
	params := make(map[string]string)

	// 处理多段通配符 **
	return urlMatchSegments(pathSegments, patternSegments, 0, 0, params, 0)
}

// urlMatchSegments 使用递归方式匹配路径段
func urlMatchSegments(pathSegs []string, patternSegs []string, pathIdx, patternIdx int, params map[string]string, depth int) bool {
	// 防止过深的递归导致栈溢出
	if depth > maxRecursionDepth {
		return false
	}

	// 基本终止条件优化
	if pathIdx >= len(pathSegs) && patternIdx >= len(patternSegs) {
		return true // 两者都结束则匹配成功
	}

	if patternIdx >= len(patternSegs) {
		return false // 模式段结束但路径段未结束，不匹配
	}

	// 当前模式段
	pattern := patternSegs[patternIdx]

	// 处理 ** 多段匹配
	if pattern == "**" {
		// 这是最后一个模式段，可以匹配路径中的所有剩余段
		if patternIdx == len(patternSegs)-1 {
			return true
		}

		// 检查剩余模式段数是否超过剩余路径段数
		remainingPatternSegs := len(patternSegs) - patternIdx - 1
		if remainingPatternSegs > len(pathSegs)-pathIdx {
			return false // 不可能匹配成功
		}

		// 尝试匹配0个或多个路径段（优化：从多到少，更容易找到成功匹配）
		for i := len(pathSegs); i >= pathIdx; i-- {
			// 创建临时参数表，避免污染原参数表
			tempParams := make(map[string]string, len(params))
			for k, v := range params {
				tempParams[k] = v
			}

			if urlMatchSegments(pathSegs, patternSegs, i, patternIdx+1, tempParams, depth+1) {
				// 成功匹配时，将临时参数合并回原参数表
				for k, v := range tempParams {
					params[k] = v
				}
				return true
			}
		}
		return false
	}

	// 如果路径段已结束但模式段未结束（且不是 **），则不匹配
	if pathIdx >= len(pathSegs) {
		return false
	}

	pathSeg := pathSegs[pathIdx]

	// 添加防护：检查路径段长度
	if len(pathSeg) > maxSegmentLength || len(pattern) > maxSegmentLength {
		return false
	}

	// 处理带正则的命名参数 {name:pattern}
	if strings.HasPrefix(pattern, "{") && strings.HasSuffix(pattern, "}") && len(pattern) >= 3 {
		// 提取命名参数内容
		paramContent := pattern[1 : len(pattern)-1]

		// 检查参数内容是否为空
		if len(paramContent) == 0 {
			return false
		}

		// 检查是否包含正则表达式部分
		if strings.Contains(paramContent, ":") {
			paramParts := strings.SplitN(paramContent, ":", 2)
			if len(paramParts) == 2 {
				paramName, regexStr := paramParts[0], paramParts[1]

				// 限制正则表达式长度
				if len(regexStr) > maxRegexLength {
					return false
				}

				// 添加更严格的参数名检查
				if paramName != "" && urlParamRegex.MatchString(paramName) {
					// 使用正则表达式缓存
					regex := getCompiledUrlRegex("^" + regexStr + "$")
					if regex != nil && regex.MatchString(pathSeg) {
						// 存储参数值
						params[paramName] = pathSeg
						return urlMatchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params, depth+1)
					}
				}
			}
			return false
		} else {
			// 简单命名参数 {param}
			paramName := paramContent
			if paramName != "" && urlParamRegex.MatchString(paramName) {
				params[paramName] = pathSeg
				return urlMatchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params, depth+1)
			}
		}
		return false // 明确返回false，如果参数名为空或无效
	}

	// 处理单个路径段
	if urlMatchSegment(pathSeg, pattern) {
		return urlMatchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params, depth+1)
	}

	return false
}

// urlMatchSegment 匹配单个路径段
func urlMatchSegment(pathSeg, patternSeg string) bool {
	// 防止过长的路径段
	if len(pathSeg) > maxSegmentLength || len(patternSeg) > maxSegmentLength {
		return false
	}

	// 精确匹配或通配符匹配
	if patternSeg == pathSeg || patternSeg == "*" {
		return true
	}

	// 问号匹配单个字符
	if strings.Contains(patternSeg, "?") {
		// 限制问号数量，防止DoS攻击
		if strings.Count(patternSeg, "?") > maxQuestionMarks {
			return false
		}

		// 构建安全的正则模式
		pattern := strings.Builder{}
		pattern.WriteString("^")
		for _, ch := range patternSeg {
			if ch == '?' {
				pattern.WriteString(".")
			} else {
				pattern.WriteString(regexp.QuoteMeta(string(ch)))
			}
		}
		pattern.WriteString("$")

		// 使用缓存的正则表达式
		regex := getCompiledUrlRegex(pattern.String())
		if regex == nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	// 字符集匹配 [abc]
	if strings.Contains(patternSeg, "[") && strings.Contains(patternSeg, "]") {
		// 限制字符集数量
		if strings.Count(patternSeg, "[") > maxBrackets {
			return false
		}

		// 先验证方括号格式是否正确
		if !isValidUrlBracketPattern(patternSeg) {
			return false
		}

		// 使用更安全的方式提取括号内容
		matches := urlBrackets.FindAllStringSubmatchIndex(patternSeg, -1)
		if matches == nil || len(matches) == 0 {
			return false // 无法正确匹配括号
		}

		// 构建正则表达式，更安全地处理每一部分
		regexPattern := strings.Builder{}
		regexPattern.WriteString("^")
		lastIndex := 0

		for _, match := range matches {
			// 确保索引有效
			if len(match) < 4 {
				return false
			}

			if match[0] < 0 || match[1] > len(patternSeg) ||
				match[2] < 0 || match[3] > len(patternSeg) ||
				match[0] >= match[1] || match[2] >= match[3] {
				return false
			}

			// 转义方括号前的内容
			if match[0] > lastIndex {
				regexPattern.WriteString(regexp.QuoteMeta(patternSeg[lastIndex:match[0]]))
			}

			// 添加方括号内容，检查长度限制
			bracketContent := patternSeg[match[2]:match[3]]
			if len(bracketContent) > maxBracketContent {
				return false
			}
			regexPattern.WriteString("[" + bracketContent + "]")
			lastIndex = match[1]
		}

		// 添加方括号后的内容
		if lastIndex < len(patternSeg) {
			regexPattern.WriteString(regexp.QuoteMeta(patternSeg[lastIndex:]))
		}
		regexPattern.WriteString("$")

		// 使用正则表达式缓存
		regex := getCompiledUrlRegex(regexPattern.String())
		if regex == nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	// 范围匹配 {min-max}，增强安全性
	if urlRangeRegex.MatchString(patternSeg) {
		matches := urlRangeRegex.FindStringSubmatch(patternSeg)
		if len(matches) != 3 {
			return false
		}

		minn, minErr := strconv.Atoi(matches[1])
		maxx, maxErr := strconv.Atoi(matches[2])

		// 增强数值校验，防止范围过大
		if minErr != nil || maxErr != nil ||
			minn < 0 || maxx < 0 || minn > maxx ||
			maxx > maxRangeValue || maxx-minn > maxRangeDifference {
			return false
		}

		// 构建正则模式
		parts := urlRangeRegex.Split(patternSeg, -1)
		regexPattern := strings.Builder{}
		regexPattern.WriteString("^")
		for i, part := range parts {
			regexPattern.WriteString(regexp.QuoteMeta(part))
			if i < len(parts)-1 {
				regexPattern.WriteString(fmt.Sprintf("\\d{%d,%d}", minn, maxx))
			}
		}
		regexPattern.WriteString("$")

		// 使用正则表达式缓存
		regex := getCompiledUrlRegex(regexPattern.String())
		if regex == nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	// 处理通配符 *，增强安全性
	if strings.Contains(patternSeg, "*") {
		// 限制星号数量
		if strings.Count(patternSeg, "*") > maxWildcards {
			return false
		}

		var pattern strings.Builder
		pattern.WriteString("^")
		for i := 0; i < len(patternSeg); i++ {
			if patternSeg[i] == '*' {
				pattern.WriteString(".*")
			} else {
				pattern.WriteString(regexp.QuoteMeta(string(patternSeg[i])))
			}
		}
		pattern.WriteString("$")

		// 使用正则表达式缓存
		regex := getCompiledUrlRegex(pattern.String())
		if regex == nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	return false
}

// isValidUrlBracketPattern 检查字符集合模式是否有效
func isValidUrlBracketPattern(pattern string) bool {
	// 检查是否有未闭合的方括号
	openCount := strings.Count(pattern, "[")
	closeCount := strings.Count(pattern, "]")
	if openCount != closeCount || openCount == 0 {
		return false
	}

	// 检查方括号内是否有内容及嵌套方括号
	matches := urlBrackets.FindAllStringSubmatch(pattern, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// 检查是否为空字符集合
			if len(match[1]) == 0 {
				return false
			}
			// 检查是否有嵌套方括号
			if strings.Contains(match[1], "[") || strings.Contains(match[1], "]") {
				return false
			}
		}
	}

	// 确保方括号正确配对
	stack := 0
	for _, char := range pattern {
		if char == '[' {
			stack++
		} else if char == ']' {
			stack--
			if stack < 0 {
				return false // 右括号在左括号之前出现
			}
		}
	}

	return stack == 0
}

// getCompiledUrlRegex 从缓存中获取编译后的正则表达式，如果不存在则编译并缓存
// 通过正则表达式缓存提高性能，避免重复编译相同的正则表达式
func getCompiledUrlRegex(pattern string) *regexp.Regexp {
	// 先尝试从缓存中读取
	urlRegexPoolMutex.RLock()
	cache := urlRegexPool.Get().(map[string]*regexp.Regexp)
	regex, found := cache[pattern]
	urlRegexPoolMutex.RUnlock()

	if found {
		// 确保归还缓存
		urlRegexPool.Put(cache)
		return regex
	}

	// 未找到则编译
	compiledRegex, err := regexp.Compile(pattern)
	if err != nil {
		urlRegexPool.Put(cache) // 确保归还缓存
		return nil
	}

	// 写入缓存
	urlRegexPoolMutex.Lock()
	cache[pattern] = compiledRegex
	urlRegexPoolMutex.Unlock()

	// 归还缓存
	urlRegexPool.Put(cache)

	return compiledRegex
}
