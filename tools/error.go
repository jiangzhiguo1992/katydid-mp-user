package tools

import (
	"fmt"
)

// MultiCodeError 是一个可以包裹多层错误的结构体
type MultiCodeError struct {
	Code   int
	Errors []*CodeError
}

func NewMultiCodeError(err error) *MultiCodeError {
	return &MultiCodeError{Code: 0, Errors: []*CodeError{NewCodeError(err)}}
}

func (m *MultiCodeError) WithCode(code int) *MultiCodeError {
	m.Code = code
	return m
}

func (m *MultiCodeError) Error() string {
	if len(m.Errors) == 0 {
		return ""
	}
	result := "Multiple:"
	for _, err := range m.Errors {
		result += fmt.Sprintf("\n\t-%s", err.Error())
	}
	return result
}

// WrapError 包裹一个新的错误
func (m *MultiCodeError) WrapError(err error) *MultiCodeError {
	return m.WrapCodeError(NewCodeError(err))
}

func (m *MultiCodeError) WrapCodeError(err *CodeError) *MultiCodeError {
	m.Errors = append(m.Errors, err)
	return m
}

// Unwrap 返回第一个错误
func (m *MultiCodeError) Unwrap() error {
	if len(m.Errors) > 0 {
		return m.Errors[0]
	}
	return nil
}

// CodeError 是一个可以包裹错误码的结构体
type CodeError struct {
	Code   int
	Err    error
	Prefix string
	Suffix string
}

func NewCodeError(err error) *CodeError {
	return &CodeError{Code: 0, Err: err, Prefix: "", Suffix: ""}
}

func (c *CodeError) WithCode(code int) *CodeError {
	c.Code = code
	return c
}

func (c *CodeError) WithPrefix(prefix string) *CodeError {
	c.Prefix = prefix
	return c
}

func (c *CodeError) WithSuffix(suffix string) *CodeError {
	c.Suffix = suffix
	return c
}

func (c *CodeError) Error() string {
	prefix := ""
	if len(c.Prefix) > 0 {
		prefix = fmt.Sprintf("%s_: ", c.Prefix)
	}
	suffix := ""
	if len(c.Suffix) > 0 {
		suffix = fmt.Sprintf(" :_%s", c.Suffix)
	}
	return fmt.Sprintf("%s%s%s", prefix, c.Err.Error(), suffix)
}
