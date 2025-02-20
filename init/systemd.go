package init

import (
	err2 "katydid-mp-user/assets/err"
	"katydid-mp-user/configs"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
	"time"
)

func init() {
	// configs
	config, e := configs.Init(
		configs.ConfDir,
		func() bool {
			// TODO:GG reload
			return true
		},
	)
	if e != nil {
		log.Fatal("■ ■ Init ■ ■ init config failed", log.Err(e))
	}

	// logger
	log.Init(log.Config{
		OutEnable: config.LogConf.OutEnable,
		OutDir:    configs.LogDir,
		OutLevel:  config.LogConf.OutLevel,
		OutFormat: config.LogConf.OutFormat,
		CheckInt:  time.Duration(config.LogConf.CheckInterval) * time.Minute,
		MaxAge:    time.Duration(config.LogConf.FileMaxAge) * time.Hour * 24,
		MaxSize:   config.LogConf.FileMaxSize << 20,
	})

	// error
	err.Init(err2.CodeMsgIds, err2.MsgPatterns)

	// i18n
	e = i18n.Init(i18n.Config{
		DefaultLang: config.DefLang,
		DocDirs:     configs.LangDirs,
		OnErr: func(msg string, fields map[string]any) {
			log.Error(msg, log.Any("fields", fields))
		},
	})
	if e != nil {
		log.Fatal("■ ■ Init ■ ■ init i18n failed", log.Err(e))
	}
}
