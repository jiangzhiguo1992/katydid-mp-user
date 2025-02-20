package err

import "strings"

// Error 定义带前后缀的错误结构
type Error struct {
	err    error
	prefix string
	suffix string
}

// NewError 创建新的 Error
func NewError(err error) *Error {
	return &Error{err: err}
}

// WithPrefix 设置前缀
func (e *Error) WithPrefix(prefix string) *Error {
	e.prefix = prefix
	return e
}

// WithSuffix 设置后缀
func (e *Error) WithSuffix(suffix string) *Error {
	e.suffix = suffix
	return e
}

// Error 实现 error 接口
func (e *Error) Error() string {
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
