package num

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FormatFloat 格式化浮点数，保留指定小数位
func FormatFloat(value float64, precision int) string {
	return strconv.FormatFloat(value, 'f', precision, 64)
}

// FormatThousands 将数字格式化为千分位分隔形式
func FormatThousands(num int64) string {
	str := strconv.FormatInt(num, 10)
	result := ""

	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}

	return result
}

// FormatMoney 货币格式化，保留两位小数并添加千分位
func FormatMoney(amount float64) string {
	intPart := int64(amount)
	fracPart := int64(math.Round((amount - float64(intPart)) * 100))

	return FormatThousands(intPart) + fmt.Sprintf(".%02d", fracPart)
}

// Round 四舍五入到指定小数位
func Round(value float64, precision int) float64 {
	factor := math.Pow10(precision)
	return math.Round(value*factor) / factor
}

// Ceil 向上取整到指定小数位
func Ceil(value float64, precision int) float64 {
	factor := math.Pow10(precision)
	return math.Ceil(value*factor) / factor
}

// Floor 向下取整到指定小数位
func Floor(value float64, precision int) float64 {
	factor := math.Pow10(precision)
	return math.Floor(value*factor) / factor
}

// ParseInt 安全地将字符串转换为整数
func ParseInt(s string, defaultVal int) int {
	val, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return defaultVal
	}
	return val
}

// ParseFloat 安全地将字符串转换为浮点数
func ParseFloat(s string, defaultVal float64) float64 {
	val, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return defaultVal
	}
	return val
}

// Max 返回多个整数中的最大值
func Max(first int, rest ...int) int {
	maxx := first
	for _, v := range rest {
		if v > maxx {
			maxx = v
		}
	}
	return maxx
}

// Min 返回多个整数中的最小值
func Min(first int, rest ...int) int {
	minn := first
	for _, v := range rest {
		if v < minn {
			minn = v
		}
	}
	return minn
}

// Sum 计算一组整数的和
func Sum(nums ...int) int {
	sum := 0
	for _, v := range nums {
		sum += v
	}
	return sum
}

// Average 计算一组整数的平均值
func Average(nums ...int) float64 {
	if len(nums) == 0 {
		return 0
	}
	return float64(Sum(nums...)) / float64(len(nums))
}
