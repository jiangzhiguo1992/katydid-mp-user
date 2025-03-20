package storage

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

// 数据库类型常量
const (
	DBKindPgSQL  DBKind = "pgsql"
	DBKindMySQL  DBKind = "mysql"
	DBKindSQLite DBKind = "sqlite"
)

// 全局连接池，可以支持多个数据库实例
var (
	dbInstances = make(map[string]*gorm.DB)
	dbMutex     sync.RWMutex
)

type (
	// DBKind 数据库类型
	DBKind string

	// DBConfig 数据库配置结构
	DBConfig struct {
		Kind   DBKind // 数据库类型
		Logger logger.Interface

		Host     string // 主机地址
		Port     int    // 端口号
		User     string // 用户名
		Password string // 密码
		DBName   string // 数据库名称

		MaxRetries int // 重试次数
		RetryDelay int // 重试间隔，单位秒

		MaxOpen     int           // 最大连接数，默认1000 (看数据库性能)
		MaxIdle     int           // 最大空闲连接数，默认=Open
		MaxLifeTime time.Duration // 连接最大存活时间，默认3m
		MaxIdleTime time.Duration // 空闲连接最大存活时间，默认1m

		PgsqlSSLMode  string // SSL模式
		PgsqlTimeZone string // 时区
		SQLiteFile    string // SQLite文件路径 (file::memory:?cache=shared，是内存sqlite的意思)
	}
)

// InitConnect 初始化数据库连接
func InitConnect(name string, config DBConfig) (*gorm.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if _, exists := dbInstances[name]; exists {
		return nil, fmt.Errorf("■ ■ connect ■ ■ 数据库连接已存在: %s", name)
	}

	var dialector gorm.Dialector
	switch config.Kind {
	case DBKindMySQL:
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.DBName,
		)
		dialector = mysql.New(mysql.Config{
			DSN:                       dsn,
			SkipInitializeWithVersion: false, // 自动适配 MySQL 版本特性
			DisableWithReturning:      false, // 保留RETURNING子句以提高效率
			DisableDatetimePrecision:  false, // 自动解析时间
			//DefaultStringSize:         256,   // 默认值
			//DisableDatetimePrecision:  true,  // 如果使用 MySQL 5.6 及以下版本
			//DontSupportRenameIndex:    true,  // 如果使用 MySQL 5.7 及以下版本
			//DontSupportRenameColumn:   true,  // 如果使用 MySQL 8.0 以下版本
		})
	case DBKindPgSQL:
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s",
			config.Host, config.Port, config.User, config.Password, config.DBName,
		)
		if len(config.PgsqlSSLMode) > 0 {
			dsn += fmt.Sprintf(" sslmode=%s", config.PgsqlSSLMode)
		}
		if len(config.PgsqlTimeZone) > 0 {
			dsn += fmt.Sprintf(" TimeZone=%s", config.PgsqlTimeZone)
		}
		dialector = postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: false, // 保持为false以获得更好的预处理语句性能
			WithoutReturning:     false, // 保留RETURNING子句以提高效率
			//WithoutQuotingCheck:  false, // 需要手动处理标识符时设为true
			// DriverName:        "pgx",  // 高性能场景可考虑使用pgx驱动
		})
	case DBKindSQLite:
		if len(config.SQLiteFile) == 0 {
			return nil, fmt.Errorf("■ ■ connect ■ ■ SQLite数据库文件路径不能为空")
		}
		dialector = sqlite.Open(config.SQLiteFile)
	default:
		return nil, fmt.Errorf("■ ■ connect ■ ■ 不支持的数据库类型: %s", config.Kind)
	}

	if config.Logger == nil {
		config.Logger = logger.Default.LogMode(logger.Warn)
	}

	gormConfig := &gorm.Config{
		Logger:                   config.Logger,
		DisableAutomaticPing:     false, // 初始化后，自动ping数据库
		SkipDefaultTransaction:   true,  // 跳过默认事务，提升性能
		DisableNestedTransaction: false, // 打开嵌套事务
		AllowGlobalUpdate:        false, // 不允许进行全局update/delete
		PrepareStmt:              true,  // 缓存预编译语句以提高性能
		QueryFields:              false, // 允许 SELECT *
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		DisableForeignKeyConstraintWhenMigrating: true, // (自动迁移/创建表)不要自动创建外键
		IgnoreRelationshipsWhenMigrating:         true, // (自动迁移/创建表)忽略关系
		//DryRun: false, // 是否启用干运行模式
		//NamingStrategy: schema.NamingStrategy{},
		//FullSaveAssociations
		//PropagateUnscoped:
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
		return nil, fmt.Errorf("■ ■ connect ■ ■ 数据库连接失败: %s, %w", name, err)
	} else if db == nil {
		return nil, fmt.Errorf("■ ■ connect ■ ■ 数据库连接为空: %s", name)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetMaxIdleConns(config.MaxIdle)
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
		return fmt.Errorf("■ ■ connect ■ ■ 关闭: 未找到指定的数据库连接:%s", name)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	delete(dbInstances, name)
	return sqlDB.Close()
}

// CloseAllDBs 关闭所有数据库连接
func CloseAllDBs() []error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	var errs []error
	for name, db := range dbInstances {
		if db == nil {
			errs = append(errs, fmt.Errorf("■ ■ connect ■ ■ 关闭: 未找到指定的数据库连接:%s", name))
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if err = sqlDB.Close(); err != nil {
			errs = append(errs, err)
		}
		delete(dbInstances, name)
	}
	return errs
}

// Ping 检查数据库连接状态
func Ping(name string) error {
	db := GetDB(name)
	if db == nil {
		return fmt.Errorf("■ ■ connect ■ ■ Ping: 未找到指定的数据库连接:%s", name)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats 获取数据库连接池统计信息
func Stats(name string) (sql.DBStats, error) {
	db := GetDB(name)
	if db == nil {
		return sql.DBStats{}, fmt.Errorf("■ ■ connect ■ ■ Stats: 未找到指定的数据库连接:%s", name)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return sql.DBStats{}, err
	}
	return sqlDB.Stats(), nil
}
