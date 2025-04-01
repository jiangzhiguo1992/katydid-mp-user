package valid

import (
	"regexp"
	"strings"
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
func IsEmail(email string) (*EmailComponents, bool) {
	// 先检查总长度限制
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
	username = strings.TrimSpace(username)
	if username == "" {
		return false
	} else if len(username) > 64 {
		// RFC 5321 本地部分最大长度
		return false
	}

	// 验证用户名中的字符是否符合RFC规范
	basicRegex := regexp.MustCompile(`^[a-zA-Z0-9!#$%&'*+\-/=?^_` + "`" + `{|}~.]+$`)
	if !basicRegex.MatchString(username) {
		return false
	}

	// 确保点号不在开头、结尾或连续
	if strings.HasPrefix(username, ".") || strings.HasSuffix(username, ".") || strings.Contains(username, "..") {
		return false
	}

	return true
}

// IsEmailDomain 验证电子邮件的域名部分
func IsEmailDomain(domain string) bool {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return false
	} else if len(domain) > 255 {
		return false
	}

	// 基本域名验证
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)
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
	// 使用更精确的正则表达式直接匹配可能有效的邮箱
	emailRegex := regexp.MustCompile(`\b[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}\b`)
	matches := emailRegex.FindAllString(text, -1)

	validEmails := make([]string, 0, len(matches))
	for _, match := range matches {
		// 使用IsEmail进行最终验证
		if _, ok := IsEmail(match); ok {
			validEmails = append(validEmails, match)
		}
	}

	return validEmails
}

// parseEmail 将电子邮件分割成不同组件
func parseEmail(email string) (*EmailComponents, bool) {
	// 移除空格
	email = strings.TrimSpace(email)

	// 转为小写（尽管技术上用户名可以区分大小写）
	email = strings.ToLower(email)

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

	// 提取TLD和实体部分
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
