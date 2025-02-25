package err

import (
	"fmt"
	"strings"
)

// CodeErrs 定义多错误结构
type CodeErrs struct {
	code      int
	errs      []error
	localeIds []string
	templates []map[string]any
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

// WrapLocalize 添加本地化信息
func (c *CodeErrs) WrapLocalize(localeId string, template map[string]any) *CodeErrs {
	c.localeIds = append(c.localeIds, localeId)
	if template == nil {
		template = map[string]any{}
	}
	c.templates = append(c.templates, template)
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

// LocaleIds 获取本地化ID
func (c *CodeErrs) LocaleIds() []string {
	return c.localeIds
}

// Templates 获取模板数据
func (c *CodeErrs) Templates() []map[string]any {
	return c.templates
}

// Err 获取错误
func (c *CodeErrs) Err() error {
	if len(c.errs) == 0 {
		return nil
	}
	return c
}

func (c *CodeErrs) ToLocales(fun func(string, map[string]any) string) string {
	if len(c.localeIds) > 0 {
		var builder strings.Builder
		for k, v := range c.localeIds {
			msg := fun(v, c.templates[k])
			builder.WriteString(msg)
			if k < len(c.localeIds)-1 {
				builder.WriteString("\n")
			}
		}
		return builder.String()
	}
	return fmt.Sprintf("%d: unknown error(localeIds)", c.code)
}
