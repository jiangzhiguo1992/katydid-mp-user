package str

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

	// 优化：特殊情况处理
	if pattern == "**" || pattern == "/**" {
		return true // 匹配所有路径
	}

	// 处理前缀匹配（优先于正则匹配，提高性能）
	if len(pattern) > 1 && strings.HasSuffix(pattern, "*") && !strings.Contains(pattern[:len(pattern)-1], "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(path, prefix)
	}
	if strings.HasSuffix(pattern, "*") && !strings.Contains(pattern[:len(pattern)-1], "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(path, prefix)
	}

	// 处理后缀匹配
	if len(pattern) > 1 && strings.HasPrefix(pattern, "*") && !strings.Contains(pattern[1:], "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(path, suffix)
	}
	if strings.HasPrefix(pattern, "*") && !strings.Contains(pattern[1:], "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(path, suffix)
	}

	// 将路径和模式按斜杠分割成段
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")

	// 处理多段通配符 **
	return matchSegments(pathSegments, patternSegments, 0, 0, make(map[string]string))
}

// matchSegments 使用递归方式匹配路径段
// params 用于存储命名参数的值（如果需要使用）
func matchSegments(pathSegs []string, patternSegs []string, pathIdx, patternIdx int, params map[string]string) bool {
	// 基本终止条件
	if patternIdx >= len(patternSegs) {
		return pathIdx >= len(pathSegs) // 两者都结束才匹配
	}

	// 当前模式段
	pattern := patternSegs[patternIdx]

	// 处理 ** 多段匹配
	if pattern == "**" {
		// 这是最后一个模式段，可以匹配路径中的所有剩余段
		if patternIdx == len(patternSegs)-1 {
			return true
		}

		// 尝试匹配0个或多个路径段
		for i := pathIdx; i <= len(pathSegs); i++ {
			if matchSegments(pathSegs, patternSegs, i, patternIdx+1, params) {
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

	// 处理带正则的命名参数 {name:pattern}
	if strings.Contains(pattern, ":") && strings.HasPrefix(pattern, "{") && strings.HasSuffix(pattern, "}") {
		paramParts := strings.SplitN(pattern[1:len(pattern)-1], ":", 2)
		if len(paramParts) == 2 {
			paramName, regexStr := paramParts[0], paramParts[1]
			// 编译正则表达式
			regex, err := regexp.Compile("^" + regexStr + "$")
			if err == nil && regex.MatchString(pathSeg) {
				// 存储参数值（如果需要）
				params[paramName] = pathSeg
				return matchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params)
			}
			return false
		}
	}

	// 处理简单命名参数 {param}
	if len(pattern) > 2 && pattern[0] == '{' && pattern[len(pattern)-1] == '}' {
		// 提取参数名
		paramName := pattern[1 : len(pattern)-1]
		// 存储参数值（如果需要）
		params[paramName] = pathSeg
		return matchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params)
	}

	// 处理单个路径段
	if matchSegment(pathSeg, pattern) {
		return matchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params)
	}

	return false
}

// matchSegment 匹配单个路径段
func matchSegment(pathSeg, patternSeg string) bool {
	// 精确匹配
	if patternSeg == pathSeg {
		return true
	}

	// 单段通配符 *
	if patternSeg == "*" {
		return true
	}

	// 问号匹配单个字符
	if strings.Contains(patternSeg, "?") {
		regexPattern := strings.Replace(patternSeg, "?", ".", -1)
		regexPattern = "^" + regexPattern + "$"
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	// 字符集匹配 [abc]
	if strings.Contains(patternSeg, "[") && strings.Contains(patternSeg, "]") {
		regexPattern := "^" + patternSeg + "$"
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	// 范围匹配 {min-max}
	if rangeRegex := regexp.MustCompile(`\{(\d+)-(\d+)\}`); rangeRegex.MatchString(patternSeg) {
		matches := rangeRegex.FindStringSubmatch(patternSeg)
		if len(matches) == 3 {
			min, _ := strconv.Atoi(matches[1])
			max, _ := strconv.Atoi(matches[2])

			// 替换成正则表达式进行匹配
			replPattern := strings.Replace(patternSeg, matches[0], fmt.Sprintf("[0-9]{%d,%d}", min, max), 1)
			regex, err := regexp.Compile("^" + replPattern + "$")
			if err != nil {
				return false
			}
			return regex.MatchString(pathSeg)
		}
	}

	// 处理通配符 *
	if strings.Contains(patternSeg, "*") {
		regexPattern := "^" + strings.Replace(patternSeg, "*", ".*", -1) + "$"
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	return false
}
