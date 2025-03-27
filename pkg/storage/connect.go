package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/goccy/go-json"
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
		// info
		Host     string // 主机地址
		Port     int    // 端口号
		DBName   string // 数据库名称
		User     string // 用户名
		Password string // 密码
		// cluster - 增加集群支持
		OnlyMaster bool            // 始终使用主库（即使有只读查询）
		Replicas   []ReplicaConfig // 只读副本配置
		// retry
		MaxRetries int // 重试次数 (自动纠正<=0)
		RetryDelay int // 重试间隔，单位秒 (自动纠正<=0)
		// pool
		MaxOpen     int           // 最大连接数，一般=1000 (看数据库性能)
		MaxIdle     int           // 最大空闲连接数，一般==Open
		MaxLifeTime time.Duration // 连接最大存活时间，一般=3m (<=0永生)
		MaxIdleTime time.Duration // 空闲连接最大存活时间，一般=1m (<=0永生)
		// health
		HealthCheckInterval time.Duration // 健康检查间隔(>0开启)
		AutoReconnect       bool          // 自动重连
		QueryTimeout        time.Duration // 查询超时
		// extra
		Params     map[string]string // 额外连接参数
		SQLiteFile string            // SQLite文件路径 (file::memory:?cache=shared，是内存sqlite的意思)
	}

	// ReplicaConfig 只读副本配置
	ReplicaConfig struct {
		Host     string            // 副本主机地址
		Port     int               // 副本端口号
		User     string            // 用户名
		Password string            // 密码
		Params   map[string]string // 额外连接参数
		Weight   int               // 权重，用于负载均衡，默认为1
	}

	// DBInstance 封装数据库实例和相关元数据
	DBInstance struct {
		DB           *gorm.DB     // 主数据库连接
		ReadReplicas []*gorm.DB   // 只读副本连接
		Name         string       // 实例名称
		Config       *DBConfig    // 配置信息
		CreatedAt    time.Time    // 创建时间
		LastPingTime time.Time    // 最后一次Ping时间
		Healthy      bool         // 健康状态
		mutex        sync.RWMutex // 用于保护本实例内部并发访问
	}
)

// InitConnect 初始化数据库连接
func InitConnect(name string, config DBConfig) (*gorm.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// 打印配置
	marshal, _ := json.MarshalIndent(config, "", "\t")
	config.Logger.Info(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 数据库-连接配置:%s :\n%s", name, marshal))

	// 检查是否已存在同名连接
	if _, exists := dbInstances[name]; exists {
		return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库-连接已存在: %s", name)
	}

	// 根据数据库类型创建对应的方言
	dialector, err := createDialector(config)
	if err != nil {
		return nil, err
	}
	config.Logger.Info(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 数据库-连接方言:%s :%s", name, dialector))

	// 配置GORM
	gormConfig := createGormConfig(config)

	// 重试连接逻辑
	db, err := connectWithRetries(dialector, gormConfig, config.MaxRetries, time.Duration(config.RetryDelay)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库-连接失败: %s, %w", name, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库-连接获取失败: %s, %w", name, err)
	}
	config.Logger.Info(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 数据库-连接成功:%s", name))

	// 设置连接池参数
	configureConnectionPool(sqlDB, config)

	// 保存到实例map中
	now := time.Now()
	dbInstances[name] = &DBInstance{
		DB:           db,
		ReadReplicas: []*gorm.DB{},
		Name:         name,
		Config:       &config,
		CreatedAt:    now,
		LastPingTime: now,
		Healthy:      true,
	}

	// 初始化副本连接
	if len(config.Replicas) > 0 {
		for i, replica := range config.Replicas {
			replicaConfig := config
			replicaConfig.Host = replica.Host
			replicaConfig.Port = replica.Port
			if replica.User != "" {
				replicaConfig.User = replica.User
			}
			if replica.Password != "" {
				replicaConfig.Password = replica.Password
			}
			if replica.Params != nil {
				replicaConfig.Params = replica.Params
			}

			replicaDialector, err := createDialector(replicaConfig)
			if err != nil {
				config.Logger.Warn(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 数据库-副本连接方言创建失败: %d, %v", i, err))
				continue
			}

			replicaDB, err := connectWithRetries(replicaDialector, gormConfig, config.MaxRetries, time.Duration(config.RetryDelay)*time.Second)
			if err != nil {
				config.Logger.Warn(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 数据库-副本连接失败: %d, %v", i, err))
				continue
			}

			replicaSQLDB, err := replicaDB.DB()
			if err != nil {
				return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库-副本连接获取失败: %d, %s, %w", i, name, err)
			} else {
				configureConnectionPool(replicaSQLDB, replicaConfig)
			}

			dbInstances[name].ReadReplicas = append(dbInstances[name].ReadReplicas, replicaDB)
			config.Logger.Info(context.Background(), fmt.Sprintf("■ ■ Storage ■ ■ 数据库-副本连接成功: %d, %s:%d", i, replica.Host, replica.Port))
		}
	}

	// 如果启用了健康检查，开始后台健康监控
	if config.HealthCheckInterval > 0 {
		go startHealthCheck(name, config.HealthCheckInterval,
			config.AutoReconnect, config.QueryTimeout)
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
			return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库-SQLite文件路径不能为空")
		}
		return sqlite.Open(config.SQLiteFile), nil
	default:
		return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库-不支持的类型: %s", config.Kind)
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
	// 注意params先后顺序
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s",
		config.Host, config.Port, config.DBName, config.User, config.Password)

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
		retryInterval = 3 * time.Second // 默认重试间隔3秒
	}

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(dialector, config)
		if err == nil && db != nil {
			return db, err
		}
		config.Logger.Warn(ctx, fmt.Sprintf("■ ■ Storage ■ ■ 数据库-连接重试(%d/%d):%v\n%v", i+1, maxRetries, dialector, err))

		select {
		case <-time.After(retryInterval):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	config.Logger.Error(ctx, fmt.Sprintf("■ ■ Storage ■ ■ 数据库-连接重试用尽(%d):%v\n%v", maxRetries, dialector, err))

	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("■ ■ Storage ■ ■ 数据库-连接失败: 达到最大重试次数")
}

// 配置连接池
func configureConnectionPool(sqlDB *sql.DB, config DBConfig) {
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetConnMaxLifetime(config.MaxLifeTime)
	sqlDB.SetConnMaxIdleTime(config.MaxIdleTime)
}

// 启动健康检查
func startHealthCheck(name string, interval time.Duration, autoReconnect bool, queryTimeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		instance := GetDBInstance(name)
		if instance == nil {
			// 实例已被删除，停止健康检查
			return
		}
		instance.Config.Logger.Info(
			context.Background(),
			fmt.Sprintf("■ ■ Storage ■ ■ 数据库-健康检查: %s", name))

		// 检查主连接
		err := Ping(name, queryTimeout)
		now := time.Now()
		instance.mutex.Lock()

		if err != nil {
			instance.Healthy = false

			if autoReconnect {
				instance.Config.Logger.Error(
					context.Background(),
					fmt.Sprintf("■ ■ Storage ■ ■ 数据库-健康No: %s\n %v", name, err))

				// 尝试重新连接
				reconnectMainDB(instance, now)
			}
		} else {
			instance.Healthy = true
			instance.LastPingTime = now

			instance.Config.Logger.Info(
				context.Background(),
				fmt.Sprintf("■ ■ Storage ■ ■ 数据库-健康OK: %s", name))
		}

		// 检查副本连接
		for i, replica := range instance.ReadReplicas {
			sqlDB, _ := replica.DB()
			ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
			replicaErr := sqlDB.PingContext(ctx)
			cancel()

			if replicaErr != nil && autoReconnect {
				instance.Config.Logger.Error(
					context.Background(),
					fmt.Sprintf("■ ■ Storage ■ ■ 数据库-副本健康No: %s [%d]\n %v", name, i, replicaErr))

				// 重新连接副本
				reconnectReplica(instance, i)
			} else {
				instance.Config.Logger.Info(
					context.Background(),
					fmt.Sprintf("■ ■ Storage ■ ■ 数据库-副本健康OK: %s [%d]", name, i))
			}
		}
		instance.mutex.Unlock()
	}
}

// 重新连接主数据库
func reconnectMainDB(instance *DBInstance, now time.Time) bool {
	// 关闭旧连接
	sqlDB, _ := instance.DB.DB()
	if sqlDB != nil {
		_ = sqlDB.Close()
	}

	// 创建新连接
	dialector, err := createDialector(*instance.Config)
	if err != nil {
		return false
	}

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
			return true
		}
	}

	instance.Config.Logger.Error(
		context.Background(),
		fmt.Sprintf("■ ■ Storage ■ ■ 数据库-重连失败: %s %v", instance.Name, reconnectErr))
	return false
}

// 重新连接副本数据库
func reconnectReplica(instance *DBInstance, replicaIndex int) bool {
	if replicaIndex >= len(instance.Config.Replicas) || replicaIndex >= len(instance.ReadReplicas) {
		return false
	}

	replica := instance.Config.Replicas[replicaIndex]

	// 创建副本配置
	replicaConfig := *instance.Config
	replicaConfig.Host = replica.Host
	replicaConfig.Port = replica.Port
	if replica.User != "" {
		replicaConfig.User = replica.User
	}
	if replica.Password != "" {
		replicaConfig.Password = replica.Password
	}
	if replica.Params != nil {
		replicaConfig.Params = replica.Params
	}

	// 关闭现有连接
	oldDB := instance.ReadReplicas[replicaIndex]
	if sqlDB, err := oldDB.DB(); err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}

	// 创建新连接
	dialector, err := createDialector(replicaConfig)
	if err != nil {
		return false
	}

	newDB, err := connectWithRetries(
		dialector,
		createGormConfig(replicaConfig),
		replicaConfig.MaxRetries,
		time.Duration(replicaConfig.RetryDelay)*time.Second,
	)

	if err == nil && newDB != nil {
		// 配置连接池
		if sqlDB, err := newDB.DB(); err == nil {
			configureConnectionPool(sqlDB, replicaConfig)
		}

		// 更新实例
		instance.ReadReplicas[replicaIndex] = newDB
		return true
	}

	return false
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

// GetReadDB 获取读库连接（支持负载均衡选择副本）
func GetReadDB(name string) *gorm.DB {
	instance := GetDBInstance(name)
	if instance == nil {
		return nil
	}

	instance.mutex.RLock()
	defer instance.mutex.RUnlock()

	// 如果无副本或配置为始终使用主库，返回主连接
	replicaCount := len(instance.ReadReplicas)
	if replicaCount == 0 || instance.Config.OnlyMaster {
		return instance.DB
	}

	// 简单轮询负载均衡（可优化为加权轮询）
	index := int(time.Now().UnixNano()) % replicaCount
	return instance.ReadReplicas[index]
}

// CloseDB 关闭指定的数据库连接
func CloseDB(name string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	instance, exists := dbInstances[name]
	if !exists || instance == nil {
		return fmt.Errorf("■ ■ Storage ■ ■ 数据库-关闭:未找到指定的连接:%s", name)
	}

	// 关闭主连接
	sqlDB, err := instance.DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()

	// 关闭副本连接
	for _, replica := range instance.ReadReplicas {
		if sqlDB, err := replica.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}

	delete(dbInstances, name)
	return err
}

// CloseAllDBs 关闭所有数据库连接
func CloseAllDBs() []error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	var errs []error
	for name, instance := range dbInstances {
		if instance == nil || instance.DB == nil {
			errs = append(errs, fmt.Errorf("■ ■ Storage ■ ■ 数据库-关闭:无效的连接:%s", name))
			continue
		}

		// 关闭主连接
		sqlDB, err := instance.DB.DB()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if err = sqlDB.Close(); err != nil {
			errs = append(errs, err)
		}

		// 关闭副本连接
		for i, replica := range instance.ReadReplicas {
			if sqlDB, err := replica.DB(); err == nil && sqlDB != nil {
				if err = sqlDB.Close(); err != nil {
					errs = append(errs, fmt.Errorf("关闭副本 %s[%d] 失败: %w", name, i, err))
				}
			}
		}

		delete(dbInstances, name)
	}
	return errs
}

// Ping 检查数据库连接状态
func Ping(name string, timeout time.Duration) error {
	instance := GetDBInstance(name)
	if instance == nil {
		return fmt.Errorf("■ ■ Storage ■ ■ 数据库-Ping:未找到指定的连接:%s", name)
	}

	sqlDB, err := instance.DB.DB()
	if err != nil {
		return err
	}

	// 使用带超时的上下文进行ping
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return sqlDB.PingContext(ctx)
	}

	return sqlDB.Ping()
}

// Stats 获取数据库连接池统计信息
func Stats(name string) (sql.DBStats, error) {
	instance := GetDBInstance(name)
	if instance == nil {
		return sql.DBStats{}, fmt.Errorf("■ ■ Storage ■ ■ 数据库-Stats:未找到指定的连接:%s", name)
	}

	sqlDB, err := instance.DB.DB()
	if err != nil {
		return sql.DBStats{}, err
	}
	return sqlDB.Stats(), nil
}
