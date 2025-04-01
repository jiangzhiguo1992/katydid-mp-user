package str

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	rangeRegex        = regexp.MustCompile(`\{(\d+)-(\d+)\}`)
	paramRegex        = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	brackets          = regexp.MustCompile(`\[(.*?)\]`)
	questionMarkRegex = regexp.MustCompile(`\?`)
	maxRecursionDepth = 50 // 减小递归深度限制
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

	// 标准化输入，移除多余的斜杠
	path = strings.TrimSpace(path)
	pattern = strings.TrimSpace(pattern)

	// 优化：特殊情况处理
	if pattern == "**" || pattern == "/**" {
		return true // 匹配所有路径
	}

	// 处理前缀匹配（优先于正则匹配，提高性能）
	if strings.HasSuffix(pattern, "*") && pattern != "*" && len(pattern) > 1 {
		prefix := pattern[:len(pattern)-1]
		if !strings.Contains(prefix, "*") && !strings.Contains(prefix, "?") &&
			!strings.Contains(prefix, "{") && !strings.Contains(prefix, "[") {
			return strings.HasPrefix(path, prefix)
		}
	}

	// 处理后缀匹配
	if strings.HasPrefix(pattern, "*") && pattern != "*" && len(pattern) > 1 {
		suffix := pattern[1:]
		if !strings.Contains(suffix, "*") && !strings.Contains(suffix, "?") &&
			!strings.Contains(suffix, "{") && !strings.Contains(suffix, "[") {
			return strings.HasSuffix(path, suffix)
		}
	}

	// 将路径和模式按斜杠分割成段
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")

	// 处理多段通配符 **
	return matchSegments(pathSegments, patternSegments, 0, 0, make(map[string]string), 0)
}

// matchSegments 使用递归方式匹配路径段
// params 用于存储命名参数的值（如果需要使用）
func matchSegments(pathSegs []string, patternSegs []string, pathIdx, patternIdx int, params map[string]string, depth int) bool {
	// 防止过深的递归导致栈溢出
	if depth > 100 {
		return false
	}

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
			if matchSegments(pathSegs, patternSegs, i, patternIdx+1, params, depth+1) {
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
	if strings.HasPrefix(pattern, "{") && strings.HasSuffix(pattern, "}") {
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
				// 添加更严格的参数名检查
				if paramName != "" && paramRegex.MatchString(paramName) {
					// 尝试编译正则表达式，添加错误恢复
					regex, err := regexp.Compile("^" + regexStr + "$")
					if err == nil && regex.MatchString(pathSeg) {
						// 存储参数值
						params[paramName] = pathSeg
						return matchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params, depth+1)
					}
				}
			}
			return false
		} else {
			// 简单命名参数 {param}
			paramName := paramContent
			if paramName != "" && paramRegex.MatchString(paramName) {
				params[paramName] = pathSeg
				return matchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params, depth+1)
			}
		}
		return false // 明确返回false，如果参数名为空或无效
	}

	// 处理单个路径段
	if matchSegment(pathSeg, pattern) {
		return matchSegments(pathSegs, patternSegs, pathIdx+1, patternIdx+1, params, depth+1)
	}

	return false
}

// matchSegment 匹配单个路径段
func matchSegment(pathSeg, patternSeg string) bool {
	// 精确匹配
	if patternSeg == pathSeg || patternSeg == "*" {
		return true
	}

	// 问号匹配单个字符
	if strings.Contains(patternSeg, "?") {
		// 将模式中除了问号外的其他正则元字符进行转义
		pattern := ""
		for _, ch := range patternSeg {
			if ch == '?' {
				pattern += "."
			} else {
				pattern += regexp.QuoteMeta(string(ch))
			}
		}
		// 添加错误恢复
		regex, err := regexp.Compile("^" + pattern + "$")
		if err != nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	// 字符集匹配 [abc]
	if strings.Contains(patternSeg, "[") && strings.Contains(patternSeg, "]") {
		// 先验证方括号格式是否正确
		if !isValidBracketPattern(patternSeg) {
			return false
		}

		// 使用更安全的方式提取括号内容
		matches := brackets.FindAllStringSubmatchIndex(patternSeg, -1)
		if matches == nil || len(matches) == 0 {
			return false // 无法正确匹配括号
		}

		// 构建正则表达式，更安全地处理每一部分
		regexPattern := "^"
		lastIndex := 0

		for _, match := range matches {
			// 确保索引有效
			if match[0] < 0 || match[1] > len(patternSeg) ||
				match[2] < 0 || match[3] > len(patternSeg) ||
				match[2] >= match[3] {
				return false
			}

			// 转义方括号前的内容
			if match[0] > lastIndex {
				regexPattern += regexp.QuoteMeta(patternSeg[lastIndex:match[0]])
			}

			// 添加方括号内容，转义特殊字符
			bracketContent := patternSeg[match[2]:match[3]]
			// 检查是否为特殊正则表达式
			if strings.ContainsAny(bracketContent, "\\^-]") {
				// 对特殊字符进行转义处理
				var escaped strings.Builder
				for _, ch := range bracketContent {
					if ch == '\\' || ch == '^' || ch == '-' || ch == ']' {
						escaped.WriteRune('\\')
					}
					escaped.WriteRune(ch)
				}
				bracketContent = escaped.String()
			}
			regexPattern += "[" + bracketContent + "]"
			lastIndex = match[1]
		}

		// 添加方括号后的内容
		if lastIndex < len(patternSeg) {
			regexPattern += regexp.QuoteMeta(patternSeg[lastIndex:])
		}
		regexPattern += "$"

		// 添加错误处理
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	// 范围匹配 {min-max}
	if rangeRegex.MatchString(patternSeg) {
		matches := rangeRegex.FindStringSubmatch(patternSeg)
		if len(matches) != 3 {
			return false
		}

		minn, minErr := strconv.Atoi(matches[1])
		maxx, maxErr := strconv.Atoi(matches[2])

		// 增强数值校验，防止范围过大
		if minErr != nil || maxErr != nil ||
			minn < 0 || maxx < 0 || minn > maxx || maxx > 100 {
			return false
		}

		// 构建正则模式
		parts := rangeRegex.Split(patternSeg, -1)
		regexPattern := "^"
		for i, part := range parts {
			regexPattern += regexp.QuoteMeta(part)
			if i < len(parts)-1 {
				regexPattern += fmt.Sprintf("\\d{%d,%d}", minn, maxx)
			}
		}
		regexPattern += "$"

		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	// 处理通配符 *
	if strings.Contains(patternSeg, "*") {
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

		// 添加错误处理
		regex, err := regexp.Compile(pattern.String())
		if err != nil {
			return false
		}
		return regex.MatchString(pathSeg)
	}

	return false
}

// 添加辅助函数检查字符集合模式是否有效
func isValidBracketPattern(pattern string) bool {
	// 检查是否有未闭合的方括号
	openCount := strings.Count(pattern, "[")
	closeCount := strings.Count(pattern, "]")
	if openCount != closeCount || openCount == 0 {
		return false
	}

	// 检查方括号内是否有内容及嵌套方括号
	matches := brackets.FindAllStringSubmatch(pattern, -1)
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
