package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// StorageType 存储类型
type StorageType int

const (
	// Memory 内存存储
	Memory StorageType = iota
	// Redis Redis存储
	Redis
)

// LimiterOptions 限流器配置选项
type LimiterOptions struct {
	Code         int                       // 错误码
	Limit        int                       // 限制的请求数量
	Duration     time.Duration             // 时间窗口
	Message      string                    // 自定义错误信息
	StorageType  StorageType               // 存储类型
	RedisClient  *redis.Client             // Redis客户端(当StorageType为Redis时使用)
	WhitelistIPs []string                  // IP白名单
	KeyFunc      func(*gin.Context) string // 自定义键生成函数
	LogFunc      func(string, ...any)      // 日志函数
}

// LimiterStats 限流器统计信息
type LimiterStats struct {
	TotalRequests   int64 // 总请求数
	LimitedRequests int64 // 被限制的请求数
	ActiveIPs       int   // 活跃IP数
}

// LimitRule 限流规则
type LimitRule struct {
	Path        string                  // 路径
	Method      string                  // 方法
	Limit       int                     // 限制
	Duration    time.Duration           // 时间窗口
	WhitelistFn func(*gin.Context) bool // 白名单判断函数
}

// Storage 存储接口
type Storage interface {
	Allow(key string, limit int, duration time.Duration) bool
	Close() error
}

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	sync.RWMutex
	timestamps  map[string][]int64
	lastCleanup time.Time
}

// RedisStorage Redis存储实现
type RedisStorage struct {
	client *redis.Client
}

// Limiter 限流器
type Limiter struct {
	options     LimiterOptions
	storage     Storage
	stats       LimiterStats
	mu          sync.RWMutex
	rules       []LimitRule
	defaultRule LimitRule
}

// NewLimiter 创建限流器
func NewLimiter(limit int, duration time.Duration) *Limiter {
	limiter := &Limiter{
		options: LimiterOptions{
			Code:        429,
			Limit:       limit,
			Duration:    duration,
			Message:     "请求频率超过限制，请稍后再试",
			StorageType: Memory,
			LogFunc:     func(format string, args ...any) {},
		},
		stats: LimiterStats{},
		rules: make([]LimitRule, 0),
	}

	// 设置默认规则
	limiter.defaultRule = LimitRule{
		Path:     "*",
		Method:   "*",
		Limit:    limit,
		Duration: duration,
	}

	// 初始化存储
	limiter.initStorage()
	return limiter
}

// WithOptions 设置限流器选项
func (l *Limiter) WithOptions(options LimiterOptions) *Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 覆盖非零值
	if options.Code != 0 {
		l.options.Code = options.Code
	}
	if options.Limit != 0 {
		l.options.Limit = options.Limit
		l.defaultRule.Limit = options.Limit
	}
	if options.Duration != 0 {
		l.options.Duration = options.Duration
		l.defaultRule.Duration = options.Duration
	}
	if options.Message != "" {
		l.options.Message = options.Message
	}
	if options.StorageType != l.options.StorageType {
		l.options.StorageType = options.StorageType
		// 重新初始化存储
		if l.storage != nil {
			_ = l.storage.Close()
		}
		l.initStorage()
	}
	if options.RedisClient != nil {
		l.options.RedisClient = options.RedisClient
		if l.options.StorageType == Redis {
			l.initStorage()
		}
	}
	if options.WhitelistIPs != nil {
		l.options.WhitelistIPs = options.WhitelistIPs
	}
	if options.KeyFunc != nil {
		l.options.KeyFunc = options.KeyFunc
	}
	if options.LogFunc != nil {
		l.options.LogFunc = options.LogFunc
	}

	return l
}

// AddRule 添加限流规则
func (l *Limiter) AddRule(rule LimitRule) *Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rules = append(l.rules, rule)
	return l
}

// initStorage 初始化存储
func (l *Limiter) initStorage() {
	switch l.options.StorageType {
	case Redis:
		if l.options.RedisClient != nil {
			l.storage = &RedisStorage{client: l.options.RedisClient}
		} else {
			l.options.LogFunc("Redis客户端未配置，使用内存存储")
			l.storage = &MemoryStorage{
				timestamps:  make(map[string][]int64),
				lastCleanup: time.Now(),
			}
		}
	default:
		l.storage = &MemoryStorage{
			timestamps:  make(map[string][]int64),
			lastCleanup: time.Now(),
		}
	}
}

// GetStats 获取限流器统计信息
func (l *Limiter) GetStats() LimiterStats {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.stats
}

// findRule 查找匹配的规则
func (l *Limiter) findRule(c *gin.Context) LimitRule {
	path := c.FullPath()
	method := c.Request.Method

	for _, rule := range l.rules {
		if (rule.Path == "*" || rule.Path == path) &&
			(rule.Method == "*" || rule.Method == method) {
			return rule
		}
	}

	return l.defaultRule
}

// Middleware 限流中间件
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 更新总请求计数
		l.mu.Lock()
		l.stats.TotalRequests++
		l.mu.Unlock()

		// 获取请求标识
		var key string
		if l.options.KeyFunc != nil {
			key = l.options.KeyFunc(c)
		} else {
			key = c.ClientIP()
		}

		// 查找适用的规则
		rule := l.findRule(c)

		// 检查白名单
		if l.isInWhitelist(key, c, rule) {
			c.Next()
			return
		}

		// 限流检查
		if !l.storage.Allow(key, rule.Limit, rule.Duration) {
			// 更新被限制请求计数
			l.mu.Lock()
			l.stats.LimitedRequests++
			l.mu.Unlock()

			// 记录限流日志
			l.options.LogFunc("请求被限流: IP=%s, 路径=%s, 方法=%s", key, c.FullPath(), c.Request.Method)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": l.options.Code,
				"msg":  l.options.Message,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isInWhitelist 检查是否在白名单中
func (l *Limiter) isInWhitelist(key string, c *gin.Context, rule LimitRule) bool {
	// 检查规则特定的白名单函数
	if rule.WhitelistFn != nil && rule.WhitelistFn(c) {
		return true
	}

	// 检查IP白名单
	for _, ip := range l.options.WhitelistIPs {
		if ip == key {
			return true
		}
	}
	return false
}

// Allow 内存存储实现的限流检查
func (ms *MemoryStorage) Allow(key string, limit int, duration time.Duration) bool {
	ms.Lock()
	defer ms.Unlock()

	now := time.Now().Unix()
	expiredTime := now - int64(duration.Seconds())

	// 懒清理：定期(每10分钟)清理过期记录
	if time.Since(ms.lastCleanup) > 10*time.Minute {
		go ms.cleanup(expiredTime)
	}

	// 获取当前键的时间戳记录
	timestamps, exists := ms.timestamps[key]
	if !exists {
		timestamps = make([]int64, 0, limit)
	}

	// 过滤过期记录
	validCount := 0
	for _, ts := range timestamps {
		if ts >= expiredTime {
			validCount++
		}
	}

	// 检查是否超过限制
	if validCount >= limit {
		return false
	}

	// 添加新记录
	ms.timestamps[key] = append(timestamps, now)
	return true
}

// cleanup 清理过期记录
func (ms *MemoryStorage) cleanup(expiredTime int64) {
	ms.Lock()
	defer ms.Unlock()

	for key, timestamps := range ms.timestamps {
		valid := make([]int64, 0, len(timestamps))
		for _, ts := range timestamps {
			if ts >= expiredTime {
				valid = append(valid, ts)
			}
		}

		if len(valid) > 0 {
			ms.timestamps[key] = valid
		} else {
			delete(ms.timestamps, key)
		}
	}

	ms.lastCleanup = time.Now()
}

// Close 关闭内存存储
func (ms *MemoryStorage) Close() error {
	ms.Lock()
	ms.timestamps = make(map[string][]int64)
	ms.Unlock()
	return nil
}

// Allow Redis存储实现的限流检查
func (rs *RedisStorage) Allow(key string, limit int, duration time.Duration) bool {
	ctx := context.Background()
	now := time.Now().UnixMilli()

	// 使用Redis的ZREMRANGEBYSCORE移除过期记录
	expiredTime := now - duration.Milliseconds()
	rs.client.ZRemRangeByScore(ctx, "ratelimit:"+key, "0", fmt.Sprintf("%d", expiredTime))

	// 添加当前请求时间戳
	rs.client.ZAdd(ctx, "ratelimit:"+key, redis.Z{Score: float64(now), Member: now})

	// 设置过期时间以避免内存泄漏
	rs.client.Expire(ctx, "ratelimit:"+key, duration*2)

	// 计数当前窗口内的请求数
	count, err := rs.client.ZCard(ctx, "ratelimit:"+key).Result()
	if err != nil {
		return false // 发生错误时限流
	}

	return count <= int64(limit)
}

// Close 关闭Redis存储
func (rs *RedisStorage) Close() error {
	return nil // Redis客户端由外部管理，无需在此关闭
}

// RateLimiter 创建默认限流中间件
func RateLimiter(limit int, duration time.Duration) gin.HandlerFunc {
	return NewLimiter(limit, duration).Middleware()
}

// IPKeyFunc 按IP限流的键生成函数
func IPKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// PathKeyFunc 按路径限流的键生成函数
func PathKeyFunc(c *gin.Context) string {
	return c.FullPath()
}

// UserKeyFunc 按用户ID限流的键生成函数(需要认证中间件)
func UserKeyFunc(c *gin.Context) string {
	userID, exists := c.Get(AuthKeyUserID)
	if !exists {
		return c.ClientIP() // 回退到IP
	}
	return fmt.Sprintf("user:%v", userID)
}
