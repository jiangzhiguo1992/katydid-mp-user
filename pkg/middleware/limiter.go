package middleware

import (
	"context"
	"fmt"
	"hash/fnv"
	"katydid-mp-user/pkg/log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// LimiterStorageType 存储类型
type LimiterStorageType int

const (
	// Memory 内存存储
	Memory LimiterStorageType = iota
	// Redis Redis存储
	Redis
)

// LimiterOptions 限流器配置选项
type LimiterOptions struct {
	Code         int                       // 错误码
	Message      string                    // 自定义错误信息
	StorageType  LimiterStorageType        // 存储类型
	ShardCount   int                       // 分片数量，用于内存存储的分片锁
	RedisClient  *redis.Client             // Redis客户端(当StorageType为Redis时使用)
	WhitelistIPs []string                  // IP白名单
	KeyFunc      func(*gin.Context) string // 自定义键生成函数

	// 仅传值给defaultRule (全局的)
	DefLimit    int           // 限制的请求数量 (-1不限制，0关闭)
	DefDuration time.Duration // 时间窗口 (<=0不限时)
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
	Limit       int                     // 限制的请求数量 (-1不限制，0关闭)
	Duration    time.Duration           // 时间窗口 (<=0不限时)
	WhitelistFn func(*gin.Context) bool // 白名单判断函数
}

// ILimiterStorage 存储接口
type ILimiterStorage interface {
	Allow(key string, limit int, duration time.Duration) bool
	Close() error
}

var _ ILimiterStorage = (*LimiterMemoryStorage)(nil)
var _ ILimiterStorage = (*LimiterRedisStorage)(nil)

// LimiterMemoryStorage 内存存储实现
type LimiterMemoryStorage struct {
	shards       []*shard
	shardMask    uint32 // 分片掩码
	lastCleanup  int64  // 上次清理时间的原子访问
	cleanupMutex sync.Mutex
}

// shard 内存存储分片
type shard struct {
	sync.RWMutex
	timestamps map[string][]int64 // 时间戳列表(access)
}

// LimiterRedisStorage Redis存储实现
type LimiterRedisStorage struct {
	client *redis.Client
}

// Limiter 限流器
type Limiter struct {
	options LimiterOptions
	storage ILimiterStorage
	stats   LimiterStats
	rules   []LimitRule
	defRule LimitRule
	mu      sync.RWMutex
}

// NewLimiter 创建限流器
func NewLimiter(limit int, duration time.Duration) *Limiter {
	limiter := &Limiter{
		options: LimiterOptions{
			Code:         http.StatusTooManyRequests, // 429,
			Message:      "The request frequency exceeds the limit, please try again later",
			StorageType:  Memory,
			ShardCount:   32, // 默认32个分片
			RedisClient:  nil,
			WhitelistIPs: []string{},
			KeyFunc:      IPKeyFunc,
			//DefLimit:    limit,
			//DefDuration: duration,
		},
		stats: LimiterStats{},
		rules: make([]LimitRule, 0),
	}

	// 设置默认规则
	limiter.defRule = LimitRule{
		Path:        "*",
		Method:      "*",
		Limit:       limit,
		Duration:    duration,
		WhitelistFn: nil,
	}

	// 初始化存储
	limiter.initStorage()
	return limiter
}

// NewLimiterWithOptions 创建限流器并添加规则
func NewLimiterWithOptions(options LimiterOptions, rules ...LimitRule) *Limiter {
	limiter := NewLimiter(options.DefLimit, options.DefDuration)
	limiter.WithOptions(options)
	for _, rule := range rules {
		limiter.AddRule(rule)
	}
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
	if options.Message != "" {
		l.options.Message = options.Message
	}
	if options.ShardCount > 0 {
		l.options.ShardCount = options.ShardCount
	}
	if options.StorageType != l.options.StorageType {
		l.options.StorageType = options.StorageType
		l.options.RedisClient = options.RedisClient
		// 重新初始化存储
		if l.storage != nil {
			_ = l.storage.Close()
		}
		l.initStorage()
	}
	//if options.RedisClient != nil {
	//	l.options.RedisClient = options.RedisClient
	//	if l.options.StorageType == Redis {
	//		l.initStorage()
	//	}
	//}
	if options.WhitelistIPs != nil {
		l.options.WhitelistIPs = options.WhitelistIPs
	}
	if options.KeyFunc != nil {
		l.options.KeyFunc = options.KeyFunc
	}

	if options.DefLimit != 0 {
		l.defRule.Limit = options.DefLimit
	}
	if options.DefDuration != 0 {
		l.defRule.Duration = options.DefDuration
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
			l.storage = &LimiterRedisStorage{client: l.options.RedisClient}
		} else {
			log.Error("■ ■ Limiter ■ ■  Redis客户端未配置，使用内存存储")
			l.storage = newMemoryStorage(l.options.ShardCount)
		}
	default:
		l.storage = newMemoryStorage(l.options.ShardCount)
	}
}

// newMemoryStorage 创建内存存储
func newMemoryStorage(shardCount int) *LimiterMemoryStorage {
	// 确保分片数是2的幂
	n := 1
	for n < shardCount {
		n *= 2
	}

	ms := &LimiterMemoryStorage{
		shards:      make([]*shard, n),
		shardMask:   uint32(n - 1),
		lastCleanup: time.Now().Unix(),
	}

	for i := 0; i < n; i++ {
		ms.shards[i] = &shard{
			timestamps: make(map[string][]int64),
		}
	}
	return ms
}

// getShard 获取key对应的分片
func (ms *LimiterMemoryStorage) getShard(key string) *shard {
	h := fnv.New32()
	_, _ = h.Write([]byte(key))
	return ms.shards[h.Sum32()&ms.shardMask]
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

	// 找不到匹配的规则，返回默认规则
	return l.defRule
}

// Middleware 限流中间件
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 更新总请求计数
		atomic.AddInt64(&l.stats.TotalRequests, 1)

		// 获取请求标识
		var key string
		if l.options.KeyFunc != nil {
			key = l.options.KeyFunc(c)
		} else {
			key = IPKeyFunc(c)
		}

		// 查找适用的规则
		rule := l.findRule(c)

		// 检查白名单
		if l.isInWhitelist(key, c, rule) {
			log.DebugFmt("■ ■ Limiter ■ ■  白名单通过: IP=%s, 路径=%s, 方法=%s", key, c.FullPath(), c.Request.Method)
			c.Next()
			return
		}

		// 限流检查
		if !l.storage.Allow(key, rule.Limit, rule.Duration) {
			// 更新被限制请求计数
			atomic.AddInt64(&l.stats.LimitedRequests, 1)

			// 记录限流日志
			if gin.Mode() == gin.DebugMode {
				log.WarnFmt("■ ■ Limiter ■ ■  请求被限流: IP=%s, 路径=%s, 方法=%s", key, c.FullPath(), c.Request.Method)
			} else {
				log.WarnFmtOutput("■ ■ Limiter ■ ■  请求被限流: IP=%s, 路径=%s, 方法=%s", true, key, c.FullPath(), c.Request.Method)
			}

			// 返回限流响应
			ResponseData(c, http.StatusTooManyRequests, gin.H{
				"code": l.options.Code,
				"msg":  l.options.Message,
			})
			c.Abort()
			return
		}
		log.DebugFmt("■ ■ Limiter ■ ■  可以通行: IP=%s, 路径=%s, 方法=%s", key, c.FullPath(), c.Request.Method)

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
func (ms *LimiterMemoryStorage) Allow(key string, limit int, duration time.Duration) bool {
	// 获取对应分片
	s := ms.getShard(key)

	now := time.Now().Unix()
	expiredTime := now - int64(duration.Seconds())

	// 懒清理：定期清理过期记录
	if duration > 0 {
		lastCleanup := atomic.LoadInt64(&ms.lastCleanup)
		if now-lastCleanup > int64(duration.Seconds()/2) {
			// CAS操作，只有一个goroutine会成功更新lastCleanup
			if atomic.CompareAndSwapInt64(&ms.lastCleanup, lastCleanup, now) {
				go ms.cleanupExpired(expiredTime)
			}
		}
	}

	// 获取当前键的时间戳记录
	s.RLock()
	timestamps, exists := s.timestamps[key]
	s.RUnlock()

	// 键不存在，需要添加新记录
	if !exists {
		s.Lock()
		defer s.Unlock()

		// 再次检查，防止在获取锁期间被修改
		if timestamps, exists = s.timestamps[key]; !exists {
			s.timestamps[key] = []int64{now}
			return limit != 0
		}
	}

	// 计算有效请求数及有效时间戳
	validCount := 0
	validTimestamps := make([]int64, 0, len(timestamps)+1)
	for _, ts := range timestamps {
		if (duration <= 0) || (ts >= expiredTime) {
			validCount++
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// 如果已超限制，直接返回false
	if (limit >= 0) && (validCount >= limit) {
		return false
	}

	// 更新记录
	s.Lock()
	defer s.Unlock()

	// 再次检查，因为可能在释放读锁到获取写锁期间有变化
	timestamps, exists = s.timestamps[key]
	if !exists {
		s.timestamps[key] = []int64{now}
		return limit != 0
	}

	// 重新计算有效请求
	validCount = 0
	validTimestamps = validTimestamps[:0] // 重置切片但保留容量

	for _, ts := range timestamps {
		if (duration <= 0) || (ts >= expiredTime) {
			validCount++
			validTimestamps = append(validTimestamps, ts)
		}
	}

	if (limit >= 0) && (validCount >= limit) {
		return false
	}

	// 添加新记录并过滤过期记录
	s.timestamps[key] = append(validTimestamps, now)
	return true
}

// cleanupExpired 清理所有分片中的过期记录
func (ms *LimiterMemoryStorage) cleanupExpired(expiredTime int64) {
	log.InfoFmt("■ ■ Limiter ■ ■  清理过期记录START: %s", time.Unix(expiredTime, 0).Format("2006-01-02 15:04:05"))

	// 使用固定数量的goroutine并发清理各个分片，避免创建过多goroutine
	const maxWorkers = 4
	workers := min(maxWorkers, len(ms.shards))

	if workers <= 1 {
		// 分片数量少，直接清理
		for _, s := range ms.shards {
			ms.cleanupShard(s, expiredTime)
		}
		return
	}

	// 使用工作池并发清理
	var wg sync.WaitGroup
	shardChan := make(chan *shard, len(ms.shards))

	// 创建worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for s := range shardChan {
				ms.cleanupShard(s, expiredTime)
			}
		}()
	}

	// 分发任务
	for _, s := range ms.shards {
		shardChan <- s
	}
	close(shardChan)

	wg.Wait()
	log.InfoFmt("■ ■ Limiter ■ ■  清理过期记录OK: %s", time.Unix(expiredTime, 0).Format("2006-01-02 15:04:05"))
}

// cleanupShard 清理单个分片的过期记录
func (ms *LimiterMemoryStorage) cleanupShard(s *shard, expiredTime int64) {
	s.Lock()
	defer s.Unlock()
	log.InfoFmt("■ ■ Limiter ■ ■  清理过期记录(分片): %s", time.Unix(expiredTime, 0).Format("2006-01-02 15:04:05"))

	for key, timestamps := range s.timestamps {
		// 优化内存分配，预估需要的容量
		validCount := 0
		for _, ts := range timestamps {
			if ts >= expiredTime {
				validCount++
			}
		}

		// 所有记录都过期，直接删除键
		if validCount == 0 {
			delete(s.timestamps, key)
			continue
		}

		// 部分记录过期，过滤保留有效记录
		if validCount < len(timestamps) {
			valid := make([]int64, 0, validCount)
			for _, ts := range timestamps {
				if ts >= expiredTime {
					valid = append(valid, ts)
				}
			}
			s.timestamps[key] = valid
		}
		// 所有记录都有效，不需要处理
	}
}

// Close 关闭内存存储
func (ms *LimiterMemoryStorage) Close() error {
	for _, s := range ms.shards {
		s.Lock()
		s.timestamps = make(map[string][]int64)
		s.Unlock()
	}
	return nil
}

// Allow Redis存储实现的限流检查
func (rs *LimiterRedisStorage) Allow(key string, limit int, duration time.Duration) bool {
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
func (rs *LimiterRedisStorage) Close() error {
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

// AccountKeyFunc 按账号ID限流的键生成函数(需要认证中间件)
func AccountKeyFunc(c *gin.Context) string {
	userID, exists := c.Get(AuthKeyAccountID)
	if !exists {
		return c.ClientIP() // 回退到IP
	}
	return fmt.Sprintf("user:%v", userID)
}
