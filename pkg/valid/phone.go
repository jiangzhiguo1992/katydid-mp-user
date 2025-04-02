package valid

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// PhoneCountryCode 表示国家电话代码
type PhoneCountryCode struct {
	Code     string   // 国家代码
	Name     string   // 国家名称
	Patterns []string // 电话号码模式
}

var (
	// 用于国际电话号码验证的正则表达式
	internationalPhoneRegex = regexp.MustCompile(`^\+([1-9]\d{0,3})[ -]?(\(?\d{1,4}\)?[ -]?(?:\d{1,4}[ -]?){1,4})$`)

	// 用于仅验证号码部分(纯数字)
	numberOnlyRegex = regexp.MustCompile(`^[0-9]{6,15}$`)

	// 用于从文本中提取电话号码
	phoneExtractRegex = regexp.MustCompile(`(?:\+|00)[1-9]\d{0,3}[\s.-]*(?:\(?\d{1,4}\)?[\s.-]*\d{2,4}[\s.-]*\d{2,4}[\s.-]*\d{1,4})`)

	// CountryCodes 国家代码列表
	CountryCodes = []PhoneCountryCode{
		{"1", "United States/Canada", []string{"^[2-9]\\d{2}[2-9]\\d{6}$"}},
		{"7", "Russia", []string{"^9\\d{9}$", "^[1-8]\\d{9}$"}},
		{"20", "Egypt", []string{"^01[0125]\\d{7}$"}},
		{"27", "South Africa", []string{"^[1-9]\\d{8}$"}},
		{"30", "Greece", []string{"^[2-7]\\d{9}$"}},
		{"31", "Netherlands", []string{"^[1-9]\\d{8}$"}},
		{"33", "France", []string{"^[1-9]\\d{8}$"}},
		{"34", "Spain", []string{"^[6-9]\\d{8}$"}},
		{"36", "Hungary", []string{"^[237]\\d{8}$"}},
		{"39", "Italy", []string{"^[3-9]\\d{8,9}$"}},
		{"40", "Romania", []string{"^[2-8]\\d{8}$"}},
		{"41", "Switzerland", []string{"^[2-9]\\d{8}$"}},
		{"43", "Austria", []string{"^[1-7]\\d{8}$"}},
		{"44", "United Kingdom", []string{"^7\\d{10}$"}},
		{"45", "Denmark", []string{"^[2-9]\\d{7}$"}},
		{"46", "Sweden", []string{"^[1-9]\\d{8}$"}},
		{"47", "Norway", []string{"^[2-9]\\d{7}$"}},
		{"48", "Poland", []string{"^[1-9]\\d{8}$"}},
		{"49", "Germany", []string{"^1[67]\\d{8,10}$", "^[2-9]\\d{9,10}$"}},
		{"51", "Peru", []string{"^9\\d{8}$"}},
		{"52", "Mexico", []string{"^[1-9]\\d{9}$"}},
		{"53", "Cuba", []string{"^[5-7]\\d{7}$"}},
		{"54", "Argentina", []string{"^[1-9]\\d{9}$"}},
		{"55", "Brazil", []string{"^[1-9]{2}9\\d{8}$"}},
		{"56", "Chile", []string{"^9\\d{8}$", "^(2|3)\\d{8}$"}},
		{"57", "Colombia", []string{"^[3]\\d{9}$"}},
		{"58", "Venezuela", []string{"^[4]\\d{9}$"}},
		{"60", "Malaysia", []string{"^1\\d{8,9}$"}},
		{"61", "Australia", []string{"^4\\d{8}$"}},
		{"62", "Indonesia", []string{"^8\\d{9,10}$"}},
		{"63", "Philippines", []string{"^[2-9]\\d{9}$"}},
		{"64", "New Zealand", []string{"^[2-9]\\d{7,8}$"}},
		{"65", "Singapore", []string{"^[689]\\d{7}$"}},
		{"66", "Thailand", []string{"^[689]\\d{8}$"}},
		{"81", "Japan", []string{"^[7-9]0\\d{8}$"}},
		{"82", "South Korea", []string{"^1\\d{8,9}$"}},
		{"84", "Vietnam", []string{"^[3-9]\\d{8}$"}},
		{"86", "China", []string{"^1[3-9]\\d{9}$"}},
		{"90", "Turkey", []string{"^5\\d{8,9}$"}},
		{"91", "India", []string{"^[6-9]\\d{9}$"}},
		{"92", "Pakistan", []string{"^3\\d{9}$"}},
		{"93", "Afghanistan", []string{"^7\\d{8}$"}},
		{"94", "Sri Lanka", []string{"^7\\d{8}$"}},
		{"95", "Myanmar", []string{"^[4-9]\\d{7,8}$"}},
		{"98", "Iran", []string{"^9\\d{9}$"}},
		{"212", "Morocco", []string{"^[5-9]\\d{8}$"}},
		{"213", "Algeria", []string{"^[5-7]\\d{8}$"}},
		{"216", "Tunisia", []string{"^[2-9]\\d{7}$"}},
		{"218", "Libya", []string{"^[9]\\d{8}$"}},
		{"220", "Gambia", []string{"^[5-9]\\d{6}$"}},
		{"221", "Senegal", []string{"^[7]\\d{8}$"}},
		{"222", "Mauritania", []string{"^[2-9]\\d{7}$"}},
		{"223", "Mali", []string{"^[6-9]\\d{7}$"}},
		{"224", "Guinea", []string{"^[6-9]\\d{8}$"}},
		{"225", "Ivory Coast", []string{"^\\d{10}$"}},
		{"227", "Niger", []string{"^\\d{8}$"}},
		{"229", "Benin", []string{"^[5-9]\\d{7}$"}},
		{"231", "Liberia", []string{"^[4-6]\\d{7,8}$"}},
		{"232", "Sierra Leone", []string{"^[2-9]\\d{7}$"}},
		{"233", "Ghana", []string{"^[2-9]\\d{8}$"}},
		{"234", "Nigeria", []string{"^[7-9]\\d{9}$"}},
		{"236", "Central African Republic", []string{"^[2-9]\\d{7}$"}},
		{"237", "Cameroon", []string{"^[2-9]\\d{7}$"}},
		{"238", "Cape Verde", []string{"^[5-9]\\d{6}$"}},
		{"239", "São Tomé and Príncipe", []string{"^[2-9]\\d{6}$"}},
		{"241", "Gabon", []string{"^\\d{7,8}$"}},
		{"242", "Congo", []string{"^\\d{9}$"}},
		{"243", "Democratic Republic of the Congo", []string{"^\\d{9}$"}},
		{"244", "Angola", []string{"^\\d{9}$"}},
		{"248", "Seychelles", []string{"^[2-8]\\d{5}$"}},
		{"249", "Sudan", []string{"^\\d{9}$"}},
		{"250", "Rwanda", []string{"^[7-9]\\d{8}$"}},
		{"251", "Ethiopia", []string{"^\\d{9}$"}},
		{"252", "Somalia", []string{"^[1-9]\\d{6}$"}},
		{"254", "Kenya", []string{"^\\d{9}$"}},
		{"255", "Tanzania", []string{"^\\d{9}$"}},
		{"256", "Uganda", []string{"^\\d{9}$"}},
		{"257", "Burundi", []string{"^\\d{8}$"}},
		{"258", "Mozambique", []string{"^\\d{9}$"}},
		{"260", "Zambia", []string{"^\\d{9}$"}},
		{"261", "Madagascar", []string{"^3\\d{8}$"}},
		{"262", "Réunion", []string{"^\\d{9}$"}},
		{"263", "Zimbabwe", []string{"^\\d{9}$"}},
		{"264", "Namibia", []string{"^\\d{9}$"}},
		{"265", "Malawi", []string{"^[1-9]\\d{7}$"}},
		{"266", "Lesotho", []string{"^[5-8]\\d{7}$"}},
		{"267", "Botswana", []string{"^\\d{8}$"}},
		{"268", "Eswatini", []string{"^\\d{8}$"}},
		{"269", "Comoros", []string{"^\\d{7}$"}},
		{"291", "Eritrea", []string{"^[7]\\d{6}$"}},
		{"297", "Aruba", []string{"^[5-9]\\d{6}$"}},
		{"298", "Faroe Islands", []string{"^[2-9]\\d{5}$"}},
		{"299", "Greenland", []string{"^[2-9]\\d{5}$"}},
		{"350", "Gibraltar", []string{"^[5-9]\\d{7}$"}},
		{"351", "Portugal", []string{"^9\\d{8}$"}},
		{"352", "Luxembourg", []string{"^[5-9]\\d{8}$"}},
		{"353", "Ireland", []string{"^[1-9]\\d{8}$"}},
		{"354", "Iceland", []string{"^[4-9]\\d{6}$"}},
		{"355", "Albania", []string{"^[4-9]\\d{8}$"}},
		{"356", "Malta", []string{"^[79]\\d{7}$"}},
		{"357", "Cyprus", []string{"^[9]\\d{7}$"}},
		{"358", "Finland", []string{"^[4-9]\\d{8}$"}},
		{"359", "Bulgaria", []string{"^[7-9]\\d{8}$"}},
		{"370", "Lithuania", []string{"^[6-9]\\d{7}$"}},
		{"371", "Latvia", []string{"^[2-9]\\d{7}$"}},
		{"372", "Estonia", []string{"^[5-9]\\d{6}$"}},
		{"373", "Moldova", []string{"^[6-9]\\d{7}$"}},
		{"374", "Armenia", []string{"^[4-9]\\d{7}$"}},
		{"375", "Belarus", []string{"^[1-9]\\d{8}$"}},
		{"376", "Andorra", []string{"^[3-9]\\d{5}$"}},
		{"377", "Monaco", []string{"^[4-9]\\d{7}$"}},
		{"378", "San Marino", []string{"^[5-7]\\d{7}$"}},
		{"379", "Vatican City", []string{"^\\d{8}$"}},
		{"380", "Ukraine", []string{"^[3-9]\\d{8}$"}},
		{"381", "Serbia", []string{"^[6-9]\\d{8}$"}},
		{"382", "Montenegro", []string{"^[6-9]\\d{7}$"}},
		{"383", "Kosovo", []string{"^[4-5]\\d{7}$"}},
		{"385", "Croatia", []string{"^[1-9]\\d{8}$"}},
		{"386", "Slovenia", []string{"^[3-9]\\d{7}$"}},
		{"387", "Bosnia and Herzegovina", []string{"^[6]\\d{7}$"}},
		{"389", "North Macedonia", []string{"^[7]\\d{7}$"}},
		{"420", "Czech Republic", []string{"^[2-9]\\d{8}$"}},
		{"421", "Slovakia", []string{"^[9]\\d{8}$"}},
		{"423", "Liechtenstein", []string{"^[6-7]\\d{6}$"}},
		{"500", "Falkland Islands", []string{"^[5-6]\\d{4}$"}},
		{"501", "Belize", []string{"^[6]\\d{6}$"}},
		{"502", "Guatemala", []string{"^[3-5]\\d{7}$"}},
		{"503", "El Salvador", []string{"^[67]\\d{7}$"}},
		{"504", "Honduras", []string{"^[3-9]\\d{7}$"}},
		{"505", "Nicaragua", []string{"^[5-8]\\d{7}$"}},
		{"506", "Costa Rica", []string{"^[6-8]\\d{7}$"}},
		{"507", "Panama", []string{"^[6]\\d{7}$"}},
		{"509", "Haiti", []string{"^[3-4]\\d{7}$"}},
		{"590", "Guadeloupe", []string{"^[6-7]\\d{8}$"}},
		{"591", "Bolivia", []string{"^[6-7]\\d{7}$"}},
		{"592", "Guyana", []string{"^[6]\\d{6}$"}},
		{"593", "Ecuador", []string{"^[9]\\d{8}$"}},
		{"595", "Paraguay", []string{"^[9]\\d{8}$"}},
		{"598", "Uruguay", []string{"^[9]\\d{7}$"}},
		{"599", "Curaçao", []string{"^[9]\\d{6}$"}},
		{"670", "Timor-Leste", []string{"^[7]\\d{6}$"}},
		{"673", "Brunei", []string{"^[7-8]\\d{6}$"}},
		{"674", "Nauru", []string{"^[4-5]\\d{5}$"}},
		{"675", "Papua New Guinea", []string{"^[7]\\d{6}$"}},
		{"676", "Tonga", []string{"^[7]\\d{6}$"}},
		{"677", "Solomon Islands", []string{"^[7-8]\\d{6}$"}},
		{"678", "Vanuatu", []string{"^[5-7]\\d{6}$"}},
		{"679", "Fiji", []string{"^[6-9]\\d{6}$"}},
		{"680", "Palau", []string{"^[2-8]\\d{6}$"}},
		{"682", "Cook Islands", []string{"^[5]\\d{4}$"}},
		{"685", "Samoa", []string{"^[7]\\d{5}$"}},
		{"686", "Kiribati", []string{"^[9]\\d{5}$"}},
		{"691", "Micronesia", []string{"^[3-9]\\d{6}$"}},
		{"692", "Marshall Islands", []string{"^[2-6]\\d{6}$"}},
		{"850", "North Korea", []string{"^[1-9]\\d{7}$"}},
		{"852", "Hong Kong", []string{"^[569]\\d{7}$", "^8\\d{7}$"}},
		{"853", "Macau", []string{"^[6]\\d{7}$"}},
		{"855", "Cambodia", []string{"^[1-9]\\d{7}$"}},
		{"856", "Laos", []string{"^[2-8]\\d{7}$"}},
		{"880", "Bangladesh", []string{"^1\\d{9}$"}},
		{"886", "Taiwan", []string{"^9\\d{8}$", "^8\\d{8}$"}},
		{"960", "Maldives", []string{"^[7-9]\\d{6}$"}},
		{"961", "Lebanon", []string{"^[3-9]\\d{7}$"}},
		{"962", "Jordan", []string{"^7\\d{8}$"}},
		{"963", "Syria", []string{"^9\\d{8}$"}},
		{"964", "Iraq", []string{"^7\\d{9}$"}},
		{"965", "Kuwait", []string{"^[5-9]\\d{7}$"}},
		{"966", "Saudi Arabia", []string{"^5\\d{8}$", "^9\\d{8}$"}},
		{"967", "Yemen", []string{"^7\\d{8}$"}},
		{"968", "Oman", []string{"^9\\d{7}$"}},
		{"970", "Palestine", []string{"^5\\d{8}$"}},
		{"971", "United Arab Emirates", []string{"^5\\d{8}$", "^[467]\\d{8}$"}},
		{"972", "Israel", []string{"^5\\d{8}$", "^0?5\\d{8}$"}},
		{"973", "Bahrain", []string{"^[36]\\d{7}$"}},
		{"974", "Qatar", []string{"^[3-7]\\d{7}$"}},
		{"975", "Bhutan", []string{"^[1-7]\\d{7}$"}},
		{"976", "Mongolia", []string{"^[5-9]\\d{7}$"}},
		{"977", "Nepal", []string{"^9\\d{9}$"}},
		{"992", "Tajikistan", []string{"^[1-9]\\d{8}$"}},
		{"993", "Turkmenistan", []string{"^[6]\\d{7}$"}},
		{"994", "Azerbaijan", []string{"^[4-9]\\d{8}$"}},
		{"995", "Georgia", []string{"^[5-9]\\d{8}$"}},
		{"996", "Kyrgyzstan", []string{"^[3-9]\\d{8}$"}},
		{"998", "Uzbekistan", []string{"^[1-9]\\d{8}$"}},
		{"1242", "Bahamas", []string{"^[2-9]\\d{6}$"}},
		{"1246", "Barbados", []string{"^[2-9]\\d{6}$"}},
		{"1264", "Anguilla", []string{"^[2-9]\\d{6}$"}},
		{"1268", "Antigua and Barbuda", []string{"^[2-9]\\d{6}$"}},
		{"1284", "British Virgin Islands", []string{"^[2-9]\\d{6}$"}},
		{"1340", "U.S. Virgin Islands", []string{"^[2-9]\\d{6}$"}},
		{"1345", "Cayman Islands", []string{"^[2-9]\\d{6}$"}},
		{"1441", "Bermuda", []string{"^[2-9]\\d{6}$"}},
		{"1473", "Grenada", []string{"^[2-9]\\d{6}$"}},
		{"1649", "Turks and Caicos Islands", []string{"^[2-9]\\d{6}$"}},
		{"1664", "Montserrat", []string{"^[2-9]\\d{6}$"}},
		{"1670", "Northern Mariana Islands", []string{"^[2-9]\\d{6}$"}},
		{"1671", "Guam", []string{"^[2-9]\\d{6}$"}},
		{"1684", "American Samoa", []string{"^[2-9]\\d{6}$"}},
		{"1721", "Sint Maarten", []string{"^[5-9]\\d{6}$"}},
		{"1758", "Saint Lucia", []string{"^[2-9]\\d{6}$"}},
		{"1767", "Dominica", []string{"^[2-9]\\d{6}$"}},
		{"1784", "Saint Vincent and the Grenadines", []string{"^[2-9]\\d{6}$"}},
		{"1787", "Puerto Rico", []string{"^[2-9]\\d{6}$"}},
		{"1809", "Dominican Republic", []string{"^[2-9]\\d{6}$"}},
		{"1829", "Dominican Republic", []string{"^[2-9]\\d{6}$"}},
		{"1849", "Dominican Republic", []string{"^[2-9]\\d{6}$"}},
		{"1869", "Saint Kitts and Nevis", []string{"^[2-9]\\d{6}$"}},
		{"1876", "Jamaica", []string{"^[2-9]\\d{6}$"}},
		{"1939", "Puerto Rico", []string{"^[2-9]\\d{6}$"}},
	}

	// 缓存用于快速查找
	countryCodeMap     = make(map[string]PhoneCountryCode)
	sortedCountryCodes []string
	compiledPatterns   = make(map[string][]*regexp.Regexp)
)

func init() {
	// 构建查找映射并预编译所有正则表达式
	for _, countryCode := range CountryCodes {
		countryCodeMap[countryCode.Code] = countryCode
		sortedCountryCodes = append(sortedCountryCodes, countryCode.Code)

		// 预编译模式以提升性能
		var patterns []*regexp.Regexp
		for _, pattern := range countryCode.Patterns {
			patterns = append(patterns, regexp.MustCompile(pattern))
		}
		compiledPatterns[countryCode.Code] = patterns
	}

	// 按长度降序排序国家代码，确保先匹配最长的代码
	sort.Slice(sortedCountryCodes, func(i, j int) bool {
		return len(sortedCountryCodes[i]) > len(sortedCountryCodes[j])
	})
}

// IsPhone 将phone分割成CountryCode和number，并验证数据
func IsPhone(phone string) (*PhoneCountryCode, string, bool) {
	if phone = strings.TrimSpace(phone); phone == "" {
		return nil, "", false
	}

	original := phone
	phone = cleanPhoneNumber(phone)
	if phone == "" {
		return nil, "", false
	}

	// 检查是否是国际格式电话号码
	if strings.HasPrefix(phone, "+") {
		// 使用国际电话正则表达式
		matches := internationalPhoneRegex.FindStringSubmatch(original) // 使用原始输入以保留格式
		if len(matches) == 3 {
			code := matches[1]
			number := cleanDigitsOnly(matches[2])

			// 验证国家代码是否存在
			if countryCode, exists := countryCodeMap[code]; exists {
				// 直接使用已知的国家代码进行验证
				if isValid := validateNumberWithPattern(code, number); isValid {
					return &countryCode, number, true
				}
				return &countryCode, number, false
			}
		}

		// 如果正则不匹配，尝试逐一检查前缀
		cleanPhone := strings.TrimPrefix(phone, "+")
		return findMatchingCountryCode(cleanPhone)
	}

	// 如果不是国际格式，检查是否是有效的纯数字格式
	if numberOnlyRegex.MatchString(phone) {
		return nil, phone, true
	}

	return nil, phone, false
}

// IsPhoneNumber 验证特定国家代码的电话号码
func IsPhoneNumber(code, number string) (*PhoneCountryCode, string, bool) {
	// 清理号码，移除非数字字符
	number = cleanDigitsOnly(number)
	if number == "" {
		return nil, "", false
	}

	// 检查国家代码是否有效
	countryCode, ok := countryCodeMap[code]
	if !ok {
		return nil, number, false
	}

	// 使用验证函数检查号码格式
	valid := validateNumberWithPattern(code, number)
	return &countryCode, number, valid
}

// validateNumberWithPattern 验证电话号码是否匹配国家代码的模式
func validateNumberWithPattern(code, number string) bool {
	patterns, exists := compiledPatterns[code]
	if !exists || len(patterns) == 0 {
		// 如果没有特定模式，至少验证长度是否合理
		return len(number) >= 6 && len(number) <= 15
	}

	for _, re := range patterns {
		if re != nil && re.MatchString(number) {
			return true
		}
	}
	return false
}

// IsPhoneCountryCode 检查提供的字符串是否是有效的国家代码
func IsPhoneCountryCode(code string) (*PhoneCountryCode, bool) {
	if code = strings.TrimSpace(code); code == "" {
		return nil, false
	}

	// 移除前导的'+'(如果存在)
	code = cleanPhoneNumber(code)
	if code == "" {
		return nil, false
	}

	// 移除前导的'+'(如果存在)
	if strings.HasPrefix(code, "+") {
		code = strings.TrimPrefix(code, "+")
	}

	// 检查国家代码是否有效
	if countryCode, exists := countryCodeMap[code]; exists {
		return &countryCode, true
	}
	return nil, false
}

// FormatPhone 根据国家代码格式格式化电话号码
func FormatPhone(phone string) string {
	if phone = strings.TrimSpace(phone); phone == "" {
		return ""
	}

	code, number, ok := IsPhone(phone)
	if !ok {
		// 如果无效，返回统一处理后的格式而不是原始输入
		cleanedPhone := cleanPhoneNumber(phone)
		if cleanedPhone == "" {
			return phone // 如果清理后为空，则返回原始输入
		}
		return cleanedPhone
	}

	// 如果没有国家代码，只返回号码
	if code == nil {
		return number
	}

	// 简单格式化，在国家代码和号码之间添加空格
	return "+" + code.Code + " " + number
}

// FindPhonesInText 从文本中查找可能的电话号码
func FindPhonesInText(text string) []string {
	if text = strings.TrimSpace(text); text == "" {
		return nil
	}

	// 使用更严格的正则表达式模式匹配国际号码格式
	matches := phoneExtractRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	// 过滤只保留有效号码并去重
	validNumbersMap := make(map[string]bool)
	var validNumbers []string

	for _, match := range matches {
		_, _, ok := IsPhone(match)
		if ok {
			formatted := FormatPhone(match)
			if formatted != "" && !validNumbersMap[formatted] {
				validNumbersMap[formatted] = true
				validNumbers = append(validNumbers, formatted)
			}
		}
	}

	return validNumbers
}

// FindPhoneCountryByName 通过名称搜索国家代码
func FindPhoneCountryByName(name string) []PhoneCountryCode {
	if name = strings.TrimSpace(name); name == "" {
		return nil
	}

	name = strings.ToLower(name)
	var results []PhoneCountryCode

	for _, code := range CountryCodes {
		if strings.Contains(strings.ToLower(code.Name), name) {
			results = append(results, code)
		}
	}

	return results
}

// findMatchingCountryCode 查找匹配的国家代码，通过按长度排序后的代码列表提高效率
func findMatchingCountryCode(phone string) (*PhoneCountryCode, string, bool) {
	if phone == "" {
		return nil, "", false
	}

	for _, code := range sortedCountryCodes {
		if strings.HasPrefix(phone, code) {
			number := phone[len(code):]
			if number == "" {
				continue // 确保号码部分不为空
			}

			countryCode, exists := countryCodeMap[code]
			if !exists {
				continue
			}

			// 直接使用验证函数，减少冗余调用
			valid := validateNumberWithPattern(code, number)
			return &countryCode, number, valid
		}
	}
	return nil, phone, false
}

// cleanPhoneNumber 清理电话号码，移除所有空白字符，仅保留+和数字
func cleanPhoneNumber(phone string) string {
	if phone == "" {
		return ""
	}

	phone = strings.TrimSpace(phone)
	var builder strings.Builder
	builder.Grow(len(phone))

	for _, r := range phone {
		if r == '+' || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// cleanDigitsOnly 只保留数字字符
func cleanDigitsOnly(input string) string {
	if input == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(input))

	for _, r := range input {
		if unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}
