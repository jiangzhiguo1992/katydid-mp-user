package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LimiterOptions 限流器配置选项
type LimiterOptions struct {
	Code     int           // 错误码
	Limit    int           // 限制的请求数量
	Duration time.Duration // 时间窗口
	Message  string        // 自定义错误信息
}

// Limiter 限流器
type Limiter struct {
	sync.RWMutex                    // 读写锁保证并发安全
	options      LimiterOptions     // 限流器配置
	timestamps   map[string][]int64 // 请求的时间戳
	lastCleanup  time.Time          // 上次清理缓存的时间
}

// NewLimiter 创建限流器
func NewLimiter(limit int, duration time.Duration) *Limiter {
	return &Limiter{
		options: LimiterOptions{
			Code:     0,
			Limit:    limit,
			Duration: duration,
			Message:  "TooManyRequests",
		},
		timestamps:  make(map[string][]int64),
		lastCleanup: time.Now(),
	}
}

// WithOptions 设置限流器选项
func (l *Limiter) WithOptions(options LimiterOptions) *Limiter {
	l.Lock()
	defer l.Unlock()

	// 只覆盖非零值
	if options.Code != 0 {
		l.options.Code = options.Code
	}
	if options.Limit != 0 {
		l.options.Limit = options.Limit
	}
	if options.Duration != 0 {
		l.options.Duration = options.Duration
	}
	if options.Message != "" {
		l.options.Message = options.Message
	}
	return l
}

// cleanupExpiredRecords 清理所有过期记录
func (l *Limiter) cleanupExpiredRecords() {
	l.Lock()
	defer l.Unlock()

	now := time.Now().Unix()
	expiredTime := now - int64(l.options.Duration.Seconds())

	// 遍历所有IP，清理过期记录
	for ip, timestamps := range l.timestamps {
		valid := make([]int64, 0, len(timestamps))
		for _, ts := range timestamps {
			if ts >= expiredTime {
				valid = append(valid, ts)
			}
		}

		if len(valid) > 0 {
			l.timestamps[ip] = valid
		} else {
			// 如果该IP没有有效记录，直接删除键
			delete(l.timestamps, ip)
		}
	}

	l.lastCleanup = time.Now()
}

// isAllowed 检查请求是否被允许
func (l *Limiter) isAllowed(ip string) bool {
	l.Lock()
	defer l.Unlock()

	now := time.Now().Unix()
	expiredTime := now - int64(l.options.Duration.Seconds())

	// 懒清理：定期(每10分钟)清理所有过期记录
	if time.Since(l.lastCleanup) > 10*time.Minute {
		go l.cleanupExpiredRecords()
	}

	// 获取当前IP的时间戳记录
	timestamps, exists := l.timestamps[ip]
	if !exists {
		timestamps = make([]int64, 0, l.options.Limit)
	}

	// 过滤过期记录
	validCount := 0
	for _, ts := range timestamps {
		if ts >= expiredTime {
			validCount++
		}
	}

	// 检查是否超过限制
	if validCount >= l.options.Limit {
		return false
	}

	// 添加新记录
	l.timestamps[ip] = append(timestamps, now)
	return true
}

// Middleware 限流中间件
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP() // 获取客户端IP地址

		if !l.isAllowed(ip) {
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
