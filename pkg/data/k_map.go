package data

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

// SetSlice 设置[]any类型值
func (m KSMap) SetSlice(key string, value *[]any) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetInt 设置int类型值
func (m KSMap) SetInt(key string, value *int) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetIntSlice 设置[]int类型值
func (m KSMap) SetIntSlice(key string, value *[]int) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetInt8 设置int8类型值
func (m KSMap) SetInt8(key string, value *int8) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetInt8Slice 设置[]int8类型值
func (m KSMap) SetInt8Slice(key string, value *[]int8) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetInt16 设置int16类型值
func (m KSMap) SetInt16(key string, value *int16) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetInt16Slice 设置[]int16类型值
func (m KSMap) SetInt16Slice(key string, value *[]int16) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetInt64 设置int64类型值
func (m KSMap) SetInt64(key string, value *int64) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetInt64Slice 设置[]int64类型值
func (m KSMap) SetInt64Slice(key string, value *[]int64) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetUint 设置uint类型值
func (m KSMap) SetUint(key string, value *uint) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetUintSlice 设置[]uint类型值
func (m KSMap) SetUintSlice(key string, value *[]uint) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetUint8 设置uint8类型值
func (m KSMap) SetUint8(key string, value *uint8) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetUint8Slice 设置[]uint8类型值
func (m KSMap) SetUint8Slice(key string, value *[]uint8) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetUint16 设置uint16类型值
func (m KSMap) SetUint16(key string, value *uint16) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetUint16Slice 设置[]uint16类型值
func (m KSMap) SetUint16Slice(key string, value *[]uint16) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetUint64 设置uint64类型值
func (m KSMap) SetUint64(key string, value *uint64) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetUint64Slice 设置[]uint64类型值
func (m KSMap) SetUint64Slice(key string, value *[]uint64) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetFloat32 设置float32类型值
func (m KSMap) SetFloat32(key string, value *float32) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetFloat32Slice 设置[]float32类型值
func (m KSMap) SetFloat32Slice(key string, value *[]float32) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetFloat64 设置float64类型值
func (m KSMap) SetFloat64(key string, value *float64) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetFloat64Slice 设置[]float64类型值
func (m KSMap) SetFloat64Slice(key string, value *[]float64) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetBool 设置bool类型值
func (m KSMap) SetBool(key string, value *bool) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetBoolSlice 设置[]bool类型值
func (m KSMap) SetBoolSlice(key string, value *[]bool) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetString 设置string类型值
func (m KSMap) SetString(key string, value *string) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetStringSlice 设置[]string类型值
func (m KSMap) SetStringSlice(key string, value *[]string) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetMap 设置Maps类型值
func (m KSMap) SetMap(key string, value *KSMap) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetMapSlice 设置[]Maps类型值
func (m KSMap) SetMapSlice(key string, value *[]KSMap) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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

// SetBytes 设置[]byte类型值
func (m KSMap) SetBytes(key string, value *[]byte) {
	if value == nil {
		delete(m, key)
		return
	}
	m[key] = *value
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
