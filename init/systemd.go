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
			var fs []log.Field
			if fields != nil {
				for k, v := range fields {
					fs = append(fs, log.FAny(k, v))
				}
			}
			log.InfoMust(!config.IsDebug(), msg, fs...)
		},
		OnErr: func(msg string, fields map[string]any) {
			var fs []log.Field
			if fields != nil {
				for k, v := range fields {
					fs = append(fs, log.FAny(k, v))
				}
			}
			log.ErrorMust(!config.IsDebug(), msg, fs...)
		},
	})
	if err != nil {
		log.FatalMust(!config.IsDebug(), "i18n", log.FError(err))
	}

	// error
	errs.Init(msg.ErrCodePatterns, msg.ErrMsgPatterns, func(msg string) {
		log.WarnFmtMust(!config.IsDebug(), msg)
	})

	// TODO:GG role
	//perm.Init(&perm.Config{
	//	LogEnable: true,
	//})

	// pgsql
	if PgSql := config.PgSql; PgSql != nil {
		dbConfig := storage.DBConfig{
			Kind:   storage.DBKindPgSQL,
			Logger: &StoreLogger{!config.IsDebug()},
			// db
			Host:     PgSql.Write.Host,
			Port:     PgSql.Write.Port,
			DBName:   PgSql.Write.DBName,
			User:     PgSql.Write.User,
			Password: PgSql.Write.Pwd,
			// cluster
			OnlyMaster: true,
			Replicas:   make([]storage.ReplicaConfig, 0),
			// retry
			MaxRetries:    PgSql.MaxRetries,
			RetryDelay:    PgSql.RetryDelay,
			RetryMaxDelay: PgSql.RetryMaxDelay,
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
			Params: PgSql.Params,
		}
		if PgSql.Reads != nil {
			for index, host := range PgSql.Reads.Host {
				if len(host) <= 0 {
					continue
				}
				port := 0
				if len(PgSql.Reads.Port) > index {
					port = PgSql.Reads.Port[index]
				} else if len(PgSql.Reads.Port) > 0 {
					port = PgSql.Reads.Port[0]
				}
				user := ""
				if len(PgSql.Reads.User) > index {
					user = PgSql.Reads.User[index]
				} else if len(PgSql.Reads.User) > 0 {
					user = PgSql.Reads.User[0]
				}
				pwd := ""
				if len(PgSql.Reads.Pwd) > index {
					pwd = PgSql.Reads.Pwd[index]
				} else if len(PgSql.Reads.Pwd) > 0 {
					pwd = PgSql.Reads.Pwd[0]
				}
				weight := 0
				if len(PgSql.Reads.Weight) > index {
					weight = PgSql.Reads.Weight[index]
				} else if len(PgSql.Reads.Weight) > 0 {
					weight = PgSql.Reads.Weight[0]
				}
				params := map[string]string{}
				if len(PgSql.Reads.Params) > index {
					params = PgSql.Reads.Params[index]
				} else if len(PgSql.Reads.Params) > 0 {
					params = PgSql.Reads.Params[0]
				}

				dbConfig.Replicas = append(dbConfig.Replicas, storage.ReplicaConfig{
					Host: host, Port: port, User: user, Password: pwd,
					Weight: weight, Params: params,
				})
			}
			dbConfig.OnlyMaster = len(dbConfig.Replicas) <= 0
		}
		name := "pgsql.main" // TODO:GG 放哪里
		_, err = storage.InitConnect(name, dbConfig)
		if err != nil {
			log.FatalMust(!config.IsDebug(), fmt.Sprintf("storage %s", name), log.FError(err))
		}
	}
}

// StoreLogger 实现了gorm.Logger接口
type StoreLogger struct {
	output bool
}

func (s *StoreLogger) LogMode(logger.LogLevel) logger.Interface {
	return s
}

func (s *StoreLogger) Info(ctx context.Context, msg string, params ...interface{}) {
	var field []log.Field
	if params != nil {
		field = append(field, log.FAny("params", params))
	}
	log.Info(msg, field...)
}

func (s *StoreLogger) Warn(ctx context.Context, msg string, params ...interface{}) {
	var field []log.Field
	if params != nil {
		field = append(field, log.FAny("params", params))
	}
	log.WarnMust(s.output, msg, field...)
}

func (s *StoreLogger) Error(ctx context.Context, msg string, params ...interface{}) {
	var field []log.Field
	if params != nil {
		field = append(field, log.FAny("params", params))
	}
	log.ErrorMust(s.output, msg, field...)
}

func (s *StoreLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	// 构建日志字段
	fields := []log.Field{
		log.FDuration("耗时", elapsed),
		log.FString("语句", sql),
		log.FInt64("结果", rows),
	}

	// 根据错误状态和执行时间选择日志级别
	if err != nil {
		fields = append(fields, log.FError(err))
		log.ErrorMust(s.output, "SQL执行失败", fields...)
	} else if elapsed > time.Second {
		// 慢查询警告阈值， TODO:GG 需要调整
		log.WarnMust(s.output, "SQL执行过慢", fields...)
	} else {
		// 正常查询可以选择记录为Debug级别或Info级别
		log.Debug("SQL执行成功", fields...)
	}
}
