package init

import (
	"katydid-mp-user/configs"
	"katydid-mp-user/pkg/log"
)

func init() {
	// configs
	config := configs.InitConfig()

	// logger
	log.InitLogger(config.IsProd(), configs.LogDir)
}
