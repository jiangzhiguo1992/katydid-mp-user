package perm

import (
	"net/http"
	"strings"
)

// actMap 定义 HTTP 方法到 Casbin 动作的映射
var actMap = map[string]string{
	http.MethodPost:    PolicyActAdd,
	http.MethodDelete:  PolicyActDel,
	http.MethodPut:     PolicyActMod,
	http.MethodGet:     PolicyActGet,
	http.MethodPatch:   PolicyActMod,
	http.MethodHead:    PolicyActGet,
	http.MethodOptions: PolicyActGet,
}

// RequestRule 用于转换 API 路径到请求规则
type RequestRule struct {
	Path   string         // API 路径
	Method string         // HTTP 方法
	Extra  map[string]any // 额外的规则映射
}

// NewRequestRule 创建请求规则转换器
func NewRequestRule(path, method string) *RequestRule {
	return &RequestRule{
		Path:   path,
		Method: method,
		Extra:  make(map[string]any),
	}
}

// Convert 转换 HTTP 请求到 Casbin 规则
func (r *RequestRule) Convert() *Policy {
	// 解析资源名称
	resource := r.parseResource()
	// 获取动作名称
	action := r.parseAction()

	return &Policy{
		Obj: resource,
		Act: action,
	}
}

// parseResource 解析API路径为资源名称
func (r *RequestRule) parseResource() string {
	// 移除路径参数中的具体值，保留结构
	parts := strings.Split(r.Path, "/")
	var result []string

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			// 将路径参数替换为通配符
			result = append(result, "*")
		} else if part != "" {
			result = append(result, part)
		}
	}

	return strings.Join(result, "/")
}

// parseAction 解析HTTP方法为Casbin动作
func (r *RequestRule) parseAction() string {
	if action, exists := actMap[r.Method]; exists {
		return action
	}
	return "unknown"
}

// WithExtraRules 添加额外的规则映射
func (r *RequestRule) WithExtraRules(rules map[string]any) *RequestRule {
	for k, v := range rules {
		r.Extra[k] = v
	}
	return r
}
