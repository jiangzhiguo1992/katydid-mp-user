package err

import (
	"fmt"
	"strings"
)

// CodeErrs 定义多错误结构
type CodeErrs struct {
	code      int
	errs      []error
	localeId  string
	templates map[string]any
}

func New(errs ...error) *CodeErrs {
	return &CodeErrs{errs: errs}
}

// WithCode 设置错误码
func (c *CodeErrs) WithCode(code int) *CodeErrs {
	c.code = code
	return c
}

// WrapErrs 添加新错误
func (c *CodeErrs) WrapErrs(errs ...error) *CodeErrs {
	c.errs = append(c.errs, errs...)
	return c
}

// WithLocalize 添加本地化信息
func (c *CodeErrs) WithLocalize(localeId string, templates map[string]any) *CodeErrs {
	c.localeId = localeId
	if templates == nil {
		templates = map[string]any{}
	}
	c.templates = templates
	return c
}

// Error 实现 error 接口
func (c *CodeErrs) Error() string {
	if len(c.errs) == 0 {
		return fmt.Sprintf("CodeErrs (%d)", c.code)
	} else if len(c.errs) == 1 {
		return fmt.Sprintf("CodeErrs (%d): %s", c.code, c.errs[0].Error())
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("CodeErrs (%d):", c.code))
	for _, err := range c.errs {
		builder.WriteString("\n\t- ")
		builder.WriteString(err.Error())
	}
	return builder.String()
}

// Code 获取错误码
func (c *CodeErrs) Code() int {
	return c.code
}

// Errs 获取错误
func (c *CodeErrs) Errs() []error {
	return c.errs
}

// Err 获取错误
func (c *CodeErrs) Err() error {
	if len(c.errs) == 0 {
		return nil
	}
	return c
}

func (c *CodeErrs) ToLocales(fun func(string, map[string]any) string) string {
	locale := fun(c.localeId, c.templates)
	if len(locale) == 0 {
		return fmt.Sprintf("%d: unknown error(localeIds)", c.code)
	}
	return locale
}
