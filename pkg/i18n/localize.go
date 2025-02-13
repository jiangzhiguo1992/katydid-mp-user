package i18n

import (
	"encoding/json"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert/yaml"
	"golang.org/x/text/language"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	defaultLang string

	localizes = make(map[string]*i18n.Localizer)
)

func Init(dirs []string, defLang string) {
	// default
	defaultLang = defLang
	tag, err := language.Parse(defLang)
	if err != nil {
		panic(err)
	}
	log.Printf("i18n default: %v", tag)

	// bundle
	bundle := i18n.NewBundle(tag)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.RegisterUnmarshalFunc("json", yaml.Unmarshal)

	// files
	var files []string
	for _, dir := range dirs {
		fs, err := filepath.Glob(filepath.Join(dir, "*"))
		if err != nil {
			panic(err)
		}
		for _, f := range fs {
			if fi, err := os.Stat(f); err == nil && fi.IsDir() {
				continue
			}
			files = append(files, f)
		}
	}
	log.Printf("i18n files: %v", files)

	// load files
	var langs []string
	for _, file := range files {
		_, err = bundle.LoadMessageFile(file)
		if err != nil {
			panic(err)
		}
		filename := filepath.Base(file)
		name := filename[:len(filename)-len(filepath.Ext(filename))]
		nameParts := strings.Split(filename, ".")
		if len(nameParts) > 2 {
			name = nameParts[1]
		}
		if localizes[name] == nil {
			localizes[name] = i18n.NewLocalizer(bundle, name)
		}
		langs = append(langs, name)
	}
	log.Printf("i18n langs: %v", langs)
}

func tryLocalizesKey(lang string) string {
	if localizes[lang] != nil {
		return lang
	}
	nameParts := strings.Split(lang, "-")
	if len(nameParts) <= 1 {
		return defaultLang
	}
	if localizes[nameParts[0]] != nil {
		return nameParts[0]
	}
	return defaultLang
}

func Localize(lang, msgID string, datas map[string]interface{}) string {
	lang = tryLocalizesKey(lang)
	return localizes[lang].MustLocalize(&i18n.LocalizeConfig{
		MessageID: msgID,
		//PluralCount:  1,
		TemplateData: datas,
	})
}

func LocalizeTry(lang, msgID string, datas map[string]interface{}) string {
	lang = tryLocalizesKey(lang)
	msg, _ := localizes[lang].Localize(&i18n.LocalizeConfig{
		MessageID: msgID,
		//PluralCount:  1,
		TemplateData: datas,
	})
	return msg
}

func LocalizeDef(msgID string, datas map[string]interface{}) string {
	return Localize(defaultLang, msgID, datas)
}

func LocalizeDefTry(msgID string, datas map[string]interface{}) string {
	return LocalizeTry(defaultLang, msgID, datas)
}
