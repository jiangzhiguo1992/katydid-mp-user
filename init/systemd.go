package init

import (
	"katydid-mp-user/configs"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
	"time"
)

func init() {
	// configs
	config := configs.Init(
		configs.ConfDir,
		func() bool {
			// TODO:GG reload
			return true
		},
	)

	// logger
	log.Init(log.Config{
		OutEnable: config.LogConf.OutEnable,
		OutDir:    configs.LogDir,
		OutLevel:  config.LogConf.OutLevel,
		OutFormat: config.LogConf.OutFormat,
		CheckInt:  time.Duration(config.LogConf.FileCheckInterval) * time.Minute,
		MaxAge:    time.Duration(config.LogConf.FileMaxAge) * time.Hour * 24,
		MaxSize:   config.LogConf.FileMaxSize << 20,
	})

	// i18n
	i18n.Init(
		configs.LangDirs,
		&config.DefLang,
	)
}