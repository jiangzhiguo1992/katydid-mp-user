package err

import "strings"

// CodeError 定义带错误码的错误结构
type CodeError struct {
	code   int
	err    error
	prefix string
	suffix string
}

// NewCodeError 创建新的 CodeError
func NewCodeError(err error) *CodeError {
	return &CodeError{err: err}
}

// WithCode 设置错误码
func (e *CodeError) WithCode(code int) *CodeError {
	e.code = code
	return e
}

// WithPrefix 设置前缀
func (e *CodeError) WithPrefix(prefix string) *CodeError {
	e.prefix = prefix
	return e
}

// WithSuffix 设置后缀
func (e *CodeError) WithSuffix(suffix string) *CodeError {
	e.suffix = suffix
	return e
}

// Error 实现 error 接口
func (e *CodeError) Error() string {
	parts := make([]string, 0, 3)
	if e.prefix != "" {
		parts = append(parts, e.prefix+": ")
	}
	if e.err != nil {
		parts = append(parts, e.err.Error())
	} else {
		parts = append(parts, "_nil_")
	}
	if e.suffix != "" {
		parts = append(parts, ": "+e.suffix)
	}
	return strings.Join(parts, "")
}

// Code 获取错误码
func (e *CodeError) Code() int {
	if e == nil {
		return CodeSuccess
	}
	return e.code
}
