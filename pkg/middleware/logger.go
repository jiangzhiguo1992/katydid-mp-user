package middleware

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"katydid-mp-user/pkg/log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	RequestIDKey     = "RequestID"
	XRequestIDHeader = "X-Request-ID"
	green            = "\033[97;42m"
	white            = "\033[90;47m"
	yellow           = "\033[90;43m"
	red              = "\033[97;41m"
	blue             = "\033[97;44m"
	reset            = "\033[0m"
	maxBodyLogSize   = 4096 // 最大日志记录的请求体大小(字节)
)

var (
	// 缓存状态码和方法的颜色，避免重复计算
	statusColorCache = make(map[int]string)
	methodColorCache = make(map[string]string)
	colorCacheMutex  sync.RWMutex
)

// LoggerConfig 定义ZapLogger的配置选项
type LoggerConfig struct {
	SkipPaths      []string      // 需要跳过的路径
	SkipExtensions []string      // 需要跳过的文件扩展名
	TimeFormat     string        // 时间格式
	LogParams      bool          // 是否记录请求参数
	LogHeaders     bool          // 是否记录请求头
	LogBody        bool          // 是否记录请求体
	MaxBodySize    int           // 记录的最大请求体大小
	BodyTypes      []string      // 要记录的请求体内容类型
	HeaderFilter   []string      // 要记录的请求头字段(为空时记录所有)
	TraceIDFunc    func() string // 自定义跟踪ID生成函数
}

// DefaultLoggerConfig 返回默认配置
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		SkipPaths:      []string{"/favicon.ico", "/health", "/metrics"},
		SkipExtensions: []string{".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg"},
		TimeFormat:     "02/Jan/2006 - 15:04:05",
		LogParams:      true,
		LogHeaders:     true,
		LogBody:        true,
		MaxBodySize:    maxBodyLogSize,
		BodyTypes:      []string{"application/json", "application/xml", "application/x-www-form-urlencoded", "multipart/form-data"},
		HeaderFilter:   []string{"Content-Type", "User-Agent", "Referer", "Origin", "Authorization"},
		TraceIDFunc:    func() string { return uuid.New().String() },
	}
}

// statusColor 根据HTTP状态码返回对应的颜色（带缓存）
func statusColor(code int) string {
	colorCacheMutex.RLock()
	color, exists := statusColorCache[code]
	colorCacheMutex.RUnlock()

	if exists {
		return color
	}

	var newColor string
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		newColor = green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		newColor = white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		newColor = yellow
	default:
		newColor = red
	}

	colorCacheMutex.Lock()
	statusColorCache[code] = newColor
	colorCacheMutex.Unlock()
	return newColor
}

// methodColor 根据HTTP方法返回对应的颜色（带缓存）
func methodColor(method string) string {
	colorCacheMutex.RLock()
	color, exists := methodColorCache[method]
	colorCacheMutex.RUnlock()

	if exists {
		return color
	}

	var newColor string
	switch method {
	case http.MethodGet:
		newColor = blue
	case http.MethodPost:
		newColor = green
	case http.MethodPut:
		newColor = yellow
	case http.MethodDelete:
		newColor = red
	default:
		newColor = white
	}

	colorCacheMutex.Lock()
	methodColorCache[method] = newColor
	colorCacheMutex.Unlock()
	return newColor
}

// shouldSkipPath 检查是否应该跳过日志记录
func shouldSkipPath(path string, config LoggerConfig) bool {
	// 检查完全匹配的路径
	for _, p := range config.SkipPaths {
		if p == path {
			return true
		}
		// 支持路径前缀匹配，如 /api/*
		if strings.HasSuffix(p, "/*") && strings.HasPrefix(path, p[:len(p)-1]) {
			return true
		}
	}

	// 检查文件扩展名
	ext := filepath.Ext(path)
	if ext != "" {
		for _, skipExt := range config.SkipExtensions {
			if skipExt == ext {
				return true
			}
		}
	}

	return false
}

// bodyLogReader 是一个用于读取和重放请求体的结构
type bodyLogReader struct {
	body        []byte
	bodyReader  io.ReadCloser
	contentType string
	maxSize     int
}

// newBodyLogReader 创建一个新的bodyLogReader
func newBodyLogReader(body io.ReadCloser, contentType string, maxSize int) (*bodyLogReader, error) {
	if body == nil {
		return &bodyLogReader{body: []byte{}}, nil
	}

	data, err := io.ReadAll(io.LimitReader(body, int64(maxSize+1)))
	if err != nil {
		return nil, err
	}

	return &bodyLogReader{
		body:        data,
		contentType: contentType,
		maxSize:     maxSize,
	}, nil
}

// String 返回请求体的字符串表示
func (b *bodyLogReader) String() string {
	if len(b.body) == 0 {
		return ""
	}

	if len(b.body) > b.maxSize {
		return fmt.Sprintf("[请求体过大: %d bytes]", len(b.body))
	}

	// 只记录文本类型的请求体
	if strings.Contains(b.contentType, "json") ||
		strings.Contains(b.contentType, "xml") ||
		strings.Contains(b.contentType, "text") ||
		strings.Contains(b.contentType, "form") {
		return string(b.body)
	}

	return fmt.Sprintf("[非文本请求体: %s, %d bytes]", b.contentType, len(b.body))
}

// GetReader 返回一个新的读取器用于请求体
func (b *bodyLogReader) GetReader() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(b.body))
}

// ZapLogger 返回一个Gin中间件，使用默认配置记录HTTP请求的日志
func ZapLogger() gin.HandlerFunc {
	config := DefaultLoggerConfig()
	return ZapLoggerWithConfig(config)
}

// ZapLoggerWithConfig 返回一个使用自定义配置的Gin中间件
func ZapLoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要跳过日志记录
		path := c.Request.URL.Path
		if shouldSkipPath(path, config) {
			c.Next()
			return
		}

		// 获取或生成请求ID
		requestID := c.GetHeader(XRequestIDHeader)
		if requestID == "" {
			requestID = config.TraceIDFunc()
		}
		c.Set(RequestIDKey, requestID)
		c.Header(XRequestIDHeader, requestID)

		// 记录开始时间
		start := time.Now()
		//timeFormatted := start.Format(config.TimeFormat)

		// 准备路径和查询参数
		fullPath := path
		query := c.Request.URL.RawQuery
		if query != "" {
			fullPath = fullPath + "?" + query
		}

		// 收集请求参数
		var params map[string]any
		if config.LogParams {
			params = collectRequestParams(c)
		}

		// 收集请求头
		var headers map[string]string
		if config.LogHeaders {
			headers = collectRequestHeaders(c, config.HeaderFilter)
		}

		// 收集请求体
		var bodyLog *bodyLogReader
		var err error
		if config.LogBody && shouldLogRequestBody(c.Request.Header.Get("Content-Type"), config.BodyTypes) {
			bodyLog, err = newBodyLogReader(c.Request.Body, c.Request.Header.Get("Content-Type"), config.MaxBodySize)
			if err == nil && bodyLog != nil && len(bodyLog.body) > 0 {
				c.Request.Body = bodyLog.GetReader()
			}
		}

		// 处理请求
		c.Next()

		// 计算延迟及获取响应信息
		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 构建日志消息
		statusCode := fmt.Sprintf("%s %3d %s", statusColor(status), status, reset)
		methodFormatted := fmt.Sprintf("%s %-7s %s", methodColor(method), method, reset)
		msg := fmt.Sprintf("[GIN] %s | %s %s | %12v | %15s",
			//timeFormatted,
			statusCode,
			methodFormatted,
			fullPath,
			latency,
			clientIP,
		)

		// 构建结构化日志字段
		fields := []log.Field{
			log.FInt("status", status),
			log.FString("method", method),
			log.FString("path", path),
			log.FString("ip", clientIP),
			log.FString("request_id", requestID),
			log.FDuration("latency", latency),
			log.FInt("size", size),
		}

		// 添加用户代理
		if ua := c.Request.UserAgent(); ua != "" {
			fields = append(fields, log.FString("user-agent", ua))
		}

		// 添加请求参数
		if config.LogParams && len(params) > 0 {
			for k, v := range params {
				fields = append(fields, log.FString("param_"+k, fmt.Sprintf("%v", v)))
			}
		}

		// 添加请求头
		if config.LogHeaders && len(headers) > 0 {
			for k, v := range headers {
				fields = append(fields, log.FString("header_"+k, v))
			}
		}

		// 添加请求体
		if config.LogBody && bodyLog != nil {
			bodyStr := bodyLog.String()
			if bodyStr != "" {
				fields = append(fields, log.FString("request_body", bodyStr))
			}
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				fields = append(fields, log.FString("error", e))
			}
		}

		// 根据状态码选择日志级别
		if len(c.Errors) > 0 {
			log.Error(msg, fields...)
		} else if status >= http.StatusInternalServerError {
			log.Error(msg, fields...)
		} else if status >= http.StatusBadRequest {
			log.Warn(msg, fields...)
		} else {
			log.Info(msg, fields...)
		}
	}
}

// collectRequestParams 收集请求参数
func collectRequestParams(c *gin.Context) map[string]any {
	params := make(map[string]any)

	// URL查询参数
	for k, v := range c.Request.URL.Query() {
		if len(v) > 1 {
			params[k] = v
		} else if len(v) == 1 {
			params[k] = v[0]
		}
	}

	// POST表单参数
	if c.Request.Method != http.MethodGet {
		// 对于POST/PUT请求，尝试解析表单
		_ = c.Request.ParseForm()
		for k, v := range c.Request.PostForm {
			if len(v) > 1 {
				params[k] = v
			} else if len(v) == 1 {
				params[k] = v[0]
			}
		}

		// 路径参数
		for _, param := range c.Params {
			params[param.Key] = param.Value
		}
	}

	return params
}

// collectRequestHeaders 收集请求头
func collectRequestHeaders(c *gin.Context, filter []string) map[string]string {
	headers := make(map[string]string)

	if len(filter) == 0 {
		// 记录所有请求头
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				// 敏感信息处理（例如Authorization）
				if k == "Authorization" || k == "Cookie" {
					headers[k] = "[REDACTED]"
				} else {
					headers[k] = strings.Join(v, ", ")
				}
			}
		}
	} else {
		// 只记录过滤后的请求头
		for _, k := range filter {
			if v := c.Request.Header.Get(k); v != "" {
				if k == "Authorization" || k == "Cookie" {
					headers[k] = "[REDACTED]"
				} else {
					headers[k] = v
				}
			}
		}
	}

	return headers
}

// shouldLogRequestBody 检查是否应该记录请求体
func shouldLogRequestBody(contentType string, allowedTypes []string) bool {
	if contentType == "" {
		return false
	}

	contentType = strings.ToLower(contentType)
	for _, allowed := range allowedTypes {
		if strings.Contains(contentType, strings.ToLower(allowed)) {
			return true
		}
	}

	return false
}
