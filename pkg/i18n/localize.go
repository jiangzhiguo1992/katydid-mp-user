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
)

const (
	defaultDefLang = "en"
	unknownError   = "unknown error"
)

type Config struct {
	DefaultLang string
	DocDirs     []string
	OnErr       func(string, map[string]interface{})
}

type Manager struct {
	config Config
	bundle *i18n.Bundle
}

type Logger interface {
	Error(msg string, fields ...any)
}

var defaultManager *Manager

func Init(cfg Config) error {
	m, err := NewManager(cfg)
	if err != nil {
		return fmt.Errorf("create i18n manager failed: %w", err)
	}
	defaultManager = m
	return nil
}

func NewManager(cfg Config) (*Manager, error) {
	if cfg.DefaultLang == "" {
		cfg.DefaultLang = defaultDefLang
	}

	defaultTag, err := language.Parse(cfg.DefaultLang)
	if err != nil {
		return nil, fmt.Errorf("parse default language failed: %w", err)
	}

	m := &Manager{
		config: cfg,
		bundle: i18n.NewBundle(defaultTag),
	}

	m.registerUnmarshalFuncs()

	if err := m.loadMessageFiles(); err != nil {
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
			return fmt.Errorf("glob dir %s failed: %w", dir, err)
		}
		files = append(files, filterMessageFiles(fs)...)
	}

	if len(files) == 0 {
		return fmt.Errorf("no message files found in dirs: %v", m.config.DocDirs)
	}
	fmt.Printf("loading i18n files: %v", files)

	langs := make([]string, 0, len(files))
	for _, file := range files {
		if _, err := m.bundle.LoadMessageFile(file); err != nil {
			return fmt.Errorf("load message file %s failed: %w", file, err)
		}
		langs = append(langs, extractLangFromFilename(file))
	}

	fmt.Printf("loading initialized languages: %v", langs)
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

func (m *Manager) getLocalizer(lang string) *i18n.Localizer {
	// 构建语言标签列表，实现回退链
	tags := []string{
		lang,                        // 完整标签 (如 zh-CN)
		strings.Split(lang, "-")[0], // 基础标签 (如 zh)
		m.config.DefaultLang,        // 默认语言
	}

	return i18n.NewLocalizer(m.bundle, tags...)
}

func (m *Manager) Localize(lang, msgID string, data map[string]interface{}) string {
	localizer := m.getLocalizer(lang)

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: data,
		DefaultMessage: &i18n.Message{
			ID:    msgID,
			Other: unknownError,
		},
	})

	if err != nil && m.config.OnErr != nil {
		m.config.OnErr("localize failed", map[string]interface{}{
			"msgID": msgID, "lang": lang, "error": err,
		})
	}
	return msg
}

func Localize(lang, msgID string, data map[string]interface{}) string {
	return defaultManager.Localize(lang, msgID, data)
}

func LocalizeDef(msgID string, data map[string]interface{}) string {
	return defaultManager.Localize(defaultManager.config.DefaultLang, msgID, data)
}