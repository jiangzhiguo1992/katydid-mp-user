package init

import (
	"katydid-mp-user/configs"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
	"time"
)

func init() {
	// configs
	config, err := configs.Init(
		configs.ConfDir,
		func() bool {
			// TODO:GG reload
			return true
		},
	)
	if err != nil {
		log.Fatal("init config failed", log.Err(err))
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

	// i18n
	err = i18n.Init(i18n.Config{
		DefaultLang: config.DefLang,
		DocDirs:     configs.LangDirs,
		OnErr: func(msg string, fields map[string]interface{}) {
			log.Error(msg, log.Any("fields", fields))
		},
	})
	if err != nil {
		log.Fatal("init i18n failed", log.Err(err))
	}
}
