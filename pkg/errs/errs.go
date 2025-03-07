package errs

import (
	"fmt"
	"strings"
)

// CodeErrs 定义多错误结构
type (
	CodeErrs struct {
		code      int
		errs      []error
		localizes []*localize
	}
	localize struct {
		localeId  string
		template1 []any
		template2 map[string]any
	}
)

func New(errs ...error) *CodeErrs {
	var validErrs []error
	for _, err := range errs {
		if err == nil {
			continue
		}
		validErrs = append(validErrs, err)
	}
	return &CodeErrs{errs: validErrs}
}

// WithCode 设置错误码
func (c *CodeErrs) WithCode(code int) *CodeErrs {
	c.code = code
	return c
}

// WrapErrs 添加新错误
func (c *CodeErrs) WrapErrs(errs ...error) *CodeErrs {
	for _, err := range errs {
		if err == nil {
			continue
		}
		c.errs = append(c.errs, err)
	}
	return c
}

// WrapLocalize 添加本地化信息
func (c *CodeErrs) WrapLocalize(localeId string, template1 []any, template2 map[string]any) *CodeErrs {
	if localeId == "" {
		return c
	}
	c.localizes = append(c.localizes, &localize{
		localeId:  localeId,
		template1: template1,
		template2: template2,
	})
	return c
}

// WrapCodeErrs 添加新错误
func (c *CodeErrs) WrapCodeErrs(errs ...*CodeErrs) *CodeErrs {
	for _, v := range errs {
		if v == nil {
			continue
		}
		_ = c.WrapErrs(v.Errs()...)
		for _, vv := range v.localizes {
			_ = c.WrapLocalize(vv.localeId, vv.template1, vv.template2)
		}
	}
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

func (c *CodeErrs) IsNil() bool {
	return (c.Err() == nil) && (len(c.localizes) == 0)
}

func (c *CodeErrs) Real() *CodeErrs {
	if c.IsNil() {
		return nil
	}
	return c
}

func (c *CodeErrs) ToLocales(fun func(string, []any, map[string]any) string) string {
	if len(c.localizes) == 0 {
		return fmt.Sprintf("%d: unknown error(localeIds)", c.code)
	}
	var builder strings.Builder
	for k, v := range c.localizes {
		msg := fun(v.localeId, v.template1, v.template2)
		builder.WriteString(msg)
		if k < len(c.localizes)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}
