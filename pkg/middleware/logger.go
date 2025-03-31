package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/pkg/str"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	logGreen  = "\033[97;42m"
	logWhite  = "\033[90;47m"
	logYellow = "\033[90;43m"
	logRed    = "\033[97;41m"
	logBlue   = "\033[97;44m"
	logReset  = "\033[0m"
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

	// 添加正则表达式缓存
	logRegexpCache      = make(map[string]*regexp.Regexp)
	logRegexpCacheMutex sync.RWMutex
)

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

// LoggerConfig 定义ZapLogger的配置选项
type LoggerConfig struct {
	// trace
	ServiceName     string        // 当前服务名称
	TraceIDHeader   string        // 自定义跟踪ID头部字段名
	TracePathHeader string        // 需要跟踪的路径头部字段名
	TraceIDFunc     func() string `json:"-"` // 自定义跟踪ID生成函数(一般是网关生成)
	// info
	LogParams   bool // 是否记录请求参数
	LogHeaders  bool // 是否记录请求头
	LogBody     bool // 是否记录请求体
	LogResponse bool // 是否记录响应体
	MaxBodySize int  // 记录的最大请求/返回体大小
	// skip
	SkipStatus     []int    // 需要跳过的状态码
	SkipPaths      []string // 需要跳过的路径
	SkipExtensions []string // 需要跳过的文件扩展名
	Sensitives     []string // 敏感信息关键词列表
	HeaderFilter   []string // 要记录的请求头字段(为空时记录所有)
	HeaderSkip     []string // 要跳过的请求头字段
}

// LoggerDefaultConfig 返回默认配置
func LoggerDefaultConfig(serviceName string, status []int, skips, sensitives []string, size int) LoggerConfig {
	return LoggerConfig{
		// trace
		ServiceName:     serviceName,
		TraceIDHeader:   XRequestIDHeader,
		TracePathHeader: XRequestPathHeader,
		TraceIDFunc:     func() string { return uuid.New().String() },
		// info
		LogParams:   true,
		LogHeaders:  true,
		LogBody:     true,
		LogResponse: true,
		MaxBodySize: size,
		// skip
		SkipStatus:     status,
		SkipPaths:      skips, //[]string{"/favicon.ico", "/health", "/metrics"},
		SkipExtensions: []string{".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg"},
		Sensitives:     sensitives, //[]string{"password", "token", "secret", "Authorization", "Cookie"},
		HeaderFilter:   []string{}, //[]string{"Content-Type", "User-Agent", "Referer", "Origin", "Authorization"},
		HeaderSkip:     []string{}, //[]string{"Content-Length", "Host", "Accept-Encoding", "Connection", "Upgrade-Insecure-Requests"},
	}
}

// ZapLoggerWithConfig 返回一个使用自定义配置的Gin中间件
func ZapLoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	//marshal, _ := json.MarshalIndent(config, "", "\t")
	//log.InfoMustf(gin.Mode() != gin.DebugMode, "■ ■ Log(中间件) ■ ■ 配置 ---> %s", marshal)

	return func(c *gin.Context) {
		// 检查是否需要跳过日志记录
		path := c.Request.URL.Path
		if loggerShouldSkipPath(path, config) {
			c.Next()
			return
		}

		// 添加或提取请求ID
		traceID := logExtractOrGenerateRequestID(c, config)
		tracePath := logExtractOrGenerateRequestPaths(c, config)

		// 记录开始时间
		start := time.Now()

		// 准备路径和查询参数
		fullPath := logBuildFullPath(c)

		// 收集请求参数
		var params map[string]any
		if config.LogParams {
			params = logCollectRequestParams(c)
		}

		// 收集请求头
		var headers map[string]string
		if config.LogHeaders {
			headers = logCollectRequestHeaders(c, config)
		}

		// 收集请求体
		var bodyLog *loggerBodyReader
		if config.LogBody {
			bodyLog, _ = newLoggerBodyReader(c.Request.Body, c.Request.Header.Get("Content-Type"), config.MaxBodySize)
			if bodyLog != nil && len(bodyLog.body) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewReader(bodyLog.body))
			}
		}

		// 设置响应体记录器
		var responseBodyBuffer *bytes.Buffer
		if config.LogResponse {
			responseBodyBuffer = &bytes.Buffer{}
			c.Writer = &loggerResponseBodyWriter{
				ResponseWriter: c.Writer,
				body:           responseBodyBuffer,
				maxLen:         config.MaxBodySize,
			}
		}

		// 处理请求
		c.Next()

		// 检查是否需要跳过状态码
		status := c.Writer.Status()
		if len(config.SkipStatus) > 0 {
			for _, skip := range config.SkipStatus {
				if status == skip {
					return
				}
			}
		}

		// 记录日志
		logHTTPRequest(c, config, start, traceID, tracePath, fullPath, headers, bodyLog, responseBodyBuffer)

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

// loggerShouldSkipPath 检查是否应该跳过日志记录
func loggerShouldSkipPath(path string, config LoggerConfig) bool {
	// 检查完全匹配的路径
	for _, pattern := range config.SkipPaths {
		if str.MatchURLPath(path, pattern) {
			return true
		}
	}

	// 检查文件扩展名
	if ext := filepath.Ext(path); ext != "" {
		for _, skipExt := range config.SkipExtensions {
			if skipExt == ext {
				return true
			}
		}
	}
	return false
}

// 提取或生成请求ID
func logExtractOrGenerateRequestID(c *gin.Context, config LoggerConfig) string {
	requestID := c.GetHeader(config.TraceIDHeader)
	if requestID == "" {
		requestID = config.TraceIDFunc()
		c.Header(config.TraceIDHeader, requestID)
	}
	c.Set(XRequestIDHeader, requestID)
	return requestID
}

// 提取或生成请求路径
func logExtractOrGenerateRequestPaths(c *gin.Context, config LoggerConfig) string {
	requestPath := c.GetHeader(config.TracePathHeader)
	if requestPath == "" {
		c.Header(config.TracePathHeader, config.ServiceName)
	}
	c.Set(XRequestPathHeader, requestPath+">"+config.ServiceName)
	return c.GetString(XRequestPathHeader)
}

// 构建完整路径（含查询参数）
func logBuildFullPath(c *gin.Context) string {
	fullPath := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	if query != "" {
		fullPath = fullPath + "?" + query
	}
	return fullPath
}

// logCollectRequestParams 收集请求参数
func logCollectRequestParams(c *gin.Context) map[string]any {
	params := loggerParamsPool.Get().(map[string]any)
	// 确保map是空的
	for k := range params {
		delete(params, k)
	}

	for k, v := range c.Request.URL.Query() {
		if len(v) > 1 {
			params[k] = v
		} else if len(v) == 1 {
			params[k] = v[0]
		}
	}
	return params
}

// 优化的收集请求头函数
func logCollectRequestHeaders(c *gin.Context, config LoggerConfig) map[string]string {
	headers := loggerHeadersPool.Get().(map[string]string)

	// 确保map是空的
	for k := range headers {
		delete(headers, k)
	}

	if len(config.HeaderFilter) == 0 {
		// 记录所有请求头
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				// 敏感信息处理
				if logContainsSensitiveWord(k, config.Sensitives) {
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
				if logContainsSensitiveWord(k, config.Sensitives) {
					headers[k] = "[REDACTED]"
				} else {
					headers[k] = v
				}
			}
		}
	}

	// 处理跳过的请求头
	for _, k := range config.HeaderSkip {
		delete(headers, k)
	}

	return headers
}

// 检查是否包含敏感词
func logContainsSensitiveWord(key string, sensitiveWords []string) bool {
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
	start time.Time, traceID, tracePath string,
	fullPath string, headers map[string]string,
	bodyLog *loggerBodyReader, responseBody *bytes.Buffer,
) {
	// 计算延迟及获取响应信息
	latency := time.Since(start)
	status := c.Writer.Status()
	size := c.Writer.Size()
	clientIP := c.ClientIP()
	method := c.Request.Method

	isDebug := gin.Mode() == gin.DebugMode
	statusCode := strconv.Itoa(status)
	if isDebug {
		statusCode = fmt.Sprintf("%s %3d %s", loggerStatusColor(status), status, logReset)
	}
	methodFormatted := method
	if isDebug {
		methodFormatted = fmt.Sprintf("%s %-7s %s", loggerMethodColor(method), method, logReset)
	}

	// 构建日志消息
	msg := fmt.Sprintf(
		"[GIN] %s | %12v | %15s | %4s %s \n\ttraceID: %s, tracePath: %s \n\theader: %v",
		statusCode, latency, clientIP,
		methodFormatted, fullPath,
		traceID, tracePath, headers,
	)
	if bodyLog != nil {
		bodyString := bodyLog.formatBodyString(config.Sensitives...)
		msg = msg + fmt.Sprintf("\n\trequest_body: %s", bodyString)
	}
	if responseBody != nil {
		bodyString := responseBody.String()
		if len(bodyString) > config.MaxBodySize {
			bodyString = bodyString[:config.MaxBodySize] + "... [truncated]"
		}
		msg = msg + fmt.Sprintf("\n\tresponse_body(%d): %s", size, bodyString)
	}

	// 根据状态码选择日志级别
	logByStatus(msg, status, c.Errors)
}

// logByStatus logs the message with appropriate log level based on status code
func logByStatus(msg string, status int, errors []*gin.Error, fields ...log.Field) {
	switch {
	case len(errors) > 0:
		errMsgs := make([]string, len(errors))
		for i, err := range errors {
			errMsgs[i] = err.Error()
		}
		fields = append(fields, log.FStrings("errors", errMsgs))
		log.ErrorMust(gin.Mode() != gin.DebugMode, msg, fields...)
	case status >= http.StatusInternalServerError:
		log.ErrorMust(gin.Mode() != gin.DebugMode, msg, fields...)
	case status >= http.StatusBadRequest:
		log.WarnMust(gin.Mode() != gin.DebugMode, msg, fields...)
	default:
		log.Debug(msg, fields...)
	}
}

// loggerBodyReader 是一个用于读取和重放请求体的结构
type loggerBodyReader struct {
	body        []byte
	contentType string
	maxSize     int
}

// newLoggerBodyReader 创建一个新的bodyLogReader
func newLoggerBodyReader(body io.ReadCloser, contentType string, maxSize int) (*loggerBodyReader, error) {
	if body == nil {
		return &loggerBodyReader{body: []byte{}}, nil
	}

	data, err := io.ReadAll(io.LimitReader(body, int64(maxSize+1)))
	if err != nil {
		return nil, err
	}

	return &loggerBodyReader{
		body:        data,
		contentType: contentType,
		maxSize:     maxSize,
	}, nil
}

// GetReader 返回一个新的读取器用于请求体
func (b *loggerBodyReader) GetReader() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(b.body))
}

// formatBodyString formats the request body for logging, handling sensitive data
func (b *loggerBodyReader) formatBodyString(sensitiveFields ...string) string {
	if b == nil || len(b.body) == 0 {
		return ""
	}

	bodyStr := b.String()
	if len(bodyStr) > b.maxSize {
		bodyStr = bodyStr[:b.maxSize] + "... [truncated]"
	}

	// 使用正则表达式进行更精确的敏感信息替换
	for _, field := range sensitiveFields {
		// JSON 格式敏感字段
		pattern := fmt.Sprintf(`("%s"\s*:\s*)(\"[^\"]*\"|\d+|true|false|null)`, regexp.QuoteMeta(field))
		re, err := logGetCachedRegexp(pattern)
		if err == nil {
			bodyStr = re.ReplaceAllString(bodyStr, `$1"[REDACTED]"`)
		}

		// 表单格式敏感字段
		formPattern := fmt.Sprintf(`(%s=)([^&\s]*)`, regexp.QuoteMeta(field))
		formRe, err := logGetCachedRegexp(formPattern)
		if err == nil {
			bodyStr = formRe.ReplaceAllString(bodyStr, `$1[REDACTED]`)
		}
	}

	return strings.Replace(bodyStr, "\n", "\n\t", -1)
}

// String 返回请求体的字符串表示
func (b *loggerBodyReader) String() string {
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

// 获取缓存的正则表达式
func logGetCachedRegexp(pattern string) (*regexp.Regexp, error) {
	logRegexpCacheMutex.RLock()
	re, exists := logRegexpCache[pattern]
	logRegexpCacheMutex.RUnlock()

	if exists {
		return re, nil
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	logRegexpCacheMutex.Lock()
	logRegexpCache[pattern] = compiled
	logRegexpCacheMutex.Unlock()

	return compiled, nil
}

// 优化的响应体记录器
type loggerResponseBodyWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	maxLen int
}

func (w *loggerResponseBodyWriter) Write(b []byte) (int, error) {
	if w.body != nil && w.body.Len() < w.maxLen {
		// 只记录到最大长度
		remaining := w.maxLen - w.body.Len()
		if remaining > len(b) {
			w.body.Write(b)
		} else if remaining > 0 {
			w.body.Write(b[:remaining])
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *loggerResponseBodyWriter) WriteString(s string) (int, error) {
	if w.body != nil && w.body.Len() < w.maxLen {
		// 只记录到最大长度
		remaining := w.maxLen - w.body.Len()
		if remaining > len(s) {
			w.body.WriteString(s)
		} else if remaining > 0 {
			w.body.WriteString(s[:remaining])
		}
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *loggerResponseBodyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.ResponseWriter == nil {
		return nil, nil, fmt.Errorf("ResponseWriter is nil")
	}

	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}

func (w *loggerResponseBodyWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *loggerResponseBodyWriter) CloseNotify() <-chan bool {
	if notifier, ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return notifier.CloseNotify()
	}
	return nil
}

func (w *loggerResponseBodyWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
