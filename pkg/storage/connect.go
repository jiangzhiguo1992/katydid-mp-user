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

// DBKind 数据库类型
type DBKind string

// 数据库类型常量
const (
	DBKindPgSQL  DBKind = "pgsql"
	DBKindMySQL  DBKind = "mysql"
	DBKindSQLite DBKind = "sqlite"
)

// 全局连接池，可以支持多个数据库实例
var (
	dbInstances = make(map[string]*DBInstance)
	dbMutex     sync.RWMutex
)

type (
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

		HealthCheckInterval time.Duration // 健康检查间隔(>0开启)
		AutoReconnect       bool          // 自动重连
		QueryTimeout        time.Duration // 查询超时

		Params map[string]string // 额外连接参数

		PgsqlSSLMode  string // SSL模式
		PgsqlTimeZone string // 时区
		SQLiteFile    string // SQLite文件路径 (file::memory:?cache=shared，是内存sqlite的意思)
	}

	// DBInstance 封装数据库实例和相关元数据
	DBInstance struct {
		DB           *gorm.DB
		Name         string
		Config       *DBConfig
		CreatedAt    time.Time
		LastPingTime time.Time
		Healthy      bool
	}
)

// InitConnect 初始化数据库连接
func InitConnect(name string, config DBConfig) (*gorm.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if _, exists := dbInstances[name]; exists {
		return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库连接已存在: %s", name)
	}

	// 根据数据库类型创建对应的方言
	dialector, err := createDialector(config)
	if err != nil {
		return nil, err
	}
	config.Logger.Info(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 连接数据库方言:%s :%s", name, dialector))

	// 配置GORM
	gormConfig := createGormConfig(config)

	// 重试连接逻辑
	db, err := connectWithRetries(dialector, gormConfig, config.MaxRetries, time.Duration(config.RetryDelay)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库连接失败: %s, %w", name, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	config.Logger.Info(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 连接数据库成功:%s", name))

	// 设置连接池参数
	configureConnectionPool(sqlDB, config)

	// 保存到实例map中
	now := time.Now()
	dbInstances[name] = &DBInstance{
		DB:           db,
		Name:         name,
		Config:       &config,
		CreatedAt:    now,
		LastPingTime: now,
		Healthy:      true,
	}

	// 如果启用了健康检查，开始后台健康监控
	if config.HealthCheckInterval > 0 {
		go startHealthCheck(name, config.HealthCheckInterval, config.AutoReconnect)
	}
	return db, nil
}

// 创建数据库方言
func createDialector(config DBConfig) (gorm.Dialector, error) {
	switch config.Kind {
	case DBKindMySQL:
		dsn := buildMySQLDSN(config)
		return mysql.New(mysql.Config{
			DSN:                       dsn,
			SkipInitializeWithVersion: false, // 自动适配 MySQL 版本特性
			DisableWithReturning:      false, // 保留RETURNING子句以提高效率
			DisableDatetimePrecision:  false, // 自动解析时间
			//DefaultStringSize:         256,   // 默认值
		}), nil
	case DBKindPgSQL:
		dsn := buildPgSQLDSN(config)
		return postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: false, // 保持为false以获得更好的预处理语句性能
			WithoutReturning:     false, // 保留RETURNING子句以提高效率
			//WithoutQuotingCheck:  false, // 需要手动处理标识符时设为true
			//DriverName:        "pgx",  // 高性能场景可考虑使用pgx驱动
		}), nil
	case DBKindSQLite:
		if len(config.SQLiteFile) == 0 {
			return nil, fmt.Errorf("■ ■ Storage ■ ■ SQLite数据库文件路径不能为空")
		}
		return sqlite.Open(config.SQLiteFile), nil
	default:
		return nil, fmt.Errorf("■ ■ Storage ■ ■ 不支持的数据库类型: %s", config.Kind)
	}
}

// 构建MySQL DSN
func buildMySQLDSN(config DBConfig) string {
	base := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	// 添加额外参数
	for k, v := range config.Params {
		base += fmt.Sprintf("&%s=%s", k, v)
	}
	return base
}

// 构建PostgreSQL DSN
func buildPgSQLDSN(config DBConfig) string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	if len(config.PgsqlSSLMode) > 0 {
		dsn += fmt.Sprintf(" sslmode=%s", config.PgsqlSSLMode)
	}
	if len(config.PgsqlTimeZone) > 0 {
		dsn += fmt.Sprintf(" TimeZone=%s", config.PgsqlTimeZone)
	}

	// 添加额外参数
	for k, v := range config.Params {
		dsn += fmt.Sprintf(" %s=%s", k, v)
	}
	return dsn
}

// 创建GORM配置
func createGormConfig(config DBConfig) *gorm.Config {
	if config.Logger == nil {
		config.Logger = logger.Default.LogMode(logger.Warn)
	}

	return &gorm.Config{
		Logger:  config.Logger,
		NowFunc: func() time.Time { return time.Now().Local() },

		DisableAutomaticPing:     false, // 初始化后，自动ping数据库
		SkipDefaultTransaction:   true,  // 跳过默认事务，提升性能
		DisableNestedTransaction: false, // 打开嵌套事务
		AllowGlobalUpdate:        false, // 不允许进行全局update/delete
		PrepareStmt:              true,  // 缓存预编译语句以提高性能
		QueryFields:              false, // 允许 SELECT *

		DisableForeignKeyConstraintWhenMigrating: true, // (自动迁移/创建表)不要自动创建外键
		IgnoreRelationshipsWhenMigrating:         true, // (自动迁移/创建表)忽略关系
		//DryRun: false, // 是否启用干运行模式
		//NamingStrategy: schema.NamingStrategy{},
		//FullSaveAssociations
		//PropagateUnscoped:
	}
}

// 连接重试
func connectWithRetries(dialector gorm.Dialector, config *gorm.Config, maxRetries int, retryInterval time.Duration) (*gorm.DB, error) {
	ctx := context.Background()
	var db *gorm.DB
	var err error

	if maxRetries <= 0 {
		maxRetries = 3 // 默认重试3次
	}
	if retryInterval <= 0 {
		retryInterval = 2 * time.Second // 默认重试间隔2秒
	}

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(dialector, config)
		if err == nil && db != nil {
			return db, nil
		}
		config.Logger.Warn(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 连接数据库重试(%d/%d):%s\n%v", i+1, maxRetries, dialector, err))

		select {
		case <-time.After(retryInterval):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	config.Logger.Error(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 连接数据库重试用尽(%d):%s\n%v", maxRetries, dialector, err))

	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库连接失败: 达到最大重试次数")
}

// 配置连接池
func configureConnectionPool(sqlDB *sql.DB, config DBConfig) {
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetConnMaxLifetime(config.MaxLifeTime)
	sqlDB.SetConnMaxIdleTime(config.MaxIdleTime)
}

// 启动健康检查
func startHealthCheck(dbName string, interval time.Duration, autoReconnect bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		instance := GetDBInstance(dbName)
		if instance == nil {
			// 实例已被删除，停止健康检查
			return
		}
		instance.Config.Logger.Info(
			context.Background(),
			fmt.Sprintf("■ ■ Storage ■ ■ 检查数据库健康: %s, %d, %s, %s",
				instance.Config.Host, instance.Config.Port,
				instance.Config.User, instance.Config.DBName))

		err := Ping(dbName)
		dbMutex.Lock()
		now := time.Now()

		if err != nil {
			instance.Healthy = false
			if autoReconnect {
				instance.Config.Logger.Error(
					context.Background(),
					fmt.Sprintf("■ ■ Storage ■ ■ 不健康数据库: %s, %d, %s, %s,\n %v",
						instance.Config.Host, instance.Config.Port,
						instance.Config.User, instance.Config.DBName, err))

				// 尝试重新连接
				sqlDB, _ := instance.DB.DB()
				if sqlDB != nil {
					_ = sqlDB.Close()
				}

				dialector, err := createDialector(*instance.Config)
				if err != nil {
					// 理论上不会走到这步
					return
				}

				// 使用原始配置重新连接
				db, reconnectErr := connectWithRetries(
					dialector,
					createGormConfig(*instance.Config),
					instance.Config.MaxRetries,
					time.Duration(instance.Config.RetryDelay)*time.Second,
				)

				if reconnectErr == nil {
					sqlDB, _ := db.DB()
					if sqlDB != nil {
						configureConnectionPool(sqlDB, *instance.Config)
						instance.DB = db
						instance.Healthy = true
						instance.LastPingTime = now
					}
				}

				instance.Config.Logger.Error(
					context.Background(),
					fmt.Sprintf("■ ■ Storage ■ ■ 不健康数据库检查失败: %s, %d, %s, %s,\n %v",
						instance.Config.Host, instance.Config.Port,
						instance.Config.User, instance.Config.DBName, reconnectErr))
			}
		} else {
			instance.Healthy = true
			instance.LastPingTime = now

			instance.Config.Logger.Info(
				context.Background(),
				fmt.Sprintf("■ ■ Storage ■ ■ 健康数据库: %s, %d, %s, %s",
					instance.Config.Host, instance.Config.Port,
					instance.Config.User, instance.Config.DBName))
		}
		dbMutex.Unlock()
	}
}

// GetDBInstance 获取数据库实例信息
func GetDBInstance(name string) *DBInstance {
	dbMutex.RLock()
	defer dbMutex.RUnlock()

	instance, exists := dbInstances[name]
	if !exists {
		return nil
	}
	return instance
}

// GetDB 通过名称获取数据库连接
func GetDB(name string) *gorm.DB {
	instance := GetDBInstance(name)
	if instance == nil {
		return nil
	}
	return instance.DB
}

// CloseDB 关闭指定的数据库连接
func CloseDB(name string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	instance, exists := dbInstances[name]
	if !exists || instance == nil {
		return fmt.Errorf("■ ■ Storage ■ ■ 关闭: 未找到指定的数据库连接:%s", name)
	}

	sqlDB, err := instance.DB.DB()
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
	for name, instance := range dbInstances {
		if instance == nil || instance.DB == nil {
			errs = append(errs, fmt.Errorf("■ ■ Storage ■ ■ 关闭: 无效的数据库连接:%s", name))
			continue
		}

		sqlDB, err := instance.DB.DB()
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
	instance := GetDBInstance(name)
	if instance == nil {
		return fmt.Errorf("■ ■ Storage ■ ■ Ping: 未找到指定的数据库连接:%s", name)
	}

	sqlDB, err := instance.DB.DB()
	if err != nil {
		return err
	}

	// 使用带超时的上下文进行ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}

// Stats 获取数据库连接池统计信息
func Stats(name string) (sql.DBStats, error) {
	instance := GetDBInstance(name)
	if instance == nil {
		return sql.DBStats{}, fmt.Errorf("■ ■ Storage ■ ■ Stats: 未找到指定的数据库连接:%s", name)
	}

	sqlDB, err := instance.DB.DB()
	if err != nil {
		return sql.DBStats{}, err
	}
	return sqlDB.Stats(), nil
}
