package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"katydid-mp-user/pkg/log"
)

// Trace 分布式追踪相关
func Trace(serviceName, keyID, keyPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// id
		requestID := c.GetHeader(keyID)
		if requestID == "" {
			requestID = uuid.New().String()
			if gin.Mode() != gin.DebugMode {
				log.Warnf("■ ■ Trace(中间件) ■ ■ ID缺失，自动生成: %s", requestID)
			}
			c.Header(keyID, requestID) // header
		}
		c.Set(keyID, requestID) // context

		// path
		requestPath := c.GetHeader(keyPath)
		requestPath2 := requestPath + ">" + serviceName
		if requestPath == "" {
			c.Header(keyPath, requestPath2) // header
		}
		c.Set(keyPath, requestPath2) // context

		log.Debugf("■ ■ Trace(中间件) ■ ■ ID: %s, Path: %s", requestID, requestPath2)
	}
}
