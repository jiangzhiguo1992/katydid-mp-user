package i18n

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultLang  = "en"
	unknownError = "localize unknown error"
)

var defaultManager *Manager

type Config struct {
	DefaultLang  string                       // 默认语言
	CacheMaxSize int                          // 默认1000
	DocDirs      []string                     // 解析文件夹
	OnInfo       func(string, map[string]any) `json:"-"`
	OnErr        func(string, map[string]any) `json:"-"`
}

type Manager struct {
	config     Config
	bundle     *i18n.Bundle
	langs      []string
	langsMap   map[string]bool                       // 用于快速查找语言是否支持
	localizer  *lru.Cache[string, *i18n.Localizer]   // 使用LRU缓存代替sync.Map
	localizers *lru.Cache[string, []*i18n.Localizer] // 使用LRU缓存代替sync.Map
}

func Init(cfg Config) error {
	//marshal, _ := json.MarshalIndent(cfg, "", "\t")
	//cfg.OnInfo(fmt.Sprintf("■ ■ i18n ■ ■ 配置 ---> %s", marshal), nil)

	// 确保回调函数不为nil，避免空指针异常
	if cfg.OnInfo == nil {
		cfg.OnInfo = func(string, map[string]any) {}
	}

	if cfg.OnErr == nil {
		cfg.OnErr = func(string, map[string]any) {}
	}

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
		return defaultLang
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

	// 创建LRU缓存，设置合理的大小限制
	if cfg.CacheMaxSize <= 0 {
		cfg.CacheMaxSize = 1000 // 使用合理的默认值
	}
	localizerCache, err := lru.New[string, *i18n.Localizer](cfg.CacheMaxSize)
	if err != nil {
		return nil, fmt.Errorf("■ ■ i18n ■ ■ 创建localizer缓存失败: %w", err)
	}

	localizersCache, err := lru.New[string, []*i18n.Localizer](cfg.CacheMaxSize)
	if err != nil {
		return nil, fmt.Errorf("■ ■ i18n ■ ■ 创建localizers缓存失败: %w", err)
	}

	m := &Manager{
		config:     cfg,
		bundle:     i18n.NewBundle(defaultTag),
		langs:      nil,
		langsMap:   make(map[string]bool),
		localizer:  localizerCache,
		localizers: localizersCache,
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
	if len(m.config.DocDirs) == 0 {
		return fmt.Errorf("■ ■ i18n ■ ■ 未配置文档目录")
	}

	var files []string
	for _, dir := range m.config.DocDirs {
		if dir == "" {
			continue // 跳过空目录配置
		}

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
	m.langsMap = make(map[string]bool, len(files))

	for _, file := range files {
		if _, err := m.bundle.LoadMessageFile(file); err != nil {
			return fmt.Errorf("■ ■ i18n ■ ■ 加载文件 %s 失败: %w", file, err)
		}

		lang := extractLangFromFilename(file)
		if lang == "" {
			continue // 跳过无法提取语言的文件
		}

		m.langs = append(m.langs, lang)
		m.langsMap[lang] = true
	}

	marshal2, _ := json.MarshalIndent(map[string]any{"languages": m.langs}, "", "\t")
	m.config.OnInfo(fmt.Sprintf("■ ■ i18n ■ ■ 本地化加载语言: %s", marshal2), nil)

	// 确保默认语言存在
	if !m.langsMap[m.config.DefaultLang] {
		m.config.OnErr("■ ■ i18n ■ ■ 本地化默认文件不存在 ", map[string]any{"lang": m.config.DefaultLang})
	}

	// 预缓存默认Localizer
	m.getLocalizers(m.config.DefaultLang)

	return nil
}

func filterMessageFiles(files []string) []string {
	if len(files) == 0 {
		return nil
	}

	result := make([]string, 0, len(files))

	for _, f := range files {
		if f == "" {
			continue
		}

		if fi, err := os.Stat(f); err == nil && !fi.IsDir() {
			result = append(result, f)
		}
	}
	return result
}

func extractLangFromFilename(file string) string {
	base := filepath.Base(file)
	ext := filepath.Ext(base)

	// 避免无扩展名的情况下可能的索引越界
	if len(ext) == 0 {
		return base
	}

	name := base[:len(base)-len(ext)]

	// 查找最后一个点的位置
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

// getLocalizers 获取缓存的localizer，如果不存在则创建
func (m *Manager) getLocalizers(lang string) []*i18n.Localizer {
	// 从LRU缓存获取
	if val, ok := m.localizers.Get(lang); ok {
		return val
	}

	// 预分配适当大小的切片，避免动态扩容
	tags := make([]string, 0, 3) // 大多数情况下不超过3个标签

	// 添加精确匹配的语言标签
	if m.langsMap[lang] {
		tags = append(tags, lang)
	}

	// 添加基本语言标签（如 "en" 从 "en-US"）
	if dashIndex := strings.IndexByte(lang, '-'); dashIndex > 0 {
		base := lang[:dashIndex]
		if base != lang && m.langsMap[base] {
			tags = append(tags, base)
		}
	}

	// 确保添加默认语言作为最后的回退选项
	if len(tags) == 0 || tags[len(tags)-1] != m.config.DefaultLang {
		tags = append(tags, m.config.DefaultLang)
	}

	// 预分配localizers切片
	localizers := make([]*i18n.Localizer, len(tags))
	for i := 0; i < len(tags); i++ {
		tt := strings.Join(tags[i:], "_")
		if val, ok := m.localizer.Get(tt); ok {
			localizers[i] = val
			continue
		}
		localizer := i18n.NewLocalizer(m.bundle, tags[i:]...)
		m.localizer.Add(tt, localizer)
		localizers[i] = localizer
	}

	// 缓存结果
	m.localizers.Add(lang, localizers)
	return localizers
}

func (m *Manager) localize(lang, msgID string, data map[string]any, fallbackToID bool) string {
	// 构建语言标签列表，实现回退链
	localizers := m.getLocalizers(lang)

	// 使用共享的LocalizeConfig对象减少内存分配
	config := &i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: data,
		DefaultMessage: &i18n.Message{
			ID: msgID,
			//Other: unknownError,
		},
	}

	// 循环查找
	for _, localizer := range localizers {
		if localizer == nil {
			continue // 避免空指针引用
		}

		msg, err := localizer.Localize(config)
		if err == nil && len(msg) > 0 {
			return msg // 找到有效翻译直接返回
		}
	}

	// 只在最后才处理错误情况
	if fallbackToID {
		return msgID
	}

	// 记录错误
	if m.config.OnErr != nil {
		m.config.OnErr(
			"■ ■ i18n ■ ■ 本地化未找到: ",
			map[string]any{"msgID": msgID, "lang": lang},
		)
	}

	return unknownError
}

func LocalizeMust(lang, msgID string, data map[string]any) string {
	if defaultManager == nil {
		return unknownError
	}

	if lang == "" {
		lang = defaultManager.config.DefaultLang
	}

	return defaultManager.localize(lang, msgID, data, false)
}

func LocalizeTry(lang, msgID string, data map[string]any) string {
	if defaultManager == nil {
		return unknownError
	}

	if lang == "" {
		lang = defaultManager.config.DefaultLang
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
