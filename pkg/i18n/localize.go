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

// Config i18n 配置
type Config struct {
	DefaultLang string   // 默认语言
	Dirs        []string // 语言文件目录
	Debug       bool     // 调试模式
}

// Manager i18n 管理器
type Manager struct {
	config    Config
	bundle    *i18n.Bundle
	localizes sync.Map // string -> *i18n.Localizer
	logger    Logger
}

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
}

var defaultManager *Manager

// Init 初始化默认管理器
func Init(cfg Config, logger Logger) error {
	m, err := NewManager(cfg, logger)
	if err != nil {
		return fmt.Errorf("create i18n manager failed: %w", err)
	}
	defaultManager = m
	return nil
}

// NewManager 创建新的i18n管理器
func NewManager(cfg Config, logger Logger) (*Manager, error) {
	if cfg.DefaultLang == "" {
		cfg.DefaultLang = "en"
	}

	tag, err := language.Parse(cfg.DefaultLang)
	if err != nil {
		return nil, fmt.Errorf("parse default language failed: %w", err)
	}

	m := &Manager{
		config: cfg,
		logger: logger,
	}

	// 初始化bundle
	m.bundle = i18n.NewBundle(tag)
	m.registerUnmarshalFuncs()

	// 加载语言文件
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
	for _, dir := range m.config.Dirs {
		fs, err := filepath.Glob(filepath.Join(dir, "*"))
		if err != nil {
			return fmt.Errorf("glob dir %s failed: %w", dir, err)
		}
		for _, f := range fs {
			if fi, err := os.Stat(f); err == nil && !fi.IsDir() {
				files = append(files, f)
			}
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no message files found in dirs: %v", m.config.Dirs)
	}

	if m.config.Debug {
		m.logger.Debug("loading i18n files", "files", files)
	}

	var langs []string
	for _, file := range files {
		if err := m.loadSingleFile(file); err != nil {
			return err
		}
		lang := m.extractLangFromFilename(file)
		langs = append(langs, lang)
	}

	m.logger.Info("i18n initialized", "languages", langs)
	return nil
}

func (m *Manager) loadSingleFile(file string) error {
	if _, err := m.bundle.LoadMessageFile(file); err != nil {
		return fmt.Errorf("load message file %s failed: %w", file, err)
	}

	lang := m.extractLangFromFilename(file)
	m.localizes.Store(lang, i18n.NewLocalizer(m.bundle, lang))
	return nil
}

func (m *Manager) extractLangFromFilename(file string) string {
	filename := filepath.Base(file)
	name := filename[:len(filename)-len(filepath.Ext(filename))]
	parts := strings.Split(filename, ".")
	if len(parts) > 2 {
		name = parts[1]
	}
	return name
}

func (m *Manager) getLocalizer(lang string) *i18n.Localizer {
	// 尝试完整语言代码
	if loc, ok := m.localizes.Load(lang); ok {
		return loc.(*i18n.Localizer)
	}

	// 尝试语言基础部分
	parts := strings.Split(lang, "-")
	if len(parts) > 1 {
		if loc, ok := m.localizes.Load(parts[0]); ok {
			return loc.(*i18n.Localizer)
		}
	}

	// 回退到默认语言
	loc, _ := m.localizes.Load(m.config.DefaultLang)
	return loc.(*i18n.Localizer)
}

// Localize 本地化消息（必须存在）
func (m *Manager) Localize(lang, msgID string, data map[string]interface{}) string {
	return m.getLocalizer(lang).MustLocalize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: data,
	})
}

// LocalizeTry 尝试本地化消息（可能不存在）
func (m *Manager) LocalizeTry(lang, msgID string, data map[string]interface{}) string {
	msg, err := m.getLocalizer(lang).Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: data,
	})
	if err != nil && m.config.Debug {
		m.logger.Error("localize failed", "error", err, "msgID", msgID)
	}
	return msg
}

func Localize(lang, msgID string, data map[string]interface{}) string {
	return defaultManager.Localize(lang, msgID, data)
}

func LocalizeTry(lang, msgID string, data map[string]interface{}) string {
	return defaultManager.LocalizeTry(lang, msgID, data)
}

func LocalizeDef(msgID string, data map[string]interface{}) string {
	return defaultManager.Localize(defaultManager.config.DefaultLang, msgID, data)
}

func LocalizeDefTry(msgID string, data map[string]interface{}) string {
	return defaultManager.LocalizeTry(defaultManager.config.DefaultLang, msgID, data)
}