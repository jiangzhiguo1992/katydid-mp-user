package utils

// Maps 扩展 map[string]any 类型
type Maps map[string]any

// GetInt 获取int类型值
func (m Maps) GetInt(key string) (int, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val, true
		case int64:
			return int(val), true
		case int32:
			return int(val), true
		case float64:
			return int(val), true
		}
	}
	return 0, false
}

// GetInt64 获取int64类型值
func (m Maps) GetInt64(key string) (int64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val, true
		case int:
			return int64(val), true
		case float64:
			return int64(val), true
		}
	}
	return 0, false
}

// GetFloat64 获取float64类型值
func (m Maps) GetFloat64(key string) (float64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val, true
		case float32:
			return float64(val), true
		case int:
			return float64(val), true
		case int64:
			return float64(val), true
		}
	}
	return 0, false
}

// GetFloat32 获取float32类型值
func (m Maps) GetFloat32(key string) (float32, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float32:
			return val, true
		case float64:
			return float32(val), true
		case int:
			return float32(val), true
		}
	}
	return 0, false
}

// GetInt8 获取int8类型值
func (m Maps) GetInt8(key string) (int8, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int8:
			return val, true
		case int:
			if val >= -128 && val <= 127 {
				return int8(val), true
			}
		case float64:
			if val >= -128 && val <= 127 {
				return int8(val), true
			}
		}
	}
	return 0, false
}

// GetInt16 获取int16类型值
func (m Maps) GetInt16(key string) (int16, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int16:
			return val, true
		case int:
			if val >= -32768 && val <= 32767 {
				return int16(val), true
			}
		case float64:
			if val >= -32768 && val <= 32767 {
				return int16(val), true
			}
		}
	}
	return 0, false
}

// GetUint8 获取uint8类型值
func (m Maps) GetUint8(key string) (uint8, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case uint8:
			return val, true
		case int:
			if val >= 0 && val <= 255 {
				return uint8(val), true
			}
		case float64:
			if val >= 0 && val <= 255 {
				return uint8(val), true
			}
		}
	}
	return 0, false
}

// GetUint16 获取uint16类型值
func (m Maps) GetUint16(key string) (uint16, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case uint16:
			return val, true
		case int:
			if val >= 0 && val <= 65535 {
				return uint16(val), true
			}
		case float64:
			if val >= 0 && val <= 65535 {
				return uint16(val), true
			}
		}
	}
	return 0, false
}

// GetUint64 获取uint64类型值
func (m Maps) GetUint64(key string) (uint64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case uint64:
			return val, true
		case uint:
			return uint64(val), true
		case int:
			if val >= 0 {
				return uint64(val), true
			}
		case float64:
			if val >= 0 {
				return uint64(val), true
			}
		}
	}
	return 0, false
}

// GetBool 获取bool类型值
func (m Maps) GetBool(key string) (bool, bool) {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// GetSlice 获取[]any类型值
func (m Maps) GetSlice(key string) ([]any, bool) {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]any); ok {
			return slice, true
		}
	}
	return nil, false
}

// GetStringSlice 获取[]string类型值
func (m Maps) GetStringSlice(key string) ([]string, bool) {
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
func (m Maps) GetMap(key string) (Maps, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case Maps:
			return val, true
		case map[string]any:
			return val, true
		}
	}
	return nil, false
}

// GetBytes 获取[]byte类型值
func (m Maps) GetBytes(key string) ([]byte, bool) {
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

// Get 获取任意类型值，需要自己做类型断言
func (m Maps) Get(key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}

// Set 设置任意类型值
func (m Maps) Set(key string, value any) {
	m[key] = value
}

// Delete 删除指定key
func (m Maps) Delete(key string) {
	delete(m, key)
}

// Has 判断是否存在指定key
func (m Maps) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// Len 获取map长度
func (m Maps) Len() int {
	return len(m)
}

// Keys 获取所有key
func (m Maps) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
