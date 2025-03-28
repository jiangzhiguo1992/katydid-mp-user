package errs

import (
	"fmt"
	"io"
	"strings"
)

var (
	defLang = "zh"

	translate = func(lang, templateID string, data map[string]any, params ...any) string {
		return fmt.Sprintf("%v, %v, %v", templateID, data, params)
	}
)

type (
	Error struct {
		// source 不会返回给client
		cause  error    // 触发错误的源
		stack  *stack   // 错误栈
		traces []string // 链路 需外部主动append
		errs   []error  // 多error 需外部主动append
		// message 会返回给client
		code    int       // 错误码 new/match的时候确定
		msg     string    // 错误信息 new/match的时候确定
		locales []*locale // 多国际化 match会确定，也可以外部主动append
	}

	locale struct {
		templateID string         // 国际化模板ID
		data       map[string]any // 命名参数 {{s}}
		params     []any          // 位置参数 %s
	}
)

func Init(
	trans func(lang, templateID string, data map[string]any, params ...any) string,
) {
	if trans != nil {
		translate = trans
	}
}

func New(cause error) *Error {
	return &Error{
		cause: cause, stack: callers(),
		code: 0, msg: cause.Error(),
	}
}

func (e *Error) ResetStack() *Error {
	e.stack = callers()
	return e
}

func (e *Error) WithTrace(trace string) *Error {
	e.traces = append(e.traces, trace)
	return e
}

func (e *Error) AppendErrors(errs ...error) *Error {
	for _, err := range errs {
		if err == nil {
			continue
		}
		e.errs = append(e.errs, err)
	}
	return e
}

func (e *Error) WithCode(code int) *Error {
	e.code = code
	return e
}

func (e *Error) WithMsg(msg string) *Error {
	e.msg = msg
	return e
}

func (e *Error) WithMsgf(format string, args ...any) *Error {
	e.msg = fmt.Sprintf(format, args...)
	return e
}

func (e *Error) AppendLocale(templateID string, data map[string]any, params ...any) *Error {
	if templateID == "" {
		return e
	}
	e.locales = append(e.locales, &locale{
		templateID: templateID,
		data:       data,
		params:     params,
	})
	return e
}

// Error 实现 error 接口
func (e *Error) Error() string {
	var builder strings.Builder

	if e.code != 0 {
		builder.WriteString("code: ")
		builder.WriteString(fmt.Sprintf("%d", e.code))
		builder.WriteString(" ")
	}

	if e.msg != "" {
		builder.WriteString("msg: ")
		builder.WriteString(e.msg)
		builder.WriteString(" ")
	}

	if len(e.traces) > 0 {
		builder.WriteString("\ntraces: ")
		for _, trace := range e.traces {
			builder.WriteString(trace)
			builder.WriteString(" <- ")
		}
	}

	if e.cause != nil {
		builder.WriteString("\nerror:")
		builder.WriteString("\n\t- ")
		builder.WriteString(e.cause.Error())
	}

	if len(e.errs) > 0 {
		builder.WriteString(fmt.Sprintf("\nerrors (%d):", len(e.errs)))
		for _, err := range e.errs {
			builder.WriteString("\n\t- ")
			builder.WriteString(err.Error())
		}
	}

	if len(e.locales) > 0 {
		builder.WriteString(fmt.Sprintf("\nlocales (%d):", len(e.locales)))
		for _, loc := range e.locales {
			builder.WriteString("\n\t- ")
			res := translate(defLang, loc.templateID, loc.data, loc.params)
			builder.WriteString(res)
		}
	}

	return builder.String()
}

// Unwrap 为 Go 1.13 错误链提供兼容性
func (e *Error) Unwrap() error { return e.cause }

// Format 实现 fmt.Formatter 接口
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", e.cause)
			e.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

func (e *Error) IsNil() bool {
	return (e.cause == nil) && (len(e.errs) == 0) && (e.code == 0) && (e.msg == "") && (len(e.locales) == 0)
}

// Wash 清洗错误，返回虚假的错误
func (e *Error) Wash() *Error {
	if e.IsNil() {
		return nil
	}
	return e
}

// Translate 返回国际化信息
func (e *Error) Translate(lang string) string {
	if (translate == nil) || (len(e.locales) == 0) {
		if e.msg != "" {
			return translate(lang, e.msg, nil)
		}
		return ""
	}

	if len(e.locales) == 1 {
		v := e.locales[0]
		if v.templateID != "" {
			return translate(lang, v.templateID, v.data, v.params)
		}
		return ""
	}

	var builder strings.Builder
	for k, v := range e.locales {
		if v.templateID == "" {
			continue
		}
		msg := translate(lang, v.templateID, v.data, v.params)
		builder.WriteString(msg)
		if k < len(e.locales)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}
