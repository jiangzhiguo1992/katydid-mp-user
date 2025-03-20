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
	logGreen       = "\033[97;42m"
	logWhite       = "\033[90;47m"
	logYellow      = "\033[90;43m"
	logRed         = "\033[97;41m"
	logBlue        = "\033[97;44m"
	logReset       = "\033[0m"
	maxBodyLogSize = 4096 // 最大日志记录的请求体大小(字节)
)

var (
	// 缓存状态码和方法的颜色，避免重复计算
	loggerStatusColorCache = make(map[int]string)
	loggerMethodColorCache = make(map[string]string)
	loggerColorCacheMutex  sync.RWMutex

	// 使用对象池减少内存分配
	loggerParamsPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]any, 8)
		},
	}

	loggerHeadersPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]string, 8)
		},
	}
)

// LoggerConfig 定义ZapLogger的配置选项
type LoggerConfig struct {
	SkipPaths      []string      // 需要跳过的路径
	SkipExtensions []string      // 需要跳过的文件扩展名
	TimeFormat     string        // 时间格式
	LogParams      bool          // 是否记录请求参数
	LogHeaders     bool          // 是否记录请求头
	LogBody        bool          // 是否记录请求体
	LogResponse    bool          // 是否记录响应体
	MaxBodySize    int           // 记录的最大请求体大小
	BodyTypes      []string      // 要记录的请求体内容类型(为空时记录所有)
	HeaderFilter   []string      // 要记录的请求头字段(为空时记录所有)
	TraceIDFunc    func() string // 自定义跟踪ID生成函数
	TraceIDHeader  string        // 自定义跟踪ID头部字段名
	SensitiveWords []string      // 敏感信息关键词列表
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
		LogResponse:    false,
		MaxBodySize:    maxBodyLogSize,
		BodyTypes:      []string{}, //[]string{"application/json", "application/xml", "application/x-www-form-urlencoded", "multipart/form-data"},
		HeaderFilter:   []string{}, //[]string{"Content-Type", "User-Agent", "Referer", "Origin", "Authorization"},
		TraceIDFunc:    func() string { return uuid.New().String() },
		TraceIDHeader:  XRequestIDHeader,
		SensitiveWords: []string{"password", "token", "secret", "Authorization", "Cookie"},
	}
}

// loggerStatusColor 根据HTTP状态码返回对应的颜色（带缓存）
func loggerStatusColor(code int) string {
	loggerColorCacheMutex.RLock()
	color, exists := loggerStatusColorCache[code]
	loggerColorCacheMutex.RUnlock()

	if exists {
		return color
	}

	var newColor string
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		newColor = logGreen
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		newColor = logWhite
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		newColor = logYellow
	default:
		newColor = logRed
	}

	loggerColorCacheMutex.Lock()
	loggerStatusColorCache[code] = newColor
	loggerColorCacheMutex.Unlock()
	return newColor
}

// loggerMethodColor 根据HTTP方法返回对应的颜色（带缓存）
func loggerMethodColor(method string) string {
	loggerColorCacheMutex.RLock()
	color, exists := loggerMethodColorCache[method]
	loggerColorCacheMutex.RUnlock()

	if exists {
		return color
	}

	var newColor string
	switch method {
	case http.MethodGet:
		newColor = logBlue
	case http.MethodPost:
		newColor = logGreen
	case http.MethodPut:
		newColor = logYellow
	case http.MethodDelete:
		newColor = logRed
	default:
		newColor = logWhite
	}

	loggerColorCacheMutex.Lock()
	loggerMethodColorCache[method] = newColor
	loggerColorCacheMutex.Unlock()
	return newColor
}

// bodyLogReader 是一个用于读取和重放请求体的结构
type bodyLogReader struct {
	body        []byte
	contentType string
	maxSize     int
}

// 优化的响应体记录器
type responseBodyWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	maxLen int
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	if r.body != nil && r.body.Len() < r.maxLen {
		r.body.Write(b)
	}
	return r.ResponseWriter.Write(b)
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

		// 添加或提取请求ID
		requestID := extractOrGenerateRequestID(c, config)

		// 记录开始时间
		start := time.Now()

		// 准备路径和查询参数
		fullPath := buildFullPath(c)

		// 收集请求参数
		var params map[string]any
		if config.LogParams {
			params = collectRequestParams(c)
		}

		// 收集请求头
		var headers map[string]string
		if config.LogHeaders {
			headers = collectRequestHeaders(c, config)
		}

		// 收集请求体
		var bodyLog *bodyLogReader
		if config.LogBody && shouldLogRequestBody(c.Request.Header.Get("Content-Type"), config.BodyTypes) {
			bodyLog, _ = newBodyLogReader(c.Request.Body, c.Request.Header.Get("Content-Type"), config.MaxBodySize)
			if bodyLog != nil && len(bodyLog.body) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewReader(bodyLog.body))
			}
		}

		// 设置响应体记录器
		var responseBodyBuffer *bytes.Buffer
		if config.LogResponse {
			responseBodyBuffer = &bytes.Buffer{}
			c.Writer = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:           responseBodyBuffer,
				maxLen:         config.MaxBodySize,
			}
		}

		// 处理请求
		c.Next()

		// 记录日志
		logHTTPRequest(c, config, start, requestID, fullPath, params, headers, bodyLog, responseBodyBuffer)

		// 回收对象池资源
		if config.LogParams && params != nil {
			for k := range params {
				delete(params, k)
			}
			loggerParamsPool.Put(params)
		}
		if config.LogHeaders && headers != nil {
			for k := range headers {
				delete(headers, k)
			}
			loggerHeadersPool.Put(headers)
		}
	}
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

// 提取或生成请求ID
func extractOrGenerateRequestID(c *gin.Context, config LoggerConfig) string {
	requestID := c.GetHeader(config.TraceIDHeader)
	if requestID == "" {
		requestID = config.TraceIDFunc()
		c.Set(RequestIDKey, requestID)
		c.Header(config.TraceIDHeader, requestID)
	} else {
		c.Set(RequestIDKey, requestID)
	}
	return requestID
}

// 构建完整路径（含查询参数）
func buildFullPath(c *gin.Context) string {
	fullPath := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	if query != "" {
		fullPath = fullPath + "?" + query
	}
	return fullPath
}

// collectRequestParams 收集请求参数
func collectRequestParams(c *gin.Context) map[string]any {
	params := loggerParamsPool.Get().(map[string]any)

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
		_ = c.Request.ParseMultipartForm(32 << 20)
		if c.Request.Form != nil {
			for k, v := range c.Request.Form {
				if len(v) > 1 {
					params[k] = v
				} else if len(v) == 1 {
					params[k] = v[0]
				}
			}
		}

		// 路径参数
		for _, param := range c.Params {
			params[param.Key] = param.Value
		}
	}
	return params
}

// 优化的收集请求头函数
func collectRequestHeaders(c *gin.Context, config LoggerConfig) map[string]string {
	headers := loggerHeadersPool.Get().(map[string]string)

	if len(config.HeaderFilter) == 0 {
		// 记录所有请求头
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				// 敏感信息处理
				if containsSensitiveWord(k, config.SensitiveWords) {
					headers[k] = "[REDACTED]"
				} else {
					headers[k] = strings.Join(v, ", ")
				}
			}
		}
	} else {
		// 只记录过滤后的请求头
		for _, k := range config.HeaderFilter {
			if v := c.Request.Header.Get(k); v != "" {
				if containsSensitiveWord(k, config.SensitiveWords) {
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
	} else if len(allowedTypes) <= 0 {
		return true
	}

	contentType = strings.ToLower(contentType)
	for _, allowed := range allowedTypes {
		if strings.Contains(contentType, strings.ToLower(allowed)) {
			return true
		}
	}
	return false
}

// 检查是否包含敏感词
func containsSensitiveWord(key string, sensitiveWords []string) bool {
	lowKey := strings.ToLower(key)
	for _, word := range sensitiveWords {
		if strings.Contains(lowKey, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// 记录HTTP请求日志（抽取复杂逻辑）
func logHTTPRequest(
	c *gin.Context, config LoggerConfig,
	start time.Time, requestID, fullPath string,
	params map[string]any, headers map[string]string,
	bodyLog *bodyLogReader, responseBody *bytes.Buffer,
) {
	// 计算延迟及获取响应信息
	latency := time.Since(start)
	status := c.Writer.Status()
	size := c.Writer.Size()
	clientIP := c.ClientIP()
	method := c.Request.Method

	// 构建日志消息
	statusCode := fmt.Sprintf("%s %3d %s", loggerStatusColor(status), status, logReset)
	methodFormatted := fmt.Sprintf("%s %-7s %s", loggerMethodColor(method), method, logReset)
	msg := fmt.Sprintf("[GIN] %s | %s %s | %12v | %15s",
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
		log.FString("path", c.Request.URL.Path),
		log.FString("ip", clientIP),
		log.FString("request-id", requestID),
		log.FDuration("latency", latency),
		log.FInt("size", size),
	}

	// 添加用户代理
	if ua := c.Request.UserAgent(); ua != "" {
		fields = append(fields, log.FString("user-agent", ua))
	}

	// 添加额外字段
	addParamsToFields(fields, params, config)
	addHeadersToFields(fields, headers, config)
	addBodyToFields(fields, bodyLog, config)
	addResponseToFields(fields, responseBody, config)
	addErrorsToFields(fields, c)

	// 根据状态码选择日志级别
	logByStatus(msg, status, c.Errors, fields...)
}

// 添加参数到日志字段
func addParamsToFields(fields []log.Field, params map[string]any, config LoggerConfig) []log.Field {
	if config.LogParams && len(params) > 0 {
		for k, v := range params {
			fields = append(fields, log.FString("param_"+k, fmt.Sprintf("%v", v)))
		}
	}
	return fields
}

// 添加头部到日志字段
func addHeadersToFields(fields []log.Field, headers map[string]string, config LoggerConfig) []log.Field {
	if config.LogHeaders && len(headers) > 0 {
		for k, v := range headers {
			fields = append(fields, log.FString("header_"+k, v))
		}
	}
	return fields
}

// 添加请求体到日志字段
func addBodyToFields(fields []log.Field, bodyLog *bodyLogReader, config LoggerConfig) []log.Field {
	if config.LogBody && bodyLog != nil {
		bodyStr := formatBodyString(bodyLog)
		if bodyStr != "" {
			fields = append(fields, log.FString("request_body", bodyStr))
		}
	}
	return fields
}

// 添加响应体到日志字段
func addResponseToFields(fields []log.Field, responseBody *bytes.Buffer, config LoggerConfig) []log.Field {
	if config.LogResponse && responseBody != nil && responseBody.Len() > 0 {
		respStr := responseBody.String()
		if len(respStr) > config.MaxBodySize {
			respStr = respStr[:config.MaxBodySize] + "... [truncated]"
		}
		fields = append(fields, log.FString("response_body", respStr))
	}
	return fields
}

// addErrorsToFields adds any errors from the context to the log fields
func addErrorsToFields(fields []log.Field, c *gin.Context) []log.Field {
	if len(c.Errors) > 0 {
		errMsgs := make([]string, len(c.Errors))
		for i, err := range c.Errors {
			errMsgs[i] = err.Error()
		}
		fields = append(fields, log.FStrings("errors", errMsgs))
	}
	return fields
}

// logByStatus logs the message with appropriate log level based on status code
func logByStatus(msg string, status int, errors []*gin.Error, fields ...log.Field) {
	switch {
	case len(errors) > 0:
		log.Error(msg, fields...)
	case status >= http.StatusInternalServerError:
		log.Error(msg, fields...)
	case status >= http.StatusBadRequest:
		log.Warn(msg, fields...)
	default:
		log.Info(msg, fields...)
	}
}

// formatBodyString formats the request body for logging, handling sensitive data
func formatBodyString(bodyLog *bodyLogReader) string {
	if bodyLog == nil {
		return ""
	}

	bodyStr := bodyLog.String()
	if bodyStr == "" {
		return ""
	}

	// Very simplified sensitive data redaction
	// In a real implementation, you might want to use regex to mask sensitive fields
	sensitiveFields := []string{"password", "token", "secret", "authorization", "cookie"}
	lowerBody := strings.ToLower(bodyStr)

	for _, field := range sensitiveFields {
		if strings.Contains(lowerBody, field) {
			// Very basic masking - in production you'd want more sophisticated regex-based replacement
			indexStart := strings.Index(lowerBody, field)
			if indexStart >= 0 {
				// Find the value portion and mask it
				// This is a simplified approach - real implementation would be more robust
				valueStart := strings.IndexAny(bodyStr[indexStart+len(field):], ":\"'") + indexStart + len(field) + 1
				valueEnd := strings.IndexAny(bodyStr[valueStart:], ",}\"'")
				if valueEnd > 0 {
					valueEnd += valueStart
					// Replace the sensitive value with [REDACTED]
					bodyStr = bodyStr[:valueStart] + "[REDACTED]" + bodyStr[valueEnd:]
					lowerBody = strings.ToLower(bodyStr) // Update lowercase version for next iterations
				}
			}
		}
	}
	return bodyStr
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
