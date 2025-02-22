package err

import (
	"fmt"
	"strings"
)

// CodeErrors 定义多错误结构
type CodeErrors struct {
	errors []error

	code   int
	prefix string
	suffix string

	locales   []string
	templates []map[string]any
}

func NewCodeError(err error) *CodeErrors {
	return &CodeErrors{errors: []error{err}}
}

func NewCodeErrors(errs []error) *CodeErrors {
	return &CodeErrors{errors: errs}
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

// WrapLocale 添加本地化信息
func (c *CodeErrors) WrapLocale(locale string, template map[string]any) *CodeErrors {
	c.locales = append(c.locales, locale)
	if template == nil {
		template = map[string]any{}
	}
	c.templates = append(c.templates, template)
	return c
}

// WrapLocales 添加本地化信息
func (c *CodeErrors) WrapLocales(locales []string, templates []map[string]any) *CodeErrors {
	c.locales = append(c.locales, locales...)
	if templates == nil {
		templates = make([]map[string]any, len(locales))
	}
	c.templates = append(c.templates, templates...)
	return c
}

// Error 实现 error 接口
func (c *CodeErrors) Error() string {
	if len(c.errors) == 0 {
		if len(c.locales) > 0 {
			return c.ToLocales(nil)
		}
		return fmt.Sprintf("CodeErrors (%d)", c.code)
	} else if len(c.errors) == 1 {
		return fmt.Sprintf("CodeErrors (%d): %s", c.code, c.errors[0].Error())
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("CodeErrors (%d):", c.code))
	for _, err := range c.errors {
		builder.WriteString("\n\t- ")
		builder.WriteString(err.Error())
	}
	return builder.String()
}

// Err 获取错误
func (c *CodeErrors) Err() error {
	if len(c.errors) == 0 {
		return nil
	}
	return c
}

// Errs 获取错误
func (c *CodeErrors) Errs() []error {
	return c.errors
}

// Code 获取错误码
func (c *CodeErrors) Code() int {
	return c.code
}

func (c *CodeErrors) ToLocales(fun func([]string, []map[string]any) []string) string {
	locales := c.locales
	if fun != nil {
		locales = fun(c.locales, c.templates)
	}
	if len(locales) == 0 {
		return fmt.Sprintf("%d: unknown error(locales)", c.code)
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%d ", c.code))
	for _, l := range locales {
		builder.WriteString(" - ")
		builder.WriteString(l)
	}
	return builder.String()
}
