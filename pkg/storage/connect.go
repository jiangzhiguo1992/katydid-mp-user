package storage

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

// 数据库类型常量
const (
	DBTypePgSQL DBType = "pgsql"
	DBTypeMySQL DBType = "mysql"
)

// 全局连接池，可以支持多个数据库实例
var (
	dbInstances = make(map[string]*gorm.DB)
	dbMutex     sync.RWMutex
)

type (
	// DBConfig 数据库配置结构
	DBConfig struct {
		Type DBType // 数据库类型

		Host     string // 主机地址
		Port     int    // 端口号
		User     string // 用户名
		Password string // 密码
		DBName   string // 数据库名称
		SSLMode  string // SSL模式
		TimeZone string // 时区

		MaxRetries int // 重试次数
		RetryDelay int // 重试间隔，单位秒

		MaxIdle     int           // 最大空闲连接数，默认1000
		MaxOpen     int           // 最大连接数，默认10000
		MaxLifeTime time.Duration // 连接最大存活时间，默认-1
		MaxIdleTime time.Duration // 空闲连接最大存活时间，默认3m

		LogLevel logger.LogLevel
		Debug    bool // 是否开启调试模式

	}

	// DBType 数据库类型
	DBType string
)

func GetDefaultDBConfig(tp DBType) DBConfig {
	return DBConfig{
		Type: tp,
	}
}

// InitConnect 初始化数据库连接
func InitConnect(name string, config DBConfig) (*gorm.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if _, exists := dbInstances[name]; exists {
		return nil, errors.New("数据库连接已存在")
	}

	var dialector gorm.Dialector
	switch config.Type {
	case DBTypeMySQL:
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.DBName,
		)
		dialector = mysql.New(mysql.Config{
			DSN: dsn,
		})
	case DBTypePgSQL:
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s",
			config.Host, config.Port, config.User, config.Password, config.DBName,
		)
		if len(config.SSLMode) > 0 {
			dsn += fmt.Sprintf(" sslmode=%s", config.SSLMode)
		}
		if len(config.TimeZone) > 0 {
			dsn += fmt.Sprintf(" TimeZone=%s", config.TimeZone)
		}
		dialector = postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: false,
			// TODO:GG 看看gorm的文档
		})
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	// TODO:GG zap 所有的日志都zap?
	logMode := config.LogLevel
	if config.Debug {
		logMode = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger:                 logger.Default.LogMode(logMode),
		PrepareStmt:            true, // 缓存预编译语句以提高性能
		SkipDefaultTransaction: true, // 跳过默认事务，提升性能
		// TODO:GG 配置 连接池，超时时间等，看看gorm的文档
	}

	// 重试连接逻辑
	var db *gorm.DB
	var err error
	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3 // 默认重试3次
	}
	retryInterval := time.Duration(config.RetryDelay) * time.Second
	if retryInterval <= 0 {
		retryInterval = 2 * time.Second // 默认重试间隔2秒
	}
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(dialector, gormConfig)
		if (db != nil) && (err == nil) {
			break // 连接成功则跳出循环
		}

		select {
		case <-time.After(retryInterval):
		case <-context.Background().Done():
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	} else if db == nil {
		return nil, fmt.Errorf("数据库连接为空")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetConnMaxLifetime(config.MaxLifeTime)
	sqlDB.SetConnMaxIdleTime(config.MaxIdleTime)

	// 保存到实例map中
	dbInstances[name] = db

	return db, nil
}

// GetDB 通过名称获取数据库连接
func GetDB(name string) *gorm.DB {
	dbMutex.RLock()
	db, ok := dbInstances[name]
	dbMutex.RUnlock()

	if !ok || db == nil {
		return nil
	}
	return db
}

// CloseDB 关闭指定的数据库连接
func CloseDB(name string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	db, exists := dbInstances[name]
	if !exists || db == nil {
		return errors.New("未找到指定的数据库连接")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	delete(dbInstances, name)
	return sqlDB.Close()
}

// CloseAllDBs 关闭所有数据库连接
func CloseAllDBs() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	var lastErr error
	for name, db := range dbInstances {
		sqlDB, err := db.DB()
		if err != nil {
			lastErr = err
			continue
		}

		if err = sqlDB.Close(); err != nil {
			lastErr = err
		}
		delete(dbInstances, name)
	}

	return lastErr
}

// Ping 检查数据库连接状态
func Ping(name string) error {
	db := GetDB(name)
	if db == nil {
		return errors.New("未找到指定的数据库连接")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats 获取数据库连接池统计信息
func Stats(name string) (interface{}, error) {
	db := GetDB(name)
	if db == nil {
		return nil, errors.New("未找到指定的数据库连接")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Stats(), nil
}
