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
	logDir := configs.LogDir
	checkInt := time.Duration(config.LogConf.FileCheckInterval) * time.Hour
	maxAge := time.Duration(config.LogConf.FileMaxAge) * time.Hour * 24
	maxSize := int64(config.LogConf.FileMaxSize * 1024 * 1024)
	log.Init(log.Config{
		OutEnable: config.LogConf.OutEnable,
		OutDir:    &logDir,
		OutLevel:  &config.LogConf.OutLevel,
		OutFormat: &config.LogConf.OutFormat,
		CheckInt:  &checkInt,
		MaxAge:    &maxAge,
		MaxSize:   &maxSize,
	})

	// i18n
	i18n.Init(
		configs.LangDirs,
		&config.DefLang,
	)
}