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
	appDir   = "app"  // 初始加载目录
	envKey   = "env"  // 环境目录Key
	initFile = "init" // 初始化文件名
)

var (
	manager *Manager
	once    sync.Once
)

// Manager 配置管理器
type Manager struct {
	main        *viper.Viper
	subs        map[int]map[string]*viper.Viper
	config      *Config
	subscribers []ChangeSubscriber
	mu          sync.RWMutex
}

// ChangeSubscriber 配置变更订阅
type ChangeSubscriber struct {
	ID       string
	Key      string
	Callback func(value any)
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
			main:        viper.GetViper(),
			subs:        make(map[int]map[string]*viper.Viper),
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
func Subscribe(key string, callback func(value any)) func() {
	if manager == nil {
		return func() {}
	}
	id := uuid.New().String()
	slog.Info("■ ■ Conf ■ ■ subscribe", slog.String("key", key), slog.String("id", id))

	manager.mu.Lock()
	manager.subscribers = append(manager.subscribers, ChangeSubscriber{
		ID:       id,
		Key:      key,
		Callback: callback,
	})
	manager.mu.Unlock()

	// 返回取消订阅函数
	return func() {
		manager.mu.Lock()
		defer manager.mu.Unlock()

		for i, sub := range manager.subscribers {
			if sub.ID == id {
				// 删除订阅
				manager.subscribers[i] = manager.subscribers[len(manager.subscribers)-1]
				manager.subscribers = manager.subscribers[:len(manager.subscribers)-1]
				slog.Info("■ ■ Conf ■ ■ subscribe cancel", slog.String("key", sub.Key), slog.String("id", sub.ID))
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
	m.main.SetEnvPrefix("APP")
	// 自动查找环境变量
	m.main.AutomaticEnv()
	// 使用 . 替换环境变量中的 _
	m.main.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 加载初始配置 (先)
	if err := m.loadConfigs(0, confDir, appDir); err != nil {
		return err
	}

	// 解析app配置(主要是为了拿envKey)
	if err := m.parseConfigs(); err != nil {
		return err
	}

	// 加载环境配置 (后)
	if envDir := m.main.GetString(envKey); envDir != "" {
		if err := m.loadConfigs(1, confDir, envDir); err != nil {
			return err
		}
		// 再次解析配置(全部)
		if err := m.parseConfigs(); err != nil {
			return err
		}
	}

	// 设置配置监听
	m.watchConfigs()

	// 远程配置加载
	if m.config.RemoteConf.Enabled {
		if err := m.loadRemoteConfig(); err != nil {
			return err
		}
	}

	// 打印配置
	m.logSettings("", m.main.AllSettings())

	// debug模式下打印配置
	if !m.config.IsProd() {
		m.main.Debug()
	}

	return nil
}

// loadConfigs 加载环境配置文件
func (m *Manager) loadConfigs(priority int, confDir string, subs ...string) error {
	dir := filepath.Join(append([]string{confDir}, subs...)...)

	// 读取目录下的所有file
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ read dir %s failed: %w", dir, err)
	}

	// files排序，init第一个
	if len(files) > 1 {
		sortedFiles := make([]os.DirEntry, 0, len(files))
		initFiles := make([]os.DirEntry, 0)
		otherFiles := make([]os.DirEntry, 0)

		for _, f := range files {
			name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			if name == initFile {
				initFiles = append(initFiles, f)
			} else {
				otherFiles = append(otherFiles, f)
			}
		}

		sortedFiles = append(sortedFiles, initFiles...)
		sortedFiles = append(sortedFiles, otherFiles...)
		files = sortedFiles
	}

	// 加载每个文件
	for _, file := range files {
		if file.IsDir() {
			return m.loadConfigs(priority, filepath.Join(dir, file.Name()))
		} else {
			err = m.loadConfig(priority, dir, file.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// loadConfigs 加载配置文件
func (m *Manager) loadConfig(priority int, confDir string, fileName string) error {
	// 为每个目录创建新的Viper实例
	if _, ok := m.subs[priority]; !ok {
		m.subs[priority] = make(map[string]*viper.Viper)
	}
	subsKey := filepath.Join(confDir, fileName)
	var subViper *viper.Viper
	if v, ok := m.subs[priority][subsKey]; ok {
		subViper = v
	} else {
		subViper = viper.New()
		m.subs[priority][subsKey] = subViper
	}

	// 根据文件扩展名设置配置类型
	ext := filepath.Ext(fileName)
	configType := "toml" // 默认类型
	switch ext {
	case ".yaml", ".yml":
		configType = "yaml"
	case ".json":
		configType = "json"
	case ".toml":
		configType = "toml"
	default:
		return nil // 跳过不支持的文件类型
	}
	subViper.SetConfigType(configType)

	// 设置配置文件路径
	subViper.AddConfigPath(confDir)

	// 提取不带扩展名的文件名
	realFileName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	subViper.SetConfigName(realFileName)

	// 读取配置文件
	if err := subViper.ReadInConfig(); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ read file %s/%s failed: %w", confDir, fileName, err)
	}
	slog.Info("■ ■ Conf ■ ■ read file success", slog.String("path", subsKey))

	return nil
}

// parseConfig 解析配置到结构体
func (m *Manager) parseConfigs() error {
	// 将子配置合并到主配置，priority从0开始覆盖
	for priority := 0; priority < len(m.subs); priority++ {
		if subs, ok := m.subs[priority]; ok {
			for _, sub := range subs {
				if sub == nil {
					continue
				}
				for _, key := range sub.AllKeys() {
					m.main.Set(key, sub.Get(key))
				}
			}
			continue
		}
		break
	}

	// 首次解析
	if err := m.main.Unmarshal(m.config); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ unmarshal failed: %w", err)
	}

	// 设置默认值
	m.config.merge()

	// 再次解析覆盖默认值
	if err := m.main.Unmarshal(m.config); err != nil {
		return fmt.Errorf("■ ■ Conf ■ ■ unmarshal again failed: %w", err)
	}

	return nil
}

// watchConfigs 监听配置变化
func (m *Manager) watchConfigs() {
	for _, subs := range m.subs {
		for _, sub := range subs {
			if sub == nil {
				continue
			}
			sub.OnConfigChange(func(e fsnotify.Event) {
				slog.Info(
					"■ ■ Conf ■ ■ local on_change success",
					slog.String("op", e.Op.String()),
					slog.String("name", e.Name),
				)

				// 保存变更前的配置快照
				prevConfig := make(map[string]interface{})
				for _, key := range m.main.AllKeys() {
					prevConfig[key] = m.main.Get(key)
				}

				// 重新加载配置
				err := m.parseConfigs()
				if err != nil {
					slog.Error("■ ■ Conf ■ ■ local on_change failed", slog.Any("err", err))
					return
				}

				// 通知订阅者
				for _, subscriber := range m.subscribers {
					if !m.main.IsSet(subscriber.Key) {
						continue
					}
					if prevConfig[subscriber.Key] == m.main.Get(subscriber.Key) {
						continue
					}

					value := m.main.Get(subscriber.Key)
					slog.Info(
						"■ ■ Conf ■ ■ local subscribe callback",
						slog.String("key", subscriber.Key),
						slog.Any("Value", value),
					)

					go subscriber.Callback(value)
				}
			})
			sub.WatchConfig()
		}
	}
}

// loadRemoteConfig 加载远程配置 TODO:GG test
func (m *Manager) loadRemoteConfig() error {
	if !m.config.RemoteConf.Enabled {
		return nil
	}

	// TODO:GG viper.SupportedRemoteProviders 支持的类型

	err := m.main.AddRemoteProvider(
		m.config.RemoteConf.Provider,
		m.config.RemoteConf.Endpoint,
		m.config.RemoteConf.Path,
	)
	if err != nil {
		return err
	}

	// 设置远程配置类型
	m.main.SetConfigType("toml")

	if err := m.main.ReadRemoteConfig(); err != nil {
		return err
	}

	// 启动远程配置监听 TODO:GG 需要改，没有结合subscribers
	refreshInterval := time.Duration(m.config.RemoteConf.RefreshInterval) * time.Second
	if refreshInterval < 10*time.Second {
		refreshInterval = 30 * time.Second // 默认最小刷新间隔
	}

	go func() {
		for {
			time.Sleep(refreshInterval)
			err := m.main.WatchRemoteConfig()
			if err != nil {
				slog.Error("■ ■ Conf ■ ■ watch remote failed", slog.Any("err", err))
				continue
			}

			m.mu.Lock()
			if err := m.parseConfigs(); err != nil {
				slog.Error("■ ■ Conf ■ ■ parse remote failed", slog.Any("err", err))
			}
			m.mu.Unlock()
		}
	}()
	return nil
}

// logSettings 打印配置
func (m *Manager) logSettings(group string, settings map[string]any) {
	for k, v := range settings {
		if vs, ok := v.(map[string]any); ok {
			nextGroup := ""
			if len(group) <= 0 {
				nextGroup = fmt.Sprintf("%s.", k)
			} else {
				nextGroup = fmt.Sprintf("%s%s.", group, k)
			}
			m.logSettings(nextGroup, vs)
			continue
		}
		key := fmt.Sprintf("%s%s ", group, k)
		val := slog.Any("", v).Value.String()
		sprintf := fmt.Sprintf("%s = %s", key, val)
		slog.Info(fmt.Sprintf("■ ■ Conf ■ ■ ---> %s", sprintf))
	}
}
