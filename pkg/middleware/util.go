package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"strings"
)

const (
	XRequestIDHeader = "X-Request-ID" // 网关生成并塞入的requestID

	RequestIDKey     = "RequestID" // 上下文中存储的请求ID的上下文键
	LanguageKey      = "Use-Language"
	AuthKeyToken     = "token"
	AuthKeyOwnKind   = "ownKind"
	AuthKeyOwnID     = "ownId"
	AuthKeyUserID    = "userId"
	AuthKeyAccountID = "accountId"
)

// ResponseData 统一响应数据
func ResponseData(c *gin.Context, code int, obj any) {
	if c == nil {
		return
	}
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
		c.JSON(code, obj) // 默认使用JSON格式
	}
}
