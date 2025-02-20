package utils

type Maps map[string]any

func (m Maps) GetString(key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func (m Maps) GetUint64(key string) (uint64, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(uint64); ok {
			return s, true
		}
	}
	return 0, false
}
