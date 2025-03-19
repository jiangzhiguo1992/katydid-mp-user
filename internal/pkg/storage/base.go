package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 表名常量定义
const (
	TableGroupAuth       TableName = "auths"
	TableAuthVerify                = TableGroupAuth + ".verify"
	TableAuthAccount               = TableGroupAuth + ".account"
	TableAuthAuth                  = TableGroupAuth + ".auth"
	TableAuthAccountAuth           = TableGroupAuth + ".account_auth"
	TableAuthToken                 = TableGroupAuth + ".token"
	TableAuthAccess                = TableGroupAuth + ".access"

	TableGroupUser TableName = "users"

	TableGroupClient TableName = "clients"

	TableGroupRole TableName = "roles"
)

// 错误定义
var (
	ErrDBNotInitialized = errors.New("数据库未初始化")
	ErrDBExists         = errors.New("数据库连接已存在")
	ErrDBNotFound       = errors.New("未找到指定的数据库连接")
)

// 全局连接池，可以支持多个数据库实例
var (
	dbInstances = make(map[string]*gorm.DB)
	dbMutex     sync.RWMutex
)

type (
	// TableName 表名类型
	TableName string

	// DBType 数据库类型
	DBType string

	// DBConfig 数据库配置结构
	DBConfig struct {
		Type     DBType
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
		MaxIdle  int
		MaxOpen  int
		Timeout  time.Duration
		TimeZone string
		LogLevel logger.LogLevel
	}

	// Base 提供基础数据库功能
	Base struct {
		ctx context.Context
	}
)

// 数据库类型常量
const (
	DBTypePostgres DBType = "postgres"
	DBTypeMySQL    DBType = "mysql"
)

// 默认连接名称
const (
	DefaultPsqlName = "default_psql"
	DefaultMsqlName = "default_mysql"
)

// NewBase 创建一个新的Base实例
func NewBase(ctx context.Context) *Base {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Base{
		ctx: ctx,
	}
}

// WithContext 设置上下文
func (b *Base) WithContext(ctx context.Context) *Base {
	return &Base{
		ctx: ctx,
	}
}

// Context 获取上下文
func (b *Base) Context() context.Context {
	if b.ctx == nil {
		return context.Background()
	}
	return b.ctx
}

// Psql 获取PostgreSQL连接
func (b *Base) Psql() *gorm.DB {
	return b.GetDB(DefaultPsqlName).WithContext(b.Context())
}

// Msql 获取MySQL连接
func (b *Base) Msql() *gorm.DB {
	return b.GetDB(DefaultMsqlName).WithContext(b.Context())
}

// GetDB 通过名称获取数据库连接
func (b *Base) GetDB(name string) *gorm.DB {
	dbMutex.RLock()
	db, ok := dbInstances[name]
	dbMutex.RUnlock()

	if !ok || db == nil {
		return nil
	}
	return db
}

// InitDB 初始化数据库连接
func InitDB(name string, config DBConfig) (*gorm.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if _, exists := dbInstances[name]; exists {
		return nil, ErrDBExists
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
	case DBTypePostgres:
		host := fmt.Sprintf("host=%s", config.Host)
		port := fmt.Sprintf("port=%d", config.Port)
		user := fmt.Sprintf("user=%s", config.User)
		password := fmt.Sprintf("password=%s", config.Password)
		dbName := fmt.Sprintf("dbname=%s", config.DBName)
		sslMode := ""
		if len(config.SSLMode) > 0 {
			sslMode = fmt.Sprintf("sslmode=%s", config.SSLMode)
		}
		timeZone := ""
		if len(config.TimeZone) > 0 {
			timeZone = fmt.Sprintf("TimeZone=%s", config.TimeZone)
		}
		dsn := fmt.Sprintf(
			"%s %s %s %s %s %s %s",
			host, port, user, password, dbName, sslMode, timeZone,
		)
		dialector = postgres.New(postgres.Config{
			DSN: dsn,
		})
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetConnMaxLifetime(config.Timeout)
	//sqlDB.SetConnMaxIdleTime()

	// 保存到实例map中
	dbInstances[name] = db

	// 设置默认实例
	if name == DefaultPsqlName {
		dbInstances["psql"] = db
	} else if name == DefaultMsqlName {
		dbInstances["msql"] = db
	}

	return db, nil
}

// CloseDB 关闭指定的数据库连接
func CloseDB(name string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	db, exists := dbInstances[name]
	if !exists || db == nil {
		return ErrDBNotFound
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

		if err := sqlDB.Close(); err != nil {
			lastErr = err
		}
		delete(dbInstances, name)
	}

	return lastErr
}

// Transaction 事务包装函数
func (b *Base) Transaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	if db == nil {
		return ErrDBNotInitialized
	}
	return db.WithContext(b.Context()).Transaction(fn)
}

// Ping 检查数据库连接状态
func (b *Base) Ping(name string) error {
	db := b.GetDB(name)
	if db == nil {
		return ErrDBNotFound
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats 获取数据库连接池统计信息
func (b *Base) Stats(name string) (interface{}, error) {
	db := b.GetDB(name)
	if db == nil {
		return nil, ErrDBNotFound
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Stats(), nil
}
