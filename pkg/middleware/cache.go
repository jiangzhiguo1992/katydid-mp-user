package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// CacheConfig 缓存中间件配置
type CacheConfig struct {
	DefaultExpiration time.Duration             // 默认缓存过期时间
	CleanupInterval   time.Duration             // 清理间隔
	KeyGenerator      func(*gin.Context) string // 自定义缓存键生成函数
	CachePaths        []string                  // 进行缓存的路径
	IgnoreQueryParams []string                  // 忽略的查询参数
	StatusCodes       []int                     // 需要缓存的状态码，默认只缓存200
}

// DefaultCacheConfig 返回默认配置
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		DefaultExpiration: 5 * time.Minute,
		CleanupInterval:   10 * time.Minute,
		KeyGenerator:      nil,
		CachePaths:        []string{},
		IgnoreQueryParams: []string{},
		StatusCodes:       []int{http.StatusOK},
	}
}

const CacheHeaderHit = "X-Cache"

var (
	store *cache.Cache // 缓存存储go-cache
)

// Cache 缓存中间件
func Cache(config CacheConfig) gin.HandlerFunc {
	// 如果没有设置键生成器，使用默认的
	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultKeyGenerator(config.IgnoreQueryParams)
	}

	return func(c *gin.Context) {
		// 判断是否需要缓存当前路径
		shouldCachePath := false
		if len(config.CachePaths) == 0 {
			// 如果没有指定缓存路径，则默认缓存所有
			shouldCachePath = true
		} else {
			path := c.Request.URL.Path
			for _, cachePath := range config.CachePaths {
				if strings.HasPrefix(path, cachePath) {
					shouldCachePath = true
					break
				}
			}
		}

		if !shouldCachePath {
			c.Next()
			return
		}

		// 非GET请求不缓存
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := config.KeyGenerator(c)

		// 尝试从缓存获取
		if response, found := store.Get(cacheKey); found {
			cachedResponse := response.(map[string]interface{})

			// 恢复状态码和头信息
			statusCode := cachedResponse["status_code"].(int)
			headers := cachedResponse["headers"].(map[string]string)
			body := cachedResponse["body"].([]byte)

			// 设置头信息
			for k, v := range headers {
				c.Header(k, v)
			}

			// 添加缓存命中标记
			c.Header(CacheHeaderHit, "HIT")

			// 设置响应
			c.Data(statusCode, headers["Content-Type"], body)
			c.Abort()
			return
		}

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
				"body":        writer.body.Bytes(),
			}

			// 存入缓存
			store.Set(cacheKey, response, config.DefaultExpiration)
		}
	}
}

// defaultKeyGenerator 默认缓存键生成器
func defaultKeyGenerator(ignoreParams []string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		// 生成包含URL路径的键
		path := c.Request.URL.Path

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
				path += "?" + strings.Join(filteredQuery, "&")
			}
		}

		// 对键进行哈希处理，避免键过长
		h := sha256.New()
		h.Write([]byte(path))
		return hex.EncodeToString(h.Sum(nil))
	}
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
