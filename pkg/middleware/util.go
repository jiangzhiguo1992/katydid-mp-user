package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"strings"
)

const (
	XRequestIDHeader   = "X-Request-ID"   // 网关生成并塞入的requestID
	XRequestPathHeader = "X-Request-Path" // 上层api塞入

	LanguageKey      = "Use-Language"
	AuthKeyToken     = "token"
	AuthKeyOwnKind   = "ownKind"
	AuthKeyOwnID     = "ownId"
	AuthKeyUserID    = "userId"
	AuthKeyAccountID = "accountId"
)

// IsContentTypeOk 检查请求的Content-Type是否符合预期
func IsContentTypeOk(c *gin.Context) bool {
	switch c.ContentType() {
	case binding.MIMEJSON,
		binding.MIMEXML,
		binding.MIMEPOSTForm,
		binding.MIMEMultipartPOSTForm:
		return true
	}
	return false
}

// AcceptContentTypes 获取支持的Content-Type列表
func AcceptContentTypes() []string {
	return []string{
		binding.MIMEJSON,
		binding.MIMEXML,
		binding.MIMEPOSTForm,
		binding.MIMEMultipartPOSTForm,
	}
}

// ResponseData 统一响应数据
func ResponseData(c *gin.Context, code int, obj any) {
	if c == nil {
		return
	}
	accept := c.GetHeader("Accept")
	// 做不做都一样
	//if accept == "" || strings.Contains(accept, "*/*") || strings.Contains(accept, "application/*") {
	//	accept = binding.MIMEJSON
	//} else if strings.Contains(accept, "msg/*") {
	//	accept = binding.MIMEXML
	//}

	switch {
	case strings.Contains(accept, binding.MIMEPROTOBUF):
		c.ProtoBuf(code, obj)
	case strings.Contains(accept, binding.MIMEXML), strings.Contains(accept, binding.MIMEXML2):
		c.XML(code, obj)
	case strings.Contains(accept, binding.MIMETOML):
		c.TOML(code, obj)
	case strings.Contains(accept, binding.MIMEYAML), strings.Contains(accept, binding.MIMEYAML2):
		c.YAML(code, obj)
	// 还没做xss
	//case strings.Contains(accept, binding.MIMEHTML):
	//	c.HTML(code, "", obj)
	default:
		c.JSON(code, obj) // 默认使用JSON格式
	}
}
