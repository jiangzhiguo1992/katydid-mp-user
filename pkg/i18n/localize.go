package i18n

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
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
	OnInfo      func(string, map[string]any) `json:"-"`
	OnErr       func(string, map[string]any) `json:"-"`
}

type Manager struct {
	config     Config
	bundle     *i18n.Bundle
	langs      []string
	langsMap   map[string]bool // 用于快速查找语言是否支持
	localizer  sync.Map        // 缓存常用localizer对象
	localizers sync.Map        // 缓存常用localizers对象
}

func Init(cfg Config) error {
	marshal, _ := json.MarshalIndent(cfg, "", "\t")
	cfg.OnInfo(fmt.Sprintf("■ ■ i18n ■ ■ 配置 ---> %s", marshal), nil)

	m, err := newManager(cfg)
	if err != nil {
		return fmt.Errorf("■ ■ i18n ■ ■ 创建 i18n 失败: %w", err)
	}
	defaultManager = m
	return nil
}

// HasLang 检查语言是否支持
func HasLang(lang string) bool {
	if defaultManager == nil {
		return false
	}
	return defaultManager.langsMap[lang]
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
		return nil, fmt.Errorf("■ ■ i18n ■ ■ 解析默认失败: %w", err)
	}

	m := &Manager{
		config:   cfg,
		bundle:   i18n.NewBundle(defaultTag),
		langs:    nil,
		langsMap: make(map[string]bool),
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
			return fmt.Errorf("■ ■ i18n ■ ■ 加载文件夹 %s 失败: %w", dir, err)
		}
		files = append(files, filterMessageFiles(fs)...)
	}
	if len(files) == 0 {
		return fmt.Errorf("■ ■ i18n ■ ■ 没有发现文件: %v", m.config.DocDirs)
	}
	marshal, _ := json.MarshalIndent(map[string]any{"files": files}, "", "\t")
	m.config.OnInfo(fmt.Sprintf("■ ■ i18n ■ ■ 本地化加载文件: %s", marshal), nil)

	// 加载文件并提取消息ID
	m.langs = make([]string, 0, len(files))
	for _, file := range files {
		if _, err := m.bundle.LoadMessageFile(file); err != nil {
			return fmt.Errorf("■ ■ i18n ■ ■ 加载文件 %s 失败: %w", file, err)
		}

		lang := extractLangFromFilename(file)
		m.langs = append(m.langs, lang)
		m.langsMap[lang] = true
	}
	marshal2, _ := json.MarshalIndent(map[string]any{"languages": m.langs}, "", "\t")
	m.config.OnInfo(fmt.Sprintf("■ ■ i18n ■ ■ 本地化加载语言: %s", marshal2), nil)

	// 确保默认语言存在
	if !m.langsMap[m.config.DefaultLang] {
		if m.config.OnErr != nil {
			m.config.OnErr("■ ■ i18n ■ ■ 本地化默认文件不存在 ", map[string]any{"lang": m.config.DefaultLang})
		}
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

	if m.langsMap[lang] {
		tags = append(tags, lang)
	}

	// 添加基本语言标签（如 "en" 从 "en-US"）
	if parts := strings.Split(lang, "-"); len(parts) > 1 {
		base := parts[0]
		if base != lang && m.langsMap[base] {
			tags = append(tags, base)
		}
	}

	// 添加默认语言
	if (len(tags) == 0) || (tags[len(tags)-1] != m.config.DefaultLang) {
		tags = append(tags, m.config.DefaultLang)
	}

	localizers := make([]*i18n.Localizer, len(tags))
	for i := 0; i < len(tags); i++ {
		tt := strings.Join(tags[i:], "_")
		if v, ok := m.localizer.Load(tt); ok {
			localizers[i] = v.(*i18n.Localizer)
			continue
		}
		localizer := i18n.NewLocalizer(m.bundle, tags[i:]...)
		m.localizer.Store(tt, localizer)
		localizers[i] = localizer
	}
	m.localizers.Store(lang, localizers)
	return localizers
}

func (m *Manager) localize(lang, msgID string, data map[string]any, fallbackToID bool) string {
	// 构建语言标签列表，实现回退链
	localizers := m.getLocalizers(lang)

	// 循环查找
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

	// 处理未找到消息的情况
	if (len(msg) == 0) && (err != nil) {
		if fallbackToID {
			msg = msgID
		} else {
			msg = unknownError
		}
	}

	// 记录错误
	if !fallbackToID && (err != nil) && (m.config.OnErr != nil) {
		m.config.OnErr("■ ■ i18n ■ ■ 本地化未找到: ", map[string]any{
			"msgID": msgID, "lang": lang, "error": err})
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
