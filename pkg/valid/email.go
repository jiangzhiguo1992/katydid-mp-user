package valid

import (
	"regexp"
	"strings"
)

var (
	// 用户名部分正则表达式 - 符合RFC 5322标准，验证邮箱@前的本地部分
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9!#$%&'*+\-/=?^_\` + "`" + `{|}~.]+$`)
	// 域名部分正则表达式，验证邮箱@后的域名部分
	domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)
	// 优化后的电子邮件匹配正则表达式 - 用于从文本中初步筛选可能的邮箱地址
	emailRegex = regexp.MustCompile(`\b[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}(?:\b|$)`)
)

// EmailComponents 保存解析后的电子邮件地址各部分
type EmailComponents struct {
	Address  string // 完整电子邮件地址
	Username string // @前的本地部分
	Domain   string // @后的域名部分
	Entity   string // 实体部分（子域名）
	TLD      string // 顶级域名（域名最后部分）
}

// IsEmail 验证电子邮件地址的各个组成部分
// 返回解析后的邮件组件和一个表示邮件是否有效的布尔值
// 如果邮件无效，但能够解析出组件，则返回组件和false
// 如果邮件格式完全无效无法解析，则返回nil和false
func IsEmail(email string) (*EmailComponents, bool) {
	// 检查是否为空
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, false
	}

	// 检查总长度限制 (RFC 5321)
	if len(email) > 254 {
		return nil, false
	}

	// 首先解析电子邮件
	components, ok := parseEmail(email)
	if !ok {
		return nil, false
	}

	// 用户名验证
	if !IsEmailUsername(components.Username) {
		return components, false
	}

	// 域名验证
	if !IsEmailDomain(components.Domain) {
		return components, false
	}

	return components, true
}

// IsEmailUsername 验证电子邮件的本地部分
func IsEmailUsername(username string) bool {
	// 检查是否为空（不需要再次调用TrimSpace，parseEmail已经处理过）
	if username == "" {
		return false
	}

	// RFC 5321 本地部分最大长度为64个字符
	if len(username) > 64 {
		return false
	}

	// 确保点号不在开头、结尾或连续
	if strings.HasPrefix(username, ".") || strings.HasSuffix(username, ".") || strings.Contains(username, "..") {
		return false
	}

	// 验证用户名中的字符是否符合RFC规范
	return usernameRegex.MatchString(username)
}

// IsEmailDomain 验证电子邮件的域名部分
func IsEmailDomain(domain string) bool {
	// 检查是否为空（不需要再次调用TrimSpace，parseEmail已经处理过）
	if domain == "" {
		return false
	}

	// 检查域名长度限制
	if len(domain) > 255 {
		return false
	}

	// 基本域名验证
	if !domainRegex.MatchString(domain) {
		return false
	}

	// 检查TLD长度，至少2个字符
	parts := strings.Split(domain, ".")
	if len(parts) < 2 || len(parts[len(parts)-1]) < 2 {
		return false
	}

	// 可选：检查域名是否存在（DNS记录）
	// 注意：这可能会导致较慢的验证过程，建议在需要严格验证时使用
	/*if _, err := net.LookupHost(domain); err != nil {
		// 尝试查询MX记录作为备选
		if _, err := net.LookupMX(domain); err != nil {
			return false
		}
	}*/

	return true
}

// FindEmailsInText 从文本中提取有效的电子邮件地址
func FindEmailsInText(text string) []string {
	// 检查文本是否为空
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	// 使用正则表达式初步筛选可能的邮箱地址
	matches := emailRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return []string{}
	}

	// 预分配空间以避免多次扩容
	validEmails := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches)) // 用于去重

	for _, match := range matches {
		// 移除可能的前后空格
		match = strings.TrimSpace(match)

		// 检查是否已处理过相同邮箱，优化放前面，减少IsEmail调用
		if _, exists := seen[match]; exists {
			continue
		}

		// 使用IsEmail进行最终验证
		if _, ok := IsEmail(match); ok {
			validEmails = append(validEmails, match)
			seen[match] = struct{}{}
		}
	}

	return validEmails
}

// parseEmail 将电子邮件分割成不同组件
func parseEmail(email string) (*EmailComponents, bool) {
	// 移除空格并转为小写
	email = strings.ToLower(strings.TrimSpace(email))

	// 检查邮件是否为空
	if email == "" {
		return nil, false
	}

	// 按@符号分割
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return nil, false
	}

	username := parts[0]
	domain := parts[1]

	// 验证基本格式
	if username == "" || domain == "" {
		return nil, false
	}

	// 提取 TLD 和实体部分
	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return nil, false
	}

	tld := domainParts[len(domainParts)-1]
	entity := strings.Join(domainParts[:len(domainParts)-1], ".")

	return &EmailComponents{
		Address:  email,
		Username: username,
		Domain:   domain,
		Entity:   entity,
		TLD:      tld,
	}, true
}
