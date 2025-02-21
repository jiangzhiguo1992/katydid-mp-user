package configs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager 配置管理器
type Manager struct {
	v      *viper.Viper
	config *Config
	mu     sync.RWMutex
}

var (
	defaultManager *Manager
	managerOnce    sync.Once
)

// GetManager 获取配置管理器单例
func GetManager() *Manager {
	managerOnce.Do(func() {
		defaultManager = &Manager{
			v:      viper.New(),
			config: new(Config),
		}
	})
	return defaultManager
}

func Get() *Config {
	m := GetManager()
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Init 初始化配置
func Init(confDir string, reload func() bool) (*Config, error) {
	m := GetManager()
	if err := m.Load(confDir, reload); err != nil {
		return nil, err
	}
	return m.config, nil
}

// Load 加载配置
func (m *Manager) Load(confDir string, reload func() bool) error {
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
	m.watchConfig(reload)

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
func (m *Manager) watchConfig(reload func() bool) {
	m.v.OnConfigChange(func(e fsnotify.Event) {
		if reload == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), reloadInterval*reloadMaxRetries)
		defer cancel()

		for i := 0; i < reloadMaxRetries; i++ {
			if reload() {
				log.Printf("■ ■ Conf ■ ■ config reloaded successfully")
				return
			}

			select {
			case <-time.After(reloadInterval):
				log.Printf("■ ■ Conf ■ ■ reload config failed, retry: %d", i+1)
			case <-ctx.Done():
				log.Printf("■ ■ Conf ■ ■ reload config timeout after %d retries", i+1)
				return
			}
		}
		log.Printf("■ ■ Conf ■ ■ reload config failed after %d retries", reloadMaxRetries)
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
	//			log.Errorf("unable to read remote config: %v", err)
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
	//	log.Fatalf("unable to read remote config: %v", err)
	//}
	//// 监听Consul配置变化
	//go func() {
	//	for {
	//		select {
	//		case <-viper.RemoteConfig.WatchChannel():
	//			log.Println("remote config changed")
	//
	//		}
	//	}
	//}()
}
