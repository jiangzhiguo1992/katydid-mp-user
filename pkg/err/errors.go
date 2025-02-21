package err

import (
	"fmt"
	"strings"
)

// CodeErrors 定义多错误结构
type CodeErrors struct {
	code   int
	prefix string
	suffix string
	errors []error
}

func NewCodeError(err error) *CodeErrors {
	return &CodeErrors{
		errors: []error{err},
	}
}

func NewCodeErrors(errs []error) *CodeErrors {
	return &CodeErrors{
		errors: errs,
	}
}

// WithCode 设置错误码
func (c *CodeErrors) WithCode(code int) *CodeErrors {
	c.code = code
	return c
}

// WithPrefix 设置前缀
func (c *CodeErrors) WithPrefix(prefix string) *CodeErrors {
	c.prefix = prefix
	return c
}

// WithSuffix 设置后缀
func (c *CodeErrors) WithSuffix(suffix string) *CodeErrors {
	c.suffix = suffix
	return c
}

// WrapError 添加新错误
func (c *CodeErrors) WrapError(err error) *CodeErrors {
	c.errors = append(c.errors, err)
	return c
}

// WrapErrors 添加新错误
func (c *CodeErrors) WrapErrors(errs []error) *CodeErrors {
	c.errors = append(c.errors, errs...)
	return c
}

// Error 实现 error 接口
func (c *CodeErrors) Error() string {
	if len(c.errors) == 0 {
		return "0 errors"
	} else if len(c.errors) == 1 {
		return fmt.Sprintf("Error (%d): %s", c.code, c.errors[0].Error())
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Errors (%d):\n\t", c.code))
	for _, err := range c.errors {
		builder.WriteString("\n\t- ")
		builder.WriteString(err.Error())
	}
	return builder.String()
}

// Code 获取错误码
func (c *CodeErrors) Code() int {
	return c.code
}

// Unwrap 实现错误链
func (c *CodeErrors) Unwrap() error {
	if len(c.errors) == 0 {
		return nil
	}
	return c.errors[0]
}

func (c *CodeErrors) Err() error {
	if len(c.errors) == 0 {
		return nil
	}
	return c
}
