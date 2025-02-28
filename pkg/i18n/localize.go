package i18n

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	defaultLang  = "en"
	unknownError = "localize unknown error"
)

var defaultManager *Manager

type Config struct {
	DefaultLang string
	DocDirs     []string
	OnErr       func(string, map[string]any)
}

type Manager struct {
	config     Config
	langs      []string
	bundle     *i18n.Bundle
	localizer  sync.Map // 缓存常用localizer对象
	localizers sync.Map // 缓存常用localizers对象
}

func Init(cfg Config) error {
	m, err := newManager(cfg)
	if err != nil {
		return fmt.Errorf("■ ■ i18n ■ ■ create i18n manager failed: %w", err)
	}
	defaultManager = m
	return nil
}

// HasLang 检查语言是否支持
func HasLang(lang string) bool {
	if defaultManager == nil {
		return false
	}
	for _, e := range defaultManager.langs {
		if lang == e {
			return true
		}
	}
	return false
}

// DefLang 获取默认语言
func DefLang() string {
	if defaultManager == nil {
		return ""
	}
	return defaultManager.config.DefaultLang
}

// GetSupportedLangs 获取所有支持的语言
func GetSupportedLangs() []string {
	if defaultManager == nil {
		return nil
	}
	result := make([]string, len(defaultManager.langs))
	copy(result, defaultManager.langs)
	return result
}

func newManager(cfg Config) (*Manager, error) {
	if cfg.DefaultLang == "" {
		cfg.DefaultLang = defaultLang
	}

	defaultTag, err := language.Parse(cfg.DefaultLang)
	if err != nil {
		return nil, fmt.Errorf("■ ■ i18n ■ ■ parse default language failed: %w", err)
	}

	m := &Manager{
		config: cfg,
		langs:  nil,
		bundle: i18n.NewBundle(defaultTag),
	}

	m.registerUnmarshalFuncs()

	if err = m.loadMessageFiles(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) registerUnmarshalFuncs() {
	m.bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	m.bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	m.bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	m.bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
}

func (m *Manager) loadMessageFiles() error {
	var files []string
	for _, dir := range m.config.DocDirs {
		fs, err := filepath.Glob(filepath.Join(dir, "*"))
		if err != nil {
			return fmt.Errorf("■ ■ i18n ■ ■ glob dir %s failed: %w", dir, err)
		}
		files = append(files, filterMessageFiles(fs)...)
	}
	if len(files) == 0 {
		return fmt.Errorf("■ ■ i18n ■ ■ no message files found in dirs: %v", m.config.DocDirs)
	}
	slog.Info("■ ■ i18n ■ ■ loading i18n files", slog.Any("files", files))

	// 加载文件并提取消息ID
	m.langs = make([]string, 0, len(files))
	for _, file := range files {
		if _, err := m.bundle.LoadMessageFile(file); err != nil {
			return fmt.Errorf("■ ■ i18n ■ ■ load message file %s failed: %w", file, err)
		}
		m.langs = append(m.langs, extractLangFromFilename(file))
	}
	slog.Info("■ ■ i18n ■ ■ loading i18n langs", slog.Any("languages", m.langs))

	// 确保默认语言存在
	hasDefault := false
	for _, lang := range m.langs {
		if lang == m.config.DefaultLang {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		slog.Warn("■ ■ i18n ■ ■ default language %q not found in message files", slog.String("lang", m.config.DefaultLang))
	}

	// 预缓存默认Localizer
	m.getLocalizers(m.config.DefaultLang)

	return nil
}

func filterMessageFiles(files []string) []string {
	var result []string
	for _, f := range files {
		if fi, err := os.Stat(f); err == nil && !fi.IsDir() {
			result = append(result, f)
		}
	}
	return result
}

func extractLangFromFilename(file string) string {
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	if parts := strings.Split(name, "."); len(parts) > 1 {
		return parts[1]
	}
	return name
}

// getLocalizers 获取缓存的localizer，如果不存在则创建
func (m *Manager) getLocalizers(lang string) []*i18n.Localizer {
	if v, ok := m.localizers.Load(lang); ok {
		return v.([]*i18n.Localizer)
	}

	// 构建语言标签列表，实现回退链
	var tags []string

	if HasLang(lang) {
		tags = append(tags, lang)
	}

	if base := strings.Split(lang, "-")[0]; base != lang && HasLang(base) {
		tags = append(tags, base)
	}

	tags = append(tags, m.config.DefaultLang)

	localizers := make([]*i18n.Localizer, len(tags))
	for i := 0; i < len(tags); i++ {
		tt := strings.Join(tags[i:], "_")
		if v, ok := m.localizer.Load(tt); ok {
			localizers = append(localizers, v.(*i18n.Localizer))
			continue
		}
		localizer := i18n.NewLocalizer(m.bundle, tags[i:]...)
		m.localizer.Store(tt, localizer)
		localizers = append(localizers, localizer)
	}
	m.localizers.Store(lang, localizers)
	return localizers
}

func (m *Manager) localize(lang, msgID string, data map[string]any, nilBackId bool) string {
	// 构建语言标签列表，实现回退链
	localizers := m.getLocalizers(lang)

	// 循环开找
	var msg string
	var err error
	for _, localizer := range localizers {
		msg, err = localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    msgID,
			TemplateData: data,
			DefaultMessage: &i18n.Message{
				ID: msgID,
				//Other: unknownError,
			},
		})
		if (len(msg) > 0) && (err == nil) {
			break
		}
	}

	// default
	if (len(msg) == 0) && (err != nil) {
		if nilBackId {
			msg = msgID
		} else {
			msg = unknownError
		}
	}

	// error
	if !nilBackId && (err != nil) && (m.config.OnErr != nil) {
		m.config.OnErr("localize failed", map[string]any{
			"msgID": msgID, "lang": lang, "error": err,
		})
	}
	return msg
}

func LocalizeMust(lang, msgID string, data map[string]any) string {
	if defaultManager == nil {
		return unknownError
	}
	return defaultManager.localize(lang, msgID, data, false)
}

func LocalizeTry(lang, msgID string, data map[string]any) string {
	if defaultManager == nil {
		return unknownError
	}
	return defaultManager.localize(lang, msgID, data, true)
}

func LocalizeMustDef(msgID string, data map[string]any) string {
	if defaultManager == nil {
		return unknownError
	}
	return defaultManager.localize(defaultManager.config.DefaultLang, msgID, data, false)
}

func LocalizeTryDef(msgID string, data map[string]any) string {
	if defaultManager == nil {
		return unknownError
	}
	return defaultManager.localize(defaultManager.config.DefaultLang, msgID, data, true)
}
