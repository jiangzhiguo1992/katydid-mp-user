package init

import (
	"context"
	"fmt"
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
	// TODO:GG 理论上每个可变动的configs都需要监听，并reset后续
	configs.Subscribe("auth.enable", func(v any) {
		//slog.Info("reload auth.enable", slog.Any("v", v))
	})

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

	// pgsql
	if PgSql := config.PgSql; PgSql != nil {
		// 配置
		getConf := func(host string, port int, dbName, user, pwd string) storage.DBConfig {
			return storage.DBConfig{
				Kind:   storage.DBKindPgSQL,
				Logger: &StoreLogger{},
				// db
				Host: host, Port: port, DBName: dbName, User: user, Password: pwd,
				// retry
				MaxRetries: PgSql.MaxRetries,
				RetryDelay: PgSql.RetryDelay,
				// pool
				MaxOpen:     PgSql.MaxOpen,
				MaxIdle:     PgSql.MaxIdle,
				MaxLifeTime: time.Duration(PgSql.MaxLifeMin) * time.Minute,
				MaxIdleTime: time.Duration(PgSql.MaxIdleMin) * time.Minute,
				// health
				HealthCheckInterval: time.Duration(PgSql.HealthCheckInterval) * time.Minute,
				AutoReconnect:       PgSql.AutoReconnect,
				QueryTimeout:        time.Duration(PgSql.QueryTimeout) * time.Second,
				// extra
				Params: map[string]string{
					"sslmode":  PgSql.SSLMode,
					"TimeZone": PgSql.TimeZone,
				},
			}
		}
		// 主数据库
		write := getConf(PgSql.Write.Host, PgSql.Write.Port,
			PgSql.Write.DBName, PgSql.Write.User, PgSql.Write.Pwd)
		name := "pgsql.main.write"
		_, err = storage.InitConnect(name, write)
		if err != nil {
			log.Fatal(fmt.Sprintf("storage %s", name), log.FError(err))
		}
		// 从数据库
		if PgSql.Read != nil {
			for index, host := range PgSql.Read.Host {
				port := PgSql.Read.Port[0]
				if len(PgSql.Read.Port) > index {
					port = PgSql.Read.Port[index]
				}
				dbName := PgSql.Read.DBName[0]
				if len(PgSql.Read.DBName) > index {
					dbName = PgSql.Read.DBName[index]
				}
				user := PgSql.Read.User[0]
				if len(PgSql.Read.User) > index {
					user = PgSql.Read.User[index]
				}
				pwd := PgSql.Read.Pwd[0]
				if len(PgSql.Read.Pwd) > index {
					pwd = PgSql.Read.Pwd[index]
				}
				read := getConf(host, port, dbName, user, pwd)
				name := fmt.Sprintf("pgsql.main.read.%d", index)
				_, err = storage.InitConnect(name, read)
				if err != nil {
					log.Fatal(fmt.Sprintf("storage %s", name), log.FError(err))
				}
			}
		}
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
