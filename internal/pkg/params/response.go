package params

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/i18n"
	"strings"
)

func Response(c *gin.Context, code int, msg string, data any) {
	if e, ok := data.(error); ok {
		ResponseErr(c, code, e)
		return
	}
	response(c, code, msg, data)
}

func ResponseErr(c *gin.Context, code int, data error) {
	lang := c.GetHeader("Accept-Language") // TODO:GG 需要过滤吗/转换吗?

	var cErr *err.CodeErrs
	var v *err.CodeErrs
	if !errors.As(data, &v) {
		cErr = err.Match(data)
	} else {
		cErr = v
	}

	msg := cErr.ToLocales(func(localize string, template1s []any, template2s map[string]any) string {
		var templates []any
		for _, v := range template1s {
			if _, ok := v.(string); !ok {
				templates = append(templates, v)
				continue
			}
			temp := i18n.LocalizeTry(lang, v.(string), nil)
			templates = append(templates, temp)
		}
		r1 := i18n.LocalizeTry(lang, localize, template2s)
		return fmt.Sprintf(r1, templates...)
	})

	if len(msg) == 0 {
		msg = i18n.LocalizeTry(lang, "unknown_err", nil)
	}

	response(c, code, msg, nil)
}

func response(c *gin.Context, code int, msg string, data any) {
	// TODO:GG 有些字段，返回的时候是要忽略的(利用json:"-"来做吗?)
	body := gin.H{"code": code, "msg": msg, "data": data}

	accept := c.GetHeader("Accept")
	if accept == "" {
		accept = binding.MIMEJSON
	} else if strings.Contains(accept, "*/*") {
		accept = binding.MIMEJSON
	} else if strings.Contains(accept, "application/*") {
		accept = binding.MIMEJSON
	} else if strings.Contains(accept, "text/*") {
		accept = binding.MIMEXML
	}
	if strings.Contains(accept, binding.MIMEJSON) {
		c.JSON(code, body)
	} else if strings.Contains(accept, binding.MIMEPROTOBUF) {
		c.ProtoBuf(code, body)
	} else if strings.Contains(accept, binding.MIMEHTML) {
		c.HTML(code, "", msg)
	} else if strings.Contains(accept, binding.MIMEXML) || strings.Contains(accept, binding.MIMEXML2) {
		c.XML(code, body)
	} else if strings.Contains(accept, binding.MIMETOML) {
		c.TOML(code, body)
	} else if strings.Contains(accept, binding.MIMEYAML) || strings.Contains(accept, binding.MIMEYAML2) {
		c.YAML(code, body)
	} else {
		c.String(code, msg)
	}
}
