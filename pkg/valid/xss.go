package valid

import (
	"github.com/microcosm-cc/bluemonday"
	"regexp"
)

var (
	// XSS攻击模式检测正则表达式
	xssPatterns = []*regexp.Regexp{
		// 脚本标签及其变种
		regexp.MustCompile(`(?i)<\s*script[\s\S]*?>`),
		regexp.MustCompile(`(?i)</\s*script\s*>`),

		// JavaScript 协议变种
		regexp.MustCompile(`(?i)javascript\s*:`),
		regexp.MustCompile(`(?i)vbscript\s*:`),
		regexp.MustCompile(`(?i)livescript\s*:`),
		regexp.MustCompile(`(?i)data\s*:.*script`),

		// JavaScript 事件处理器
		regexp.MustCompile(`(?i)\s+on\w+\s*=`),

		// 危险函数
		regexp.MustCompile(`(?i)\b(eval|setTimeout|setInterval|Function|execScript)\s*\(`),
		regexp.MustCompile(`(?i)(document|window|location|cookie|localStorage)\.(cookie|domain|write|location)`),

		// 内联框架及对象标签
		regexp.MustCompile(`(?i)<\s*(iframe|embed|object|base|applet)\b`),

		// 数据URI
		regexp.MustCompile(`(?i)data:(?:text|image|application)/(?:html|xml|xhtml|svg)`),
		regexp.MustCompile(`(?i)data:.*?;base64`),

		// 表达式和绕过
		regexp.MustCompile(`(?i)expression\s*\(`),
		regexp.MustCompile(`(?i)@import\s+`),
		regexp.MustCompile(`(?i)url\s*\(`),

		// HTML5 特性
		regexp.MustCompile(`(?i)formaction\s*=`),
		regexp.MustCompile(`(?i)srcdoc\s*=`),

		// 元素属性
		regexp.MustCompile(`(?i)\bhref\s*=\s*["']?(?:javascript:|data:text|vbscript:)`),
		regexp.MustCompile(`(?i)\bsrc\s*=\s*["']?(?:javascript:|data:text|vbscript:)`),

		// 常见的HTML注入向量
		regexp.MustCompile(`(?i)<\s*style[^>]*>.*?(expression|behavior|javascript|vbscript).*?</style>`),
		regexp.MustCompile(`(?i)<\s*link[^>]*(?:href|xlink:href)\s*=\s*["']?(?:javascript:|data:text|vbscript:)`),

		// SVG嵌入式脚本
		regexp.MustCompile(`(?i)<\s*svg[^>]*>.*?<\s*script`),
	}
)

// XSSValidator 提供检测XSS攻击的功能
type XSSValidator struct {
	patterns  []*regexp.Regexp
	sanitizer *bluemonday.Policy
}

// NewXSSValidator 返回一个配置好的XSS验证器
func NewXSSValidator(strict bool) *XSSValidator {
	validator := &XSSValidator{
		patterns: make([]*regexp.Regexp, len(xssPatterns)),
	}
	// 编译所有正则表达式模式
	for i, pattern := range xssPatterns {
		validator.patterns[i] = pattern
	}

	if strict {
		validator.sanitizer = bluemonday.StrictPolicy() // 使用严格策略，删除所有HTML
	} else {
		validator.sanitizer = bluemonday.UGCPolicy() // 使用用户生成内容策略，允许一些HTML
	}
	return validator
}

// SanitizeXSS 使用bluemonday清理文本中的XSS内容
func (v *XSSValidator) SanitizeXSS(text string) string {
	if text == "" {
		return text
	}
	return v.sanitizer.Sanitize(text)
}

// HasXSS 检查文本中是否包含可能的XSS攻击内容
func (v *XSSValidator) HasXSS(text string) bool {
	if text == "" {
		return false
	}
	for _, pattern := range v.patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	// 也可以通过比较sanitized后的结果是否与原文本相同来判断
	return v.SanitizeXSS(text) != text
}

// GetXSSMatches 返回文本中匹配到的所有XSS模式
func (v *XSSValidator) GetXSSMatches(text string) []string {
	if text == "" {
		return nil
	}
	var matches []string
	for _, pattern := range v.patterns {
		if found := pattern.FindAllString(text, -1); len(found) > 0 {
			matches = append(matches, found...)
		}
	}
	return matches
}

// ValidateInput 验证用户输入是否安全
func (v *XSSValidator) ValidateInput(input string) (bool, []string) {
	if input == "" {
		return true, nil
	} else if v.HasXSS(input) {
		return false, v.GetXSSMatches(input)
	}
	return true, nil
}
