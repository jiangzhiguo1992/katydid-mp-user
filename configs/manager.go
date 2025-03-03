package configs

import (
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
	appDir = "app" // 初始加载目录
	envKey = "env" // 环境key
)

var (
	manager *Manager
	once    sync.Once
)

// Manager 配置管理器
type Manager struct {
	v           *viper.Viper
	subs        []*viper.Viper
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
			v:           viper.GetViper(),
			subs:        make([]*viper.Viper, 0),
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

	// 加载初始配置 (先)
	if err := m.loadConfigs(confDir, appDir); err != nil {
		return err
	}
	if err := m.parseConfigs(); err != nil {
		return err
	}

	// 加载环境配置 (后)
	envDir := m.v.GetString(envKey)
	if err := m.loadConfigs(confDir, envDir); err != nil {
		return err
	}
	if err := m.parseConfigs(); err != nil {
		return err
	}

	// 设置配置监听
	m.watchConfigs()

	// 远程配置加载
	if m.config.RemoteConf.Enabled {
		if err := m.loadRemoteConfig(); err != nil {
			return err
		}
	}

	// debug模式下打印配置
	if m.config.IsDebug() {
		m.v.Debug()
	}
	return nil
}

// loadConfigs 加载环境配置文件
func (m *Manager) loadConfigs(confDir string, subs ...string) error {
	dir := filepath.Join(append([]string{confDir}, subs...)...)

	// 读取目录下的所有file
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ read configs %s failed: %w", dir, err)
	}

	// 加载每个文件
	for _, file := range files {
		if file.IsDir() {
			return m.loadConfigs(filepath.Join(dir, file.Name()))
		} else {
			err := m.loadConfig(dir, file.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// loadConfigs 加载配置文件
func (m *Manager) loadConfig(confDir string, fileName string) error {
	// 为每个目录创建新的Viper实例
	subViper := viper.New()

	// 支持多种配置格式
	subViper.SetConfigType("toml")
	if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
		subViper.SetConfigType("yaml")
	} else if strings.HasSuffix(fileName, ".json") {
		subViper.SetConfigType("json")
	}

	// 设置配置文件路径
	subViper.AddConfigPath(confDir)

	// 提取不带扩展名的文件名
	realFileName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	subViper.SetConfigName(realFileName)

	// 读取配置文件
	if err := subViper.ReadInConfig(); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ read config %s/%s failed: %w", confDir, fileName, err)
	}
	slog.Info("■ ■ Conf ■ ■ read config success", slog.String("path", filepath.Join(confDir, fileName)))

	// 添加到列表
	m.subs = append(m.subs, subViper)
	return nil
}

// parseConfig 解析配置到结构体
func (m *Manager) parseConfigs() error {
	// 将子配置合并到主配置
	for _, sub := range m.subs {
		if sub == nil {
			continue
		}
		for _, key := range sub.AllKeys() {
			m.v.Set(key, sub.Get(key))
		}
	}

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

// watchConfigs 监听配置变化
func (m *Manager) watchConfigs() {
	for _, sub := range m.subs {
		if sub == nil {
			continue
		}
		sub.OnConfigChange(func(e fsnotify.Event) {
			// 重新加载配置
			err := m.parseConfigs()
			if err != nil {
				slog.Error("■ ■ Conf ■ ■ config change failed", slog.Any("err", err))
				return
			}
			// 通知订阅者
			for _, subscriber := range m.subscribers {
				if sub.IsSet(subscriber.Key) {
					value := sub.Get(subscriber.Key)
					slog.Info(
						"■ ■ Conf ■ ■ config change success",
						slog.String("key", subscriber.Key),
						slog.Any("Value", value),
					)
					go subscriber.Callback(value)
				}
			}
		})
		sub.WatchConfig()
	}
}

// loadRemoteConfig 加载远程配置 TODO:GG test
func (m *Manager) loadRemoteConfig() error {
	if !m.config.RemoteConf.Enabled {
		return nil
	}

	// TODO:GG viper.SupportedRemoteProviders 支持的类型

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
			if err := m.parseConfigs(); err != nil {
				slog.Error("■ ■ Conf ■ ■ parse remote config failed", slog.Any("err", err))
			}
			m.mu.Unlock()
		}
	}()
	return nil
}
