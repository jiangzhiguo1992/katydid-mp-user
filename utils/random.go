package utils

import (
	"fmt"
	"math/rand"
)

func Random(count int) string {
	return fmt.Sprintf("%0*d", count, rand.Intn(int(pow10(count))))
}

func pow10(n int) int64 {
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}
