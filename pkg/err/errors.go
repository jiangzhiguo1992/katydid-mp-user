package err

import (
	"strings"
)

// CodeErrors 定义多错误结构
type CodeErrors struct {
	code   int
	errors []*Error
}

func NewCodeErrors(err error) *CodeErrors {
	return &CodeErrors{
		errors: []*Error{NewError(err)},
	}
}

// WithCode 设置错误码
func (m *CodeErrors) WithCode(code int) *CodeErrors {
	m.code = code
	return m
}

// WrapError 添加新错误
func (m *CodeErrors) WrapError(err *Error) *CodeErrors {
	m.errors = append(m.errors, err)
	return m
}

// Error 实现 error 接口
func (m *CodeErrors) Error() string {
	if len(m.errors) == 0 {
		return ""
	}

	if len(m.errors) == 1 {
		return m.errors[0].Error()
	}

	var builder strings.Builder
	builder.WriteString("Errors:")
	for _, err := range m.errors {
		builder.WriteString("\n\t- ")
		builder.WriteString(err.Error())
	}
	return builder.String()
}

// Code 获取错误码
func (m *CodeErrors) Code() int {
	return m.code
}

// Unwrap 实现错误链
func (m *CodeErrors) Unwrap() error {
	if len(m.errors) == 0 {
		return nil
	}
	return m.errors[0]
}
