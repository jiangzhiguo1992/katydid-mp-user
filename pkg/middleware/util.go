package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"strings"
)

const (
	RequestIDKey     = "RequestID"    // 本地获取的请求ID的上下文键
	XRequestIDHeader = "X-Request-ID" // 网关生成并塞入的requestID
)

func ResponseData(c *gin.Context, code int, obj any) {
	accept := c.GetHeader("Accept")
	if accept == "" || strings.Contains(accept, "*/*") || strings.Contains(accept, "application/*") {
		accept = binding.MIMEJSON
	} else if strings.Contains(accept, "msg/*") {
		accept = binding.MIMEXML
	}

	switch {
	case strings.Contains(accept, binding.MIMEPROTOBUF):
		c.ProtoBuf(code, obj)
	case strings.Contains(accept, binding.MIMEXML), strings.Contains(accept, binding.MIMEXML2):
		c.XML(code, obj)
	case strings.Contains(accept, binding.MIMETOML):
		c.TOML(code, obj)
	case strings.Contains(accept, binding.MIMEYAML), strings.Contains(accept, binding.MIMEYAML2):
		c.YAML(code, obj)
	case strings.Contains(accept, binding.MIMEHTML):
		c.HTML(code, "", obj)
	default:
		c.JSON(code, obj)
	}
}
