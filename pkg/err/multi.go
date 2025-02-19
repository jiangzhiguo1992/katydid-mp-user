package err

import (
	"strings"
)

// MultiCodeError 定义多错误结构
type MultiCodeError struct {
	code   int
	errors []*CodeError
}

func NewMultiError(err error) *MultiCodeError {
	return &MultiCodeError{
		errors: []*CodeError{NewCodeError(err)},
	}
}

// WithCode 设置错误码
func (m *MultiCodeError) WithCode(code int) *MultiCodeError {
	m.code = code
	return m
}

// WrapError 添加新错误
func (m *MultiCodeError) WrapError(err error) *MultiCodeError {
	return m.WrapCodeError(NewCodeError(err))
}

func (m *MultiCodeError) WrapCodeError(err *CodeError) *MultiCodeError {
	m.errors = append(m.errors, err)
	return m
}

// Error 实现 error 接口
func (m *MultiCodeError) Error() string {
	if len(m.errors) == 0 {
		return ""
	}

	if len(m.errors) == 1 {
		return m.errors[0].Error()
	}

	var builder strings.Builder
	builder.WriteString("Multiple errors:")
	for _, err := range m.errors {
		builder.WriteString("\n\t- ")
		builder.WriteString(err.Error())
	}
	return builder.String()
}

// Code 获取错误码
func (m *MultiCodeError) Code() int {
	return m.code
}

// Unwrap 实现错误链
func (m *MultiCodeError) Unwrap() error {
	if len(m.errors) == 0 {
		return nil
	}
	return m.errors[0]
}
