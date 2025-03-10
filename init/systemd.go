package init

import (
	"katydid-mp-user/configs"
	"katydid-mp-user/internal/pkg/msg"
	"katydid-mp-user/pkg/errs"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/pkg/perm"
	"time"
)

func init() {
	// configs
	config, e := configs.Init(configs.ConfDir)
	if e != nil {
		panic(e)
	}
	//configs.Subscribe("account.enable", func(v any) {
	//	slog.Info("reload account.enable", slog.Any("v", v))
	//}) // TODO:GG 订阅reload

	// logger
	log.Init(log.Config{
		OutDir:    configs.LogDir,
		OutLevel:  config.LogConf.OutLevel,
		OutFormat: config.LogConf.OutFormat,
		CheckInt:  time.Duration(config.LogConf.CheckInterval) * time.Minute,
		MaxAge:    time.Duration(config.LogConf.FileMaxAge) * time.Hour * 24,
		MaxSize:   config.LogConf.FileMaxSize << 20,
	})

	// i18n
	e = i18n.Init(i18n.Config{
		DefaultLang: config.DefLang,
		DocDirs:     configs.LangDirs,
		OnErr: func(msg string, fields map[string]any) {
			log.Error(msg, log.FAny("fields", fields))
		},
	})
	if e != nil {
		log.Fatal("", log.FError(e))
	}

	// error
	errs.Init(msg.ErrCodePatterns, msg.ErrMsgPatterns, func(msg string) {
		log.Error(msg)
	})

	// perm
	perm.Init(&perm.Config{
		LogEnable: true,
	})
}
