package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"katydid-mp-user/pkg/log"
	"net/http"
	"time"
)

const (
	RequestIDKey = "RequestID"
)

// ZapLogger 返回一个 Gin 中间件，用于记录HTTP请求的日志
func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 构建与Gin默认日志类似的消息格式
		// [GIN] 2023/08/07 - 12:34:56 | 200 |     1.23ms |  127.0.0.1 | GET      "/path"
		msg := fmt.Sprintf("%3d | %s %s | %12v | %s",
			status,
			method,
			path, // fmt.Sprintf("\x1b[21m%s\x1b[0m", path),
			latency,
			clientIP,
		)

		// 准备额外日志字段(不会显示在主消息中，但会包含在结构化日志中)
		fields := []log.Field{
			log.FInt("status", status),
			log.FString("method", method),
			log.FString("path", path),
			log.FString("ip", clientIP),
			log.FString("request_id", requestID),
			log.FString("user-agent", c.Request.UserAgent()),
			log.FDuration("latency", latency),
			log.FInt("size", size),
		}

		if len(c.Errors) > 0 {
			// 添加错误信息到字段
			for _, e := range c.Errors.Errors() {
				fields = append(fields, log.FString("error", e))
			}
			log.Error(msg, fields...)
		} else {
			// 根据状态码选择日志级别
			if status >= http.StatusInternalServerError {
				log.Error(msg, fields...)
			} else if status >= http.StatusBadRequest {
				log.Warn(msg, fields...)
			} else {
				log.Info(msg, fields...)
			}
		}
	}
}
