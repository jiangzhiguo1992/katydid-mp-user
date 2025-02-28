package perm

import (
	"fmt"
)

const (
	PolicyActAdd = "add"
	PolicyActDel = "del"
	PolicyActMod = "mod"
	PolicyActGet = "get"
)

// Policy 权限策略定义
type Policy struct {
	Sub string `json:"sub"` // 主体
	Obj string `json:"obj"` // 资源
	Act string `json:"act"` // 动作
	//Lv  int    `json:"lv"`  // 等级
}

// NewPolicy 创建一个新的空策略
func NewPolicy() *Policy {
	return &Policy{}
}

// NewPolicyFromTriple 从三元组构建策略
func NewPolicyFromTriple(sub, obj, act string) *Policy {
	return &Policy{
		Sub: sub,
		Obj: obj,
		Act: act,
	}
}

// NewPolicyFromRequest 从HTTP请求信息构建策略
func NewPolicyFromRequest(subject, path, method string) *Policy {
	rule := NewRequestRule(path, method)
	policy := rule.Convert()
	policy.Sub = subject
	return policy
}

// WithSubject 设置策略主体
func (p *Policy) WithSubject(sub string) *Policy {
	p.Sub = sub
	return p
}

// WithObject 设置策略资源
func (p *Policy) WithObject(obj string) *Policy {
	p.Obj = obj
	return p
}

// WithAction 设置策略动作
func (p *Policy) WithAction(act string) *Policy {
	p.Act = act
	return p
}

// Build 构建完整策略，确认所有必须字段已设置
func (p *Policy) Build() (*Policy, error) {
	if !p.IsValid() {
		return nil, fmt.Errorf("casbin invalid policy: missing required fields")
	}
	return p, nil
}

// IsValid 检查策略是否有效
func (p *Policy) IsValid() bool {
	return p.Sub != "" && p.Obj != "" && p.Act != ""
}

// String 返回策略的字符串表示
func (p *Policy) String() string {
	return fmt.Sprintf("%s, %s, %s", p.Sub, p.Obj, p.Act)
}

// Clone 复制策略
func (p *Policy) Clone() *Policy {
	return &Policy{
		Sub: p.Sub,
		Obj: p.Obj,
		Act: p.Act,
	}
}

// Check 验证当前策略是否通过
func (p *Policy) Check() (bool, error) {
	return Enforcer(p)
}
