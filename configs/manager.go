package configs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	envKey = "env" // 环境key

	reConfigMaxRetries = 3               // 重试次数
	reConfigInterval   = 2 * time.Second // 重试间隔
)

var (
	manager *Manager
	once    sync.Once
)

// Manager 配置管理器
type Manager struct {
	v        *viper.Viper
	config   *Config
	reConfig func() bool
	mu       sync.RWMutex
}

// Get 获取配置管理器单例
func Get() *Config {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.config
}

// Init 初始化配置
func Init(confDir string, reConfig func() bool) (*Config, error) {
	once.Do(func() {
		manager = &Manager{
			v:        viper.New(),
			config:   new(Config),
			reConfig: reConfig,
		}
	})
	if err := manager.Load(confDir); err != nil {
		return nil, err
	}
	return manager.config, nil
}

// Load 加载配置
func (m *Manager) Load(confDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 设置基本配置
	m.v.SetConfigType("toml")

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
		if m.reConfig == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), reConfigInterval*reConfigMaxRetries)
		defer cancel()

		for i := 0; i < reConfigMaxRetries; i++ {
			if m.reConfig() {
				slog.Info("■ ■ Conf ■ ■ config reloaded successfully")
				return
			}

			select {
			case <-time.After(reConfigInterval):
				slog.Warn("■ ■ Conf ■ ■ reload config failed", slog.Int("retry", i+1))
			case <-ctx.Done():
				slog.Warn("■ ■ Conf ■ ■ reload config timeout", slog.Int("retries", i+1))
				return
			}
		}
		slog.Error("■ ■ Conf ■ ■ reload config failed", slog.Int("retries", reConfigMaxRetries))
	})

	m.v.WatchConfig()
}

// loadRemoteConfig 加载远程配置
// TODO:GG 加载远程配置
func loadRemoteConfig() {
	//var runtime_viper = viper.New()
	//runtime_viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/hugo.yml")
	//runtime_viper.SetConfigType("yaml") // because there is no file extension in a stream of bytes, supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop", "env", "dotenv"
	//
	//// read from remote config the first time.
	//err := runtime_viper.ReadRemoteConfig()
	//
	//// unmarshal config
	//runtime_viper.Unmarshal(&runtime_conf)
	//
	//// open a goroutine to watch remote changes forever
	//go func() {
	//	for {
	//		time.Sleep(time.Second * 5) // delay after each request
	//
	//		// currently, only tested with etcd support
	//		err := runtime_viper.WatchRemoteConfig()
	//		if err != nil {
	//			slog.Errorf("unable to read remote config: %v", err)
	//			continue
	//		}
	//
	//		// unmarshal new config into our runtime config struct. you can also use channel
	//		// to implement a signal to notify the system of the changes
	//		runtime_viper.Unmarshal(&runtime_conf)
	//	}
	//}()
	//err := viper.WatchRemoteConfig()
	//if err != nil {
	//	slog.Fatalf("unable to read remote config: %v", err)
	//}
	//// 监听Consul配置变化
	//go func() {
	//	for {
	//		select {
	//		case <-viper.RemoteConfig.WatchChannel():
	//			slog.Println("remote config changed")
	//
	//		}
	//	}
	//}()
}
