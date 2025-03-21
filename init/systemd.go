package init

import (
	"context"
	"gorm.io/gorm/logger"
	"katydid-mp-user/configs"
	"katydid-mp-user/internal/pkg/msg"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/pkg/storage"
	"time"
)

func init() {
	// configs
	config, err := configs.Init(configs.ConfDir)
	if err != nil {
		panic(err)
	}
	//configs.Subscribe("auth.enable", func(v any) {
	//	slog.Info("reload auth.enable", slog.Any("v", v))
	//}) // TODO:GG 订阅reload

	// logger
	log.Init(log.Config{
		ConLevels: config.LogConf.ConLevels,
		OutLevels: config.LogConf.OutLevels,
		OutDir:    configs.LogDir,
		OutFormat: config.LogConf.OutFormat,
		CheckInt:  time.Duration(config.LogConf.CheckInterval) * time.Minute,
		MaxAge:    time.Duration(config.LogConf.FileMaxAge) * time.Hour * 24,
		MaxSize:   config.LogConf.FileMaxSize << 20,
	})

	// i18n
	err = i18n.Init(i18n.Config{
		DefaultLang: config.LangConf.Default,
		DocDirs:     configs.LangDirs,
		OnInfo: func(msg string, fields map[string]any) {
			if config.IsDebug() {
				log.Info(msg, log.FAny("fields", fields))
			} else {
				log.InfoOutput(msg, true, log.FAny("fields", fields))
			}
		},
		OnErr: func(msg string, fields map[string]any) {
			log.Error(msg, log.FAny("fields", fields))
		},
	})
	if err != nil {
		log.Fatal("i18n", log.FError(err))
	}

	// error
	errs.Init(msg.ErrCodePatterns, msg.ErrMsgPatterns, func(msg string) {
		log.WarnFmt(msg)
	})

	// TODO:GG role
	//perm.Init(&perm.Config{
	//	LogEnable: true,
	//})

	// storage
	_, err = storage.InitConnect("pgsql", storage.DBConfig{
		Kind:   storage.DBKindPgSQL,
		Logger: &StoreLogger{},
		// db
		Host:     config.PgSql.Write.Host,
		Port:     config.PgSql.Write.Port,
		User:     config.PgSql.Write.User,
		Password: config.PgSql.Write.Pwd,
		DBName:   config.PgSql.Write.DBName,
		// retry
		MaxRetries: config.PgSql.MaxRetries,
		RetryDelay: config.PgSql.RetryDelay,
		// pool
		MaxOpen:     config.PgSql.MaxOpen,
		MaxIdle:     config.PgSql.MaxIdle,
		MaxLifeTime: time.Duration(config.PgSql.MaxLifeMin) * time.Minute,
		MaxIdleTime: time.Duration(config.PgSql.MaxIdleMin) * time.Minute,
		// health
		HealthCheckInterval: time.Duration(config.PgSql.HealthCheckInterval) * time.Minute,
		AutoReconnect:       config.PgSql.AutoReconnect,
		QueryTimeout:        time.Duration(config.PgSql.HealthCheckInterval) * time.Minute,
		// extra
		Params: map[string]string{
			"sslmode":  config.PgSql.SSLMode,
			"TimeZone": config.PgSql.TimeZone,
		},
	})
	if err != nil {
		log.Fatal("storage", log.FError(err))
	}
}

type StoreLogger struct{}

func (s *StoreLogger) LogMode(logger.LogLevel) logger.Interface {
	return s
}

func (s *StoreLogger) Info(ctx context.Context, msg string, params ...interface{}) {
	log.Info(msg, log.FAny("params", params))
}

func (s *StoreLogger) Warn(ctx context.Context, msg string, params ...interface{}) {
	log.Warn(msg, log.FAny("params", params))
}

func (s *StoreLogger) Error(ctx context.Context, msg string, params ...interface{}) {
	log.Error(msg, log.FAny("params", params))
}

func (s *StoreLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, _ := fc()
	log.Warn("这他妈是什么错?")
	log.Error(sql, log.FError(err))
}
