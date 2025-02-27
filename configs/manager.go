package configs

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	Key      string
	Callback func(interface{})
}

func GetManager() *Manager {
	//manager.mu.RLock()
	//defer manager.mu.RUnlock()
	return manager
}

// Get 获取配置管理器单例
func Get() *Config {
	//manager.mu.RLock()
	//defer manager.mu.RUnlock()
	return manager.config
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
func (m *Manager) Subscribe(key string, callback func(interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers = append(m.subscribers, ChangeSubscriber{
		Key:      key,
		Callback: callback,
	})
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

	// 在Load方法中添加 TODO:GG 根据config里的配置，来启动远程配置监听
	//if m.remoteEnabled {
	//	if err := m.loadRemoteConfig(); err != nil {
	//		return fmt.Errorf("■ ■ Conf ■ ■ load remote config failed: %w", err)
	//	}
	//}
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

//// loadRemoteConfig 加载远程配置
//func (m *Manager) loadRemoteConfig(kind, addr, key string) error {
//	err := m.v.AddRemoteProvider(kind, addr, key)
//	if err != nil {
//		return err
//	}
//
//	err = m.v.ReadRemoteConfig()
//	if err != nil {
//		return err
//	}
//
//	// 启动远程配置监听
//	go func() {
//		for {
//			time.Sleep(30 * time.Second)
//			err := m.v.WatchRemoteConfig()
//			if err != nil {
//				slog.Error("■ ■ Conf ■ ■ watch remote config failed", slog.Any("err", err))
//				continue
//			}
//			if err := m.parseConfig(); err != nil {
//				slog.Error("■ ■ Conf ■ ■ parse remote config failed", slog.Any("err", err))
//			}
//		}
//	}()
//}
