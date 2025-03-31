package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Trace 分布式追踪相关
func Trace(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// id
		requestID := c.GetHeader(XRequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header(XRequestIDHeader, requestID) // header
		}
		c.Set(XRequestIDHeader, requestID) // context

		// path
		requestPath := c.GetHeader(XRequestPathHeader)
		requestPath2 := requestPath + ">" + serviceName
		if requestPath == "" {
			c.Header(XRequestPathHeader, requestPath2) // header
		}
		c.Set(XRequestPathHeader, requestPath2) // context
	}
}
