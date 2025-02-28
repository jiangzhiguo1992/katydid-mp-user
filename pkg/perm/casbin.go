package perm

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

var (
	enforcer *casbin.Enforcer
)

type (
	Policy struct {
		Sub string `json:"sub"` // 主体
		Obj string `json:"obj"` // 资源
		Act string `json:"act"` // 动作
		Lv  int    `json:"lv"`  // 等级
	}
)

// Init initializes the Casbin enforcer with an in-memory model.
// In real projects, you can load the model and policy from files or database.
func Init() {
	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act, lv

[policy_definition]
p = sub, obj, act, lv

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act && r.lv >= p.lv
`)
	if err != nil {
		panic(fmt.Sprintf("■ ■ Casbin ■ ■ failed to create m: %v", err))
	}

	enforcer, err = casbin.NewEnforcer(m)
	if err != nil {
		panic(fmt.Sprintf("■ ■ Casbin ■ ■ failed to create enforcer: %v", err))
	}

	policies := []*Policy{
		{"alice", "/data1", "GET", 2},
		{"bob", "/data2", "POST", 1},
		// Add more policies as needed
	}

	// 注册策略
	Register(policies)
}

func Get() *casbin.Enforcer {
	return enforcer
}

//func NewPolicy(role, path, method string) *Policy {
//
//}

func Enforcer(policy *Policy) (bool, error) {
	ok, err := Get().Enforce(policy.Sub, policy.Obj, policy.Act)
	// TODO:GG 还需要其他处理吗?
	return ok, err
}

func Register(policies []*Policy) {
	for _, p := range policies {
		if ok, err := enforcer.AddPolicy(p.Sub, p.Obj, p.Act); err != nil || !ok {
			slog.Error("■ ■ Casbin ■ ■ AddPolicy Failed", slog.Any("err", err))
		}
	}
}

// ActionMap 定义 HTTP 方法到 Casbin 动作的映射
var ActionMap = map[string]string{
	http.MethodGet:     "read",
	http.MethodPost:    "create",
	http.MethodPut:     "update",
	http.MethodDelete:  "delete",
	http.MethodPatch:   "update",
	http.MethodHead:    "read",
	http.MethodOptions: "read",
}

// ResourceRule 用于转换 API 路径到资源规则
type ResourceRule struct {
	Path       string            // API 路径
	Method     string            // HTTP 方法
	Resource   string            // 资源名称
	Action     string            // 操作类型
	ExtraRules map[string]string // 额外的规则映射
}

// NewResourceRule 创建资源规则转换器
func NewResourceRule(path, method string) *ResourceRule {
	return &ResourceRule{
		Path:       path,
		Method:     method,
		ExtraRules: make(map[string]string),
	}
}

// Convert 转换 HTTP 请求到 Casbin 规则
func (r *ResourceRule) Convert() *Policy {
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
func (r *ResourceRule) parseResource() string {
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
func (r *ResourceRule) parseAction() string {
	if action, exists := ActionMap[r.Method]; exists {
		return action
	}
	return "unknown"
}

// WithExtraRules 添加额外的规则映射
func (r *ResourceRule) WithExtraRules(rules map[string]string) *ResourceRule {
	for k, v := range rules {
		r.ExtraRules[k] = v
	}
	return r
}
