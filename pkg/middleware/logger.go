package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
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
	loggerBufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	// 预编译正则表达式模板，将字段名作为参数
	logJsonFieldRegexTpl = `("%s"\s*:\s*)(\"[^\"]*\"|\d+|true|false|null)`
	logFormFieldRegexTpl = `(%s=)([^&\s]*)`

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
	LogParams      bool     // 是否记录请求参数
	LogHeaders     bool     // 是否记录请求头
	LogBody        bool     // 是否记录请求体
	LogResponse    bool     // 是否记录响应体
	MaxBodySize    int      // 记录的最大请求/返回体大小
	SkipStatus     []int    // 需要跳过的状态码
	SkipPaths      []string // 需要跳过的路径
	SkipExtensions []string // 需要跳过的文件扩展名
	Sensitives     []string // 敏感信息关键词列表
	HeaderFilter   []string // 要记录的请求头字段(为空时记录所有)
	HeaderSkip     []string // 要跳过的请求头字段
}

// LoggerDefaultConfig 返回默认配置
func LoggerDefaultConfig(skipStatus []int, skipPaths, sensitives []string, size int) LoggerConfig {
	return LoggerConfig{
		LogParams:      true,
		LogHeaders:     true,
		LogBody:        true,
		LogResponse:    true,
		MaxBodySize:    size,
		SkipStatus:     skipStatus,
		SkipPaths:      skipPaths, //[]string{"/favicon.ico", "/health", "/metrics"},
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

		// 记录开始时间
		start := time.Now()

		// 准备路径和查询参数
		fullPath := logBuildFullPath(c)

		// 收集请求参数
		var params map[string]any
		if config.LogParams {
			params = logCollectRequestParams(c)
			defer func() {
				if params != nil {
					for k := range params {
						delete(params, k)
					}
					loggerParamsPool.Put(params)
				}
			}()
		}

		// 收集请求头
		var headers map[string]string
		if config.LogHeaders {
			headers = logCollectRequestHeaders(c, config)
			defer func() {
				if headers != nil {
					for k := range headers {
						delete(headers, k)
					}
					loggerHeadersPool.Put(headers)
				}
			}()
		}

		// 收集请求体
		var bodyLog *loggerBodyReader
		if config.LogBody {
			originalBody := c.Request.Body
			bodyLog, _ = newLoggerBodyReader(c.Request.Body, c.Request.Header.Get("Content-Type"), config.MaxBodySize)
			if bodyLog != nil && len(bodyLog.body) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewReader(bodyLog.body))
			}
			_ = originalBody.Close() // 添加这行，关闭原始请求体
		}

		// 设置响应体记录器
		var responseBodyBuffer *bytes.Buffer
		if config.LogResponse {
			responseBodyBuffer = loggerBufferPool.Get().(*bytes.Buffer)
			responseBodyBuffer.Reset()
			defer func() {
				responseBodyBuffer.Reset() // 清空缓冲区
				loggerBufferPool.Put(responseBodyBuffer)
			}()

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
		logHTTPRequest(c, config, start, fullPath, headers, bodyLog, responseBodyBuffer)
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

	if c.Request == nil || c.Request.URL == nil {
		return params
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
	start time.Time, fullPath string, headers map[string]string,
	bodyLog *loggerBodyReader, responseBody *bytes.Buffer,
) {
	// 计算延迟及获取响应信息
	latency := time.Since(start)
	status := c.Writer.Status()
	size := c.Writer.Size()
	clientIP := c.ClientIP()
	method := c.Request.Method

	isDebug := gin.Mode() == gin.DebugMode
	var statusCode string
	if isDebug {
		statusCode = fmt.Sprintf("%s %3d %s", loggerStatusColor(status), status, logReset)
	} else {
		statusCode = strconv.Itoa(status)
	}

	var methodFormatted string
	if isDebug {
		methodFormatted = fmt.Sprintf("%s %-7s %s", loggerMethodColor(method), method, logReset)
	} else {
		methodFormatted = method
	}

	// 构建日志消息
	var msgBuilder strings.Builder
	msgBuilder.Grow(512) // 预分配空间

	_, _ = fmt.Fprintf(&msgBuilder,
		"[GIN] %s | %12v | %15s | %4s %s \n\theader: %v \n\tcontext: %v",
		statusCode, latency, clientIP,
		methodFormatted, fullPath,
		headers, c.Keys,
	)
	if bodyLog != nil {
		if s := bodyLog.formatBodyString(config.Sensitives...); s != "" {
			_, _ = fmt.Fprintf(&msgBuilder, "\n\trequest_body: %s", s)
		}
	}

	if responseBody != nil && responseBody.Len() > 0 {
		if w, ok := c.Writer.(*loggerResponseBodyWriter); ok {
			w.mu.Lock()
			bodyString := responseBody.String()
			w.mu.Unlock()

			if config.MaxBodySize >= 0 && len(bodyString) > config.MaxBodySize {
				bodyString = bodyString[:config.MaxBodySize] + "... [truncated]"
			}
			_, _ = fmt.Fprintf(&msgBuilder, "\n\tresponse_body(%d): %s", size, bodyString)
		} else {
			// 处理类型断言失败的情况
			_, _ = fmt.Fprintf(&msgBuilder, "\n\tresponse_body(%d): [无法获取响应体]", size)
		}
	}

	// 根据状态码选择日志级别
	logByStatus(msgBuilder.String(), status, c.Errors)
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
		return &loggerBodyReader{body: []byte{}, maxSize: maxSize}, nil
	}

	// 只对可能需要记录的内容类型进行读取
	needRead := strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "text") ||
		strings.Contains(contentType, "form") ||
		contentType == ""

	if !needRead {
		// 对于二进制内容，不读取实际内容，只记录类型和大小
		return &loggerBodyReader{
			body:        []byte{},
			contentType: contentType,
			maxSize:     maxSize,
		}, nil
	}

	var data []byte
	var err error
	if maxSize >= 0 {
		data, err = io.ReadAll(io.LimitReader(body, int64(maxSize+1)))
	} else {
		data, err = io.ReadAll(body)
	}
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

	// 预处理所有需要的正则表达式，减少锁的次数
	regexps := make(map[string]*regexp.Regexp, len(sensitiveFields)*2)

	// 使用正则表达式进行更精确的敏感信息替换
	for _, field := range sensitiveFields {
		if !strings.Contains(bodyStr, field) {
			continue // 跳过不存在的字段名
		}

		// JSON 格式敏感字段
		jsonPattern := fmt.Sprintf(logJsonFieldRegexTpl, regexp.QuoteMeta(field))
		if re, err := logGetCachedRegexp(jsonPattern); err == nil {
			regexps[jsonPattern] = re
		}

		// 表单格式敏感字段
		formPattern := fmt.Sprintf(logFormFieldRegexTpl, regexp.QuoteMeta(field))
		if re, err := logGetCachedRegexp(formPattern); err == nil {
			regexps[formPattern] = re
		}
	}

	// 使用预先加载的正则表达式执行替换
	for pattern, re := range regexps {
		if strings.Contains(pattern, logJsonFieldRegexTpl) {
			bodyStr = re.ReplaceAllString(bodyStr, `$1"[REDACTED]"`)
		} else {
			bodyStr = re.ReplaceAllString(bodyStr, `$1[REDACTED]`)
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

// 在类型定义前添加接口确认检查
var (
	_ http.Hijacker = (*loggerResponseBodyWriter)(nil)
	_ http.Flusher  = (*loggerResponseBodyWriter)(nil)
	_ http.Pusher   = (*loggerResponseBodyWriter)(nil)
)

// 优化的响应体记录器
type loggerResponseBodyWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	maxLen int
	mu     sync.Mutex
}

func (w *loggerResponseBodyWriter) Write(b []byte) (int, error) {
	if w.body != nil && w.body.Len() < w.maxLen {
		w.mu.Lock()
		// 只记录到最大长度
		remaining := w.maxLen - w.body.Len()
		if remaining > len(b) || w.maxLen < 0 {
			w.body.Write(b)
		} else if remaining > 0 {
			w.body.Write(b[:remaining])
		}
		w.mu.Unlock()
	}
	return w.ResponseWriter.Write(b)
}

func (w *loggerResponseBodyWriter) WriteString(s string) (int, error) {
	if w.body != nil && w.body.Len() < w.maxLen {
		w.mu.Lock()
		// 只记录到最大长度
		remaining := w.maxLen - w.body.Len()
		if remaining > len(s) || w.maxLen < 0 {
			w.body.WriteString(s)
		} else if remaining > 0 {
			w.body.WriteString(s[:remaining])
		}
		w.mu.Unlock()
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

func (w *loggerResponseBodyWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (w *loggerResponseBodyWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *loggerResponseBodyWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}
