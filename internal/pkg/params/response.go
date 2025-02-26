package params

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/i18n"
)

func Response(c *gin.Context, code int, msg string, data any) {
	if e, ok := data.(error); ok {
		ResponseErr(c, code, e)
		return
	}
	response(c, code, msg, data)
}

func ResponseErr(c *gin.Context, code int, data error) {
	lang := c.GetHeader("Accept-Language")

	var cErr *err.CodeErrs
	if v, ok := data.(*err.CodeErrs); !ok {
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

// TODO:GG 根据accesscontentTypes来判断返回json还是prof?
func response(c *gin.Context, code int, msg string, data any) {
	// TODO:GG 有些字段，返回的时候是要忽略的
	c.JSON(code, gin.H{"code": code, "msg": msg, "data": data})
}
