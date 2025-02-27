package init

import (
	"katydid-mp-user/configs"
	"katydid-mp-user/internal/pkg/text"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/log"
	"time"
)

func init() {
	// configs
	config, e := configs.Init(configs.ConfDir)
	if e != nil {
		log.Fatal("", log.FError(e))
	}
	//configs.Subscribe() // TODO:GG 订阅reload

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
	err.Init(text.CodeMsgIds, text.MsgPatterns, func(msg string) {
		log.Error(msg)
	})
}
