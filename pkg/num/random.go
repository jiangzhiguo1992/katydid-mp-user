package num

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// InitRandom 初始化随机数生成器
func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func Random(count int) string {
	return fmt.Sprintf("%0*d", count, rand.Intn(int(pow10(count))))
}

// RandomInRange 生成指定范围内的随机整数
func RandomInRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min)
}

// RandomFloat 生成指定范围内的随机浮点数
func RandomFloat(min, max float64, precision int) float64 {
	if min >= max {
		return min
	}
	result := min + rand.Float64()*(max-min)
	factor := math.Pow10(precision)
	return math.Round(result*factor) / factor
}

// 计算10的n次方
func pow10(n int) int64 {
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}
