package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"katydid-mp-user/pkg/log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

const (
	CacheHeaderHit = "X-Cache"
	CacheHit       = "HIT"
	CacheMiss      = "MISS"
)

// CacheConfig 缓存中间件配置
type CacheConfig struct {
	DefaultExpiration time.Duration             // 默认缓存过期时间
	CleanupInterval   time.Duration             // 清理间隔
	KeyGenerator      func(*gin.Context) string // 自定义缓存键生成函数
	CachePaths        map[string]*time.Duration // 进行缓存的路径（支持前缀匹配）
	IgnoreQueryParams []string                  // 忽略的查询参数
	StatusCodes       []int                     // 需要缓存的状态码，默认只缓存200
	MaxItemSize       int                       // 最大缓存项大小(字节)，超过此大小的响应不缓存，0表示不限制
	WithHeaders       []string                  // 将这些请求头信息加入缓存键
	DisableCache      func(*gin.Context) bool   // 自定义禁用缓存的条件
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits      int64          // 缓存命中次数
	Misses    int64          // 缓存未命中次数
	ItemCount int            // 缓存项数量
	Size      int64          // 缓存大小(字节)
	Items     map[string]int // 按路径统计的缓存项数量
}

var (
	cacheStore *cache.Cache // 缓存存储go-cache

	cacheRegexps    = make(map[string]*regexp.Regexp) // 正则表达式缓存
	cacheRegexMutex sync.RWMutex                      // 正则表达式缓存的锁
	cacheStatsMutex sync.RWMutex                      // 统计信息的锁

	cacheStats = CacheStats{Items: make(map[string]int)}
)

// DefaultCacheConfig 返回默认配置
func DefaultCacheConfig(paths map[string]*time.Duration) CacheConfig {
	return CacheConfig{
		DefaultExpiration: 5 * time.Minute,
		CleanupInterval:   10 * time.Minute,
		KeyGenerator:      nil,
		CachePaths:        paths,
		IgnoreQueryParams: []string{},
		StatusCodes:       []int{http.StatusOK},
		MaxItemSize:       1024 * 1024, // 默认最大缓存1MB的响应
		WithHeaders:       []string{},
		DisableCache:      nil,
	}
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats() CacheStats {
	cacheStatsMutex.RLock()
	defer cacheStatsMutex.RUnlock()

	refreshStats()
	return cacheStats
}

// ClearCache 清空缓存
func ClearCache() {
	cacheStore.Flush()

	cacheStatsMutex.Lock()
	cacheStats.Hits = 0
	cacheStats.Misses = 0
	cacheStats.ItemCount = 0
	cacheStats.Size = 0
	cacheStats.Items = make(map[string]int)
	cacheStatsMutex.Unlock()
}

// refreshStats 刷新缓存统计信息
func refreshStats() {
	cacheStatsMutex.Lock()
	defer cacheStatsMutex.Unlock()

	cacheStats.ItemCount = cacheStore.ItemCount()
	cacheStats.Items = make(map[string]int)

	var totalSize int64
	for k, item := range cacheStore.Items() {
		if resp, ok := item.Object.(map[string]interface{}); ok {
			if body, ok := resp["body"].([]byte); ok {
				totalSize += int64(len(body))
			}

			// 按路径分组统计
			parts := strings.Split(k, "?")
			path := parts[0]
			if _, exists := cacheStats.Items[path]; exists {
				cacheStats.Items[path]++
			} else {
				cacheStats.Items[path] = 1
			}
		}
	}

	cacheStats.Size = totalSize
}

// DeleteCacheByPattern 删除符合模式的缓存
func DeleteCacheByPattern(pattern string) int {
	var count int
	for k := range cacheStore.Items() {
		if strings.Contains(k, pattern) {
			cacheStore.Delete(k)
			count++
		}
	}
	return count
}

// Cache 缓存中间件
func Cache(config CacheConfig) gin.HandlerFunc {
	cacheStore = cache.New(config.DefaultExpiration, config.CleanupInterval)

	// 如果没有设置键生成器，使用默认的
	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultCacheKeyGenerator(config.IgnoreQueryParams, config.WithHeaders)
	}

	return func(c *gin.Context) {
		// 自定义禁用缓存条件
		if config.DisableCache != nil && config.DisableCache(c) {
			c.Next()
			return
		}

		// 判断是否需要缓存当前路径
		shouldCachePath := false
		duration := config.DefaultExpiration
		path := c.Request.URL.Path
		for cachePath, du := range config.CachePaths {
			if cacheMatchRegex(path, cachePath) {
				shouldCachePath = true
				if du != nil {
					duration = *du
				}
				break
			}
		}

		if !shouldCachePath {
			c.Next()
			return
		}

		// 非GET请求不缓存
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := config.KeyGenerator(c)
		log.DebugFmt("■ ■ Cache ■ ■ 网络缓存查找:%s", cacheKey)

		// 尝试从缓存获取
		if response, found := cacheStore.Get(cacheKey); found {
			log.DebugFmt("■ ■ Cache ■ ■ 网络缓存命中:%s", cacheKey)

			cachedResponse := response.(map[string]interface{})

			// 恢复状态码和头信息
			statusCode := cachedResponse["status_code"].(int)
			headers := cachedResponse["headers"].(map[string]string)
			body := cachedResponse["body"].([]byte)

			// 设置头信息
			for k, v := range headers {
				c.Header(k, v)
			}

			// 添加缓存命中标记, 更新统计信息
			c.Header(CacheHeaderHit, CacheHit)
			atomic.AddInt64(&cacheStats.Hits, 1)

			// 设置响应
			c.Data(statusCode, headers["Content-Type"], body)

			c.Abort()
			return
		}
		log.DebugFmt("■ ■ Cache ■ ■ 网络缓存未命中:%s", cacheKey)

		// 更新未命中统计, 添加缓存未命中标记
		c.Header(CacheHeaderHit, CacheMiss)
		atomic.AddInt64(&cacheStats.Misses, 1)

		// 创建响应写入器
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 检查是否需要缓存该响应
		shouldCache := false
		for _, code := range config.StatusCodes {
			if writer.status == code {
				shouldCache = true
				break
			}
		}

		if shouldCache {
			body := writer.body.Bytes()

			// 检查响应大小是否超过限制
			if config.MaxItemSize > 0 && len(body) > config.MaxItemSize {
				return
			}

			// 收集头信息
			headers := make(map[string]string)
			for k, v := range writer.Header() {
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}

			// 创建缓存对象
			response := map[string]interface{}{
				"status_code": writer.status,
				"headers":     headers,
				"body":        body,
			}

			// 存入缓存
			cacheStore.Set(cacheKey, response, duration)
			log.DebugFmt("■ ■ Cache ■ ■ 网络缓存更新:%s, expire=%d", cacheKey, duration.Seconds())
		}
	}
}

// defaultCacheKeyGenerator 默认缓存键生成器
func defaultCacheKeyGenerator(ignoreParams []string, withHeaders []string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		var keyParts []string

		// 加入URL路径
		keyParts = append(keyParts, c.Request.URL.Path)

		// 处理查询参数，忽略���定的参数
		if len(c.Request.URL.RawQuery) > 0 {
			query := c.Request.URL.Query()
			filteredQuery := make([]string, 0)

			for key, values := range query {
				// 检查是否为需要忽略的查询参数
				ignored := false
				for _, ignoreParam := range ignoreParams {
					if key == ignoreParam {
						ignored = true
						break
					}
				}

				if !ignored {
					for _, value := range values {
						filteredQuery = append(filteredQuery, key+"="+value)
					}
				}
			}

			// 按字母顺序排序查询参数，确保相同参数不同顺序生成相同的键
			sort.Strings(filteredQuery)

			if len(filteredQuery) > 0 {
				keyParts = append(keyParts, strings.Join(filteredQuery, "&"))
			}
		}

		// 添加指定的请求头到键
		for _, header := range withHeaders {
			if value := c.GetHeader(header); value != "" {
				keyParts = append(keyParts, fmt.Sprintf("%s=%s", header, value))
			}
		}

		// 将所有部分连接起来
		key := strings.Join(keyParts, "|")

		// 对键进行哈希处理，避免键过长
		h := sha256.New()
		h.Write([]byte(key))
		return hex.EncodeToString(h.Sum(nil))
	}
}

// cacheMatchRegex 判断路径是否匹配正则表达式
func cacheMatchRegex(path string, pattern string) bool {
	// 如果不是正则表达式，直接前��匹配
	if !strings.HasPrefix(pattern, "^") && !strings.HasSuffix(pattern, "$") {
		return strings.HasPrefix(path, pattern)
	}

	// 使用正则表达式匹配
	cacheRegexMutex.RLock()
	re, exists := cacheRegexps[pattern]
	cacheRegexMutex.RUnlock()

	// 如果不存在，则编译正则表达式并缓存
	if !exists {
		cacheRegexMutex.Lock()
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			cacheRegexMutex.Unlock()
			return false
		}
		cacheRegexps[pattern] = re
		cacheRegexMutex.Unlock()
	}
	return re.MatchString(path)
}

// responseWriter 是一个记录响应的ResponseWriter
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

// Write 实现ResponseWriter接口
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 实现ResponseWriter接口
func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// WriteHeader 实现ResponseWriter接口
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// ReadFrom 实现io.ReaderFrom接口，提高大文件处理效率
func (w *responseWriter) ReadFrom(reader io.Reader) (n int64, err error) {
	// 同时写入缓冲区和原始ResponseWriter
	buf := &bytes.Buffer{}
	n, err = io.Copy(io.MultiWriter(buf, w.ResponseWriter), reader)
	w.body = buf
	return
}
