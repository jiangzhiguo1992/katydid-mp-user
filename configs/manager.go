package configs

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	envKey = "env" // 环境key
)

var (
	manager *Manager
	once    sync.Once
)

// Manager 配置管理器
type Manager struct {
	v           *viper.Viper
	config      *Config
	subscribers []ChangeSubscriber
	mu          sync.RWMutex
}

// ChangeSubscriber 配置变更订阅
type ChangeSubscriber struct {
	ID       string
	Key      string
	Callback func(interface{})
}

// Get 获取配置管理器单例
func Get() *Config {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	// 返回只读副本而非直接引用
	config := *manager.config
	return &config
}

// Init 初始化配置
func Init(confDir string) (*Config, error) {
	once.Do(func() {
		manager = &Manager{
			v:           viper.New(),
			config:      new(Config),
			subscribers: make([]ChangeSubscriber, 0),
		}
	})
	if err := manager.load(confDir); err != nil {
		return nil, err
	}
	return manager.config, nil
}

// Subscribe 注册配置监听
func Subscribe(key string, callback func(interface{})) func() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	id := uuid.New().String()
	manager.subscribers = append(manager.subscribers, ChangeSubscriber{
		ID:       id,
		Key:      key,
		Callback: callback,
	})

	// 返回取消订阅函数
	return func() {
		manager.mu.Lock()
		defer manager.mu.Unlock()

		for i, sub := range manager.subscribers {
			if sub.ID == id {
				// 删除订阅
				manager.subscribers = append(manager.subscribers[:i], manager.subscribers[i+1:]...)
				break
			}
		}
	}
}

// Load 加载配置
func (m *Manager) load(confDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 设置环境变量前缀
	m.v.SetEnvPrefix("APP")
	// 自动查找环境变量
	m.v.AutomaticEnv()
	// 使用 . 替换环境变量中的 _
	m.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 支持多种配置格式
	m.v.SetConfigType("toml") // 默认格式
	// 如果需要自动识别
	if strings.HasSuffix(confDir, ".yaml") || strings.HasSuffix(confDir, ".yml") {
		m.v.SetConfigType("yaml")
	} else if strings.HasSuffix(confDir, ".json") {
		m.v.SetConfigType("json")
	}

	// 加载内置配置
	if err := m.loadEmbedConfigs(); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ load embed configs failed: %w", err)
	}

	// 加载环境配置
	if err := m.loadEnvConfigs(confDir); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ load env configs failed: %w", err)
	}

	// 解析配置
	if err := m.parseConfig(); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ parse config failed: %w", err)
	}

	// 设置配置监听
	m.watchConfig()

	// 远程配置加载
	if m.config.RemoteConf.Enabled {
		if err := m.loadRemoteConfig(); err != nil {
			return fmt.Errorf("■ ■ Conf ■ ■ load remote config failed: %w", err)
		}
	}
	return nil
}

// loadEmbedConfigs 加载内置配置文件
func (m *Manager) loadEmbedConfigs() error {
	configs := [][]byte{fileAppInit, fileAppPub, fileAppPri}
	for _, cfg := range configs {
		if err := m.v.MergeConfig(bytes.NewReader(cfg)); err != nil {
			return fmt.Errorf("■ ■ Conf ■ ■ merge embed config failed: %w", err)
		}
	}
	return nil
}

// loadEnvConfigs 加载环境配置文件
func (m *Manager) loadEnvConfigs(confDir string) error {
	envDir := filepath.Join(confDir, m.v.GetString(envKey))
	files, err := os.ReadDir(envDir)
	if err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ read env dir failed: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(envDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("■ ■ Conf ■ ■ read file %s failed: %w", filePath, err)
			}
			if err := m.v.MergeConfig(bytes.NewReader(content)); err != nil {
				return fmt.Errorf("■ ■ Conf ■ ■ merge env config %s failed: %w", filePath, err)
			}
		}
	}
	return nil
}

// parseConfig 解析配置到结构体
func (m *Manager) parseConfig() error {
	// 首次解析
	if err := m.v.Unmarshal(m.config); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ unmarshal config failed: %w", err)
	}

	// 设置默认值
	m.config.merge()

	// 再次解析覆盖默认值
	if err := m.v.Unmarshal(m.config); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ unmarshal config with defaults failed: %w", err)
	}

	return nil
}

// watchConfig 监听配置变化
func (m *Manager) watchConfig() {
	m.v.OnConfigChange(func(e fsnotify.Event) {
		// 通知订阅者
		for _, subscriber := range m.subscribers {
			if m.v.IsSet(subscriber.Key) {
				value := m.v.Get(subscriber.Key)
				slog.Info(
					"■ ■ Conf ■ ■ reload config",
					slog.String("key", subscriber.Key),
					slog.Any("Value", value),
				)
				go subscriber.Callback(value)
			}
		}
	})

	m.v.WatchConfig()
}

// loadRemoteConfig 加载远程配置
func (m *Manager) loadRemoteConfig() error {
	if !m.config.RemoteConf.Enabled {
		return nil
	}

	err := m.v.AddRemoteProvider(
		m.config.RemoteConf.Provider,
		m.config.RemoteConf.Endpoint,
		m.config.RemoteConf.Path,
	)
	if err != nil {
		return err
	}

	// 设置远程配置类型
	m.v.SetConfigType("toml")

	if err := m.v.ReadRemoteConfig(); err != nil {
		return err
	}

	// 启动远程配置监听
	refreshInterval := time.Duration(m.config.RemoteConf.RefreshInterval) * time.Second
	if refreshInterval < 10*time.Second {
		refreshInterval = 30 * time.Second // 默认最小刷新间隔
	}

	go func() {
		for {
			time.Sleep(refreshInterval)
			err := m.v.WatchRemoteConfig()
			if err != nil {
				slog.Error("■ ■ Conf ■ ■ watch remote config failed", slog.Any("err", err))
				continue
			}

			m.mu.Lock()
			if err := m.parseConfig(); err != nil {
				slog.Error("■ ■ Conf ■ ■ parse remote config failed", slog.Any("err", err))
			}
			m.mu.Unlock()
		}
	}()
	return nil
}
