package utils

import (
	"math"
)

// KSMap 扩展 map[string]any 类型
type KSMap map[string]any

// Set 设置任意类型值
func (m KSMap) Set(key string, value any) {
	m[key] = value
}

// SetPtr 设置指针型值
func (m KSMap) SetPtr(key string, value *any) {
	if value == nil {
		m.Delete(key)
		return
	}
	m.Set(key, *value)
}

// Delete 删除指定key
func (m KSMap) Delete(key string) {
	delete(m, key)
}

// Get 获取任意类型值，需要自己做类型断言
func (m KSMap) Get(key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}

// GetSlice 获取[]any类型值
func (m KSMap) GetSlice(key string) ([]any, bool) {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]any); ok {
			return slice, true
		}
	}
	return nil, false
}

// GetInt 获取int类型值
func (m KSMap) GetInt(key string) (int, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val, true
		case int8:
			return int(val), true
		case int16:
			return int(val), true
		case int32:
			return int(val), true
		case int64:
			if val >= int64(math.MinInt32) && val <= int64(math.MaxInt32) {
				return int(val), true
			}
		case uint:
			if val <= uint(math.MaxInt32) {
				return int(val), true
			}
		case uint8:
			return int(val), true
		case uint16:
			return int(val), true
		case uint32:
			if val <= uint32(math.MaxInt32) {
				return int(val), true
			}
		case uint64:
			if val <= uint64(math.MaxInt32) {
				return int(val), true
			}
		case float32:
			if val >= float32(math.MinInt32) && val <= float32(math.MaxInt32) {
				return int(val), true
			}
		case float64:
			if val >= float64(math.MinInt32) && val <= float64(math.MaxInt32) {
				return int(val), true
			}
		}
	}
	return 0, false
}

// GetIntSlice 获取[]int类型值
func (m KSMap) GetIntSlice(key string) ([]int, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []int:
			return val, true
		case []any:
			ints := make([]int, 0, len(val))
			for _, item := range val {
				if i, ok := item.(int); ok {
					ints = append(ints, i)
				} else {
					return nil, false
				}
			}
			return ints, true
		}
	}
	return nil, false
}

// GetInt8 获取int8类型值
func (m KSMap) GetInt8(key string) (int8, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int8:
			return val, true
		case int:
			if val >= math.MinInt8 && val <= math.MaxInt8 {
				return int8(val), true
			}
		case int16:
			if val >= math.MinInt8 && val <= math.MaxInt8 {
				return int8(val), true
			}
		case int32:
			if val >= math.MinInt8 && val <= math.MaxInt8 {
				return int8(val), true
			}
		case int64:
			if val >= math.MinInt8 && val <= math.MaxInt8 {
				return int8(val), true
			}
		case uint:
			if val <= uint(math.MaxInt8) {
				return int8(val), true
			}
		case uint8:
			if val <= uint8(math.MaxInt8) {
				return int8(val), true
			}
		case uint16:
			if val <= uint16(math.MaxInt8) {
				return int8(val), true
			}
		case uint32:
			if val <= uint32(math.MaxInt8) {
				return int8(val), true
			}
		case uint64:
			if val <= uint64(math.MaxInt8) {
				return int8(val), true
			}
		case float32:
			if val >= math.MinInt8 && val <= math.MaxInt8 {
				return int8(val), true
			}
		case float64:
			if val >= math.MinInt8 && val <= math.MaxInt8 {
				return int8(val), true
			}
		}
	}
	return 0, false
}

// GetInt8Slice 获取[]int8类型值
func (m KSMap) GetInt8Slice(key string) ([]int8, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []int8:
			return val, true
		case []any:
			nums := make([]int8, 0, len(val))
			for _, item := range val {
				if i, ok := item.(int8); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetInt16 获取int16类型值
func (m KSMap) GetInt16(key string) (int16, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int16:
			return val, true
		case int8:
			return int16(val), true
		case int:
			if val >= math.MinInt16 && val <= math.MaxInt16 {
				return int16(val), true
			}
		case int32:
			if val >= math.MinInt16 && val <= math.MaxInt16 {
				return int16(val), true
			}
		case int64:
			if val >= math.MinInt16 && val <= math.MaxInt16 {
				return int16(val), true
			}
		case uint:
			if val <= uint(math.MaxInt16) {
				return int16(val), true
			}
		case uint8:
			return int16(val), true
		case uint16:
			if val <= uint16(math.MaxInt16) {
				return int16(val), true
			}
		case uint32:
			if val <= uint32(math.MaxInt16) {
				return int16(val), true
			}
		case uint64:
			if val <= uint64(math.MaxInt16) {
				return int16(val), true
			}
		case float32:
			if val >= math.MinInt16 && val <= math.MaxInt16 {
				return int16(val), true
			}
		case float64:
			if val >= math.MinInt16 && val <= math.MaxInt16 {
				return int16(val), true
			}
		}
	}
	return 0, false
}

// GetInt16Slice 获取[]int16类型值
func (m KSMap) GetInt16Slice(key string) ([]int16, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []int16:
			return val, true
		case []any:
			nums := make([]int16, 0, len(val))
			for _, item := range val {
				if i, ok := item.(int16); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetInt64 获取int64类型值
func (m KSMap) GetInt64(key string) (int64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val, true
		case int:
			return int64(val), true
		case int8:
			return int64(val), true
		case int16:
			return int64(val), true
		case int32:
			return int64(val), true
		case uint:
			return int64(val), true
		case uint8:
			return int64(val), true
		case uint16:
			return int64(val), true
		case uint32:
			return int64(val), true
		case uint64:
			if val <= uint64(math.MaxInt64) {
				return int64(val), true
			}
		case float32:
			if float32(math.MinInt64) <= val && val <= float32(math.MaxInt64) {
				return int64(val), true
			}
		case float64:
			if math.MinInt64 <= val && val <= math.MaxInt64 {
				return int64(val), true
			}
		}
	}
	return 0, false
}

// GetInt64Slice 获取[]int64类型值
func (m KSMap) GetInt64Slice(key string) ([]int64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []int64:
			return val, true
		case []any:
			nums := make([]int64, 0, len(val))
			for _, item := range val {
				if i, ok := item.(int64); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetUint 获取uint类型值
func (m KSMap) GetUint(key string) (uint, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case uint:
			return val, true
		case uint8:
			return uint(val), true
		case uint16:
			return uint(val), true
		case uint32:
			if val <= uint32(math.MaxUint32) {
				return uint(val), true
			}
		case uint64:
			if val <= uint64(math.MaxUint32) {
				return uint(val), true
			}
		case int:
			if val >= 0 && val <= math.MaxUint32 {
				return uint(val), true
			}
		case int8:
			if val >= 0 {
				return uint(val), true
			}
		case int16:
			if val >= 0 {
				return uint(val), true
			}
		case int32:
			if val >= 0 {
				return uint(val), true
			}
		case int64:
			if val >= 0 && val <= math.MaxUint32 {
				return uint(val), true
			}
		case float32:
			if val >= 0 && val <= float32(math.MaxUint32) {
				return uint(val), true
			}
		case float64:
			if val >= 0 && val <= float64(math.MaxUint32) {
				return uint(val), true
			}
		}
	}
	return 0, false
}

// GetUintSlice 获取[]uint类型值
func (m KSMap) GetUintSlice(key string) ([]uint, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []uint:
			return val, true
		case []any:
			nums := make([]uint, 0, len(val))
			for _, item := range val {
				if i, ok := item.(uint); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetUint8 获取uint8类型值
func (m KSMap) GetUint8(key string) (uint8, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case uint8:
			return val, true
		case uint:
			if val <= math.MaxUint8 {
				return uint8(val), true
			}
		case uint16:
			if val <= math.MaxUint8 {
				return uint8(val), true
			}
		case uint32:
			if val <= math.MaxUint8 {
				return uint8(val), true
			}
		case uint64:
			if val <= math.MaxUint8 {
				return uint8(val), true
			}
		case int:
			if val >= 0 && val <= math.MaxUint8 {
				return uint8(val), true
			}
		case int8:
			if val >= 0 {
				return uint8(val), true
			}
		case int16:
			if val >= 0 && val <= math.MaxUint8 {
				return uint8(val), true
			}
		case int32:
			if val >= 0 && val <= math.MaxUint8 {
				return uint8(val), true
			}
		case int64:
			if val >= 0 && val <= math.MaxUint8 {
				return uint8(val), true
			}
		case float32:
			if val >= 0 && val <= float32(math.MaxUint8) {
				return uint8(val), true
			}
		case float64:
			if val >= 0 && val <= float64(math.MaxUint8) {
				return uint8(val), true
			}
		}
	}
	return 0, false
}

// GetUint8Slice 获取[]uint8类型值
func (m KSMap) GetUint8Slice(key string) ([]uint8, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []uint8:
			return val, true
		case []any:
			nums := make([]uint8, 0, len(val))
			for _, item := range val {
				if i, ok := item.(uint8); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetUint16 获取uint16类型值
func (m KSMap) GetUint16(key string) (uint16, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case uint16:
			return val, true
		case uint8:
			return uint16(val), true
		case uint:
			if val <= math.MaxUint16 {
				return uint16(val), true
			}
		case uint32:
			if val <= math.MaxUint16 {
				return uint16(val), true
			}
		case uint64:
			if val <= math.MaxUint16 {
				return uint16(val), true
			}
		case int:
			if val >= 0 && val <= math.MaxUint16 {
				return uint16(val), true
			}
		case int8:
			if val >= 0 {
				return uint16(val), true
			}
		case int16:
			if val >= 0 {
				return uint16(val), true
			}
		case int32:
			if val >= 0 && val <= math.MaxUint16 {
				return uint16(val), true
			}
		case int64:
			if val >= 0 && val <= math.MaxUint16 {
				return uint16(val), true
			}
		case float32:
			if val >= 0 && val <= float32(math.MaxUint16) {
				return uint16(val), true
			}
		case float64:
			if val >= 0 && val <= float64(math.MaxUint16) {
				return uint16(val), true
			}
		}
	}
	return 0, false
}

// GetUint16Slice 获取[]uint16类型值
func (m KSMap) GetUint16Slice(key string) ([]uint16, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []uint16:
			return val, true
		case []any:
			nums := make([]uint16, 0, len(val))
			for _, item := range val {
				if i, ok := item.(uint16); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetUint64 获取uint64类型值
func (m KSMap) GetUint64(key string) (uint64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case uint64:
			return val, true
		case uint:
			return uint64(val), true
		case uint8:
			return uint64(val), true
		case uint16:
			return uint64(val), true
		case uint32:
			return uint64(val), true
		case int:
			if val >= 0 {
				return uint64(val), true
			}
		case int8:
			if val >= 0 {
				return uint64(val), true
			}
		case int16:
			if val >= 0 {
				return uint64(val), true
			}
		case int32:
			if val >= 0 {
				return uint64(val), true
			}
		case int64:
			if val >= 0 {
				return uint64(val), true
			}
		case float32:
			if val >= 0 && val <= float32(math.MaxUint64) {
				return uint64(val), true
			}
		case float64:
			if val >= 0 && val <= float64(math.MaxUint64) {
				return uint64(val), true
			}
		}
	}
	return 0, false
}

// GetUint64Slice 获取[]uint64类型值
func (m KSMap) GetUint64Slice(key string) ([]uint64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []uint64:
			return val, true
		case []any:
			nums := make([]uint64, 0, len(val))
			for _, item := range val {
				if i, ok := item.(uint64); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetFloat32 获取float32类型值
func (m KSMap) GetFloat32(key string) (float32, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float32:
			return val, true
		case float64:
			if val >= -math.MaxFloat32 && val <= math.MaxFloat32 {
				return float32(val), true
			}
		case int:
			return float32(val), true
		case int8:
			return float32(val), true
		case int16:
			return float32(val), true
		case int32:
			return float32(val), true
		case int64:
			if val >= math.MinInt32 && val <= math.MaxInt32 {
				return float32(val), true
			}
		case uint:
			return float32(val), true
		case uint8:
			return float32(val), true
		case uint16:
			return float32(val), true
		case uint32:
			return float32(val), true
		case uint64:
			//if val <= uint64(math.MaxFloat32) {
			return float32(val), true
			//}
		}
	}
	return 0, false
}

// GetFloat32Slice 获取[]float32类型值
func (m KSMap) GetFloat32Slice(key string) ([]float32, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []float32:
			return val, true
		case []any:
			nums := make([]float32, 0, len(val))
			for _, item := range val {
				if i, ok := item.(float32); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetFloat64 获取float64类型值
func (m KSMap) GetFloat64(key string) (float64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val, true
		case float32:
			return float64(val), true
		case int:
			return float64(val), true
		case int8:
			return float64(val), true
		case int16:
			return float64(val), true
		case int32:
			return float64(val), true
		case int64:
			return float64(val), true
		case uint:
			return float64(val), true
		case uint8:
			return float64(val), true
		case uint16:
			return float64(val), true
		case uint32:
			return float64(val), true
		case uint64:
			//if val <= uint64(math.MaxFloat64) {
			return float64(val), true
			//}
		}
	}
	return 0, false
}

// GetFloat64Slice 获取[]float64类型值
func (m KSMap) GetFloat64Slice(key string) ([]float64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []float64:
			return val, true
		case []any:
			nums := make([]float64, 0, len(val))
			for _, item := range val {
				if i, ok := item.(float64); ok {
					nums = append(nums, i)
				} else {
					return nil, false
				}
			}
			return nums, true
		}
	}
	return nil, false
}

// GetBool 获取bool类型值
func (m KSMap) GetBool(key string) (bool, bool) {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// GetBoolSlice 获取[]bool类型值
func (m KSMap) GetBoolSlice(key string) ([]bool, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []bool:
			return val, true
		case []any:
			bools := make([]bool, 0, len(val))
			for _, item := range val {
				if b, ok := item.(bool); ok {
					bools = append(bools, b)
				} else {
					return nil, false
				}
			}
			return bools, true
		}
	}
	return nil, false
}

// GetString 获取string类型值
func (m KSMap) GetString(key string) (string, bool) {
	if v, ok := m[key]; ok {
		if str, ok := v.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetStringSlice 获取[]string类型值
func (m KSMap) GetStringSlice(key string) ([]string, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []string:
			return val, true
		case []any:
			strs := make([]string, 0, len(val))
			for _, item := range val {
				if str, ok := item.(string); ok {
					strs = append(strs, str)
				} else {
					return nil, false
				}
			}
			return strs, true
		}
	}
	return nil, false
}

// GetMap 获取Maps类型值
func (m KSMap) GetMap(key string) (KSMap, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case KSMap:
			return val, true
		case map[string]any:
			return val, true
		}
	}
	return nil, false
}

// GetMapSlice 获取[]Maps类型值
func (m KSMap) GetMapSlice(key string) ([]KSMap, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []KSMap:
			return val, true
		case []any:
			ksMaps := make([]KSMap, 0, len(val))
			for _, item := range val {
				if maps, ok := item.(KSMap); ok {
					ksMaps = append(ksMaps, maps)
				} else {
					return nil, false
				}
			}
			return ksMaps, true
		}
	}
	return nil, false
}

// GetBytes 获取[]byte类型值
func (m KSMap) GetBytes(key string) ([]byte, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []byte:
			return val, true
		case string:
			return []byte(val), true
		}
	}
	return nil, false
}

// Has 判断是否存在指定key
func (m KSMap) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// Len 获取map长度
func (m KSMap) Len() int {
	return len(m)
}

// Keys 获取所有key
func (m KSMap) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
