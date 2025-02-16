package init

import (
	"katydid-mp-user/configs"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
)

func init() {
	// configs
	config := configs.Init(configs.ConfDir, func() bool {
		// TODO:GG reload
		return true
	})

	// logger
	log.Init(config.IsProd(), configs.LogDir, &config.LogFileLevel, &config.LogFileNameFormat)

	// i18n
	i18n.Init(configs.LangDirs, &config.DefLang)
}

// TODO:GG 放哪里?
func destroy() {
	log.OnExit()
}
