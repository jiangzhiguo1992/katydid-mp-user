package configs

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"time"
)

var config = new(Config)

var (
	//go:embed app/init.toml
	fileAppInit []byte
	//go:embed app/private.toml
	fileAppPri []byte
	//go:embed app/public.toml
	fileAppPub []byte
)

const (
	envKey = "env" // 环境key
)

type (
	Config struct {
		Env     string `toml:"env" mapstructure:"env"`
		EnvName string `toml:"env_name" mapstructure:"env_name"`

		AppConf `mapstructure:",squash"`

		//ServerConf `toml:"server" mapstructure:"server"`

		Account AccountConf `toml:"account" mapstructure:"account"`
		Client  ClientConf  `toml:"client" mapstructure:"client"`
		User    UserConf    `toml:"user" mapstructure:"user"`
	}

	AppConf struct {
		ModuleConf `mapstructure:",squash"`
	}

	ServerConf struct {
	}

	AccountConf struct {
		ModuleConf `mapstructure:",squash"`
	}

	ClientConf struct {
		ModuleConf `mapstructure:",squash"`
	}

	UserConf struct {
		ModuleConf `mapstructure:",squash"`
	}

	ModuleConf struct {
		Enable  bool        `toml:"enable" mapstructure:"enable"`
		PgSql   PgSqlConf   `toml:"pgsql" mapstructure:"pgsql"`
		Redis   RedisConf   `toml:"redis"  mapstructure:"redis"`
		MongoDB MongoDBConf `toml:"mongo_db"  mapstructure:"mongo_db"`
	}

	PgSqlConf struct {
		Write struct {
			Host string `toml:"host" mapstructure:"host"`
			Port string `toml:"port" mapstructure:"port"`
			DB   string `toml:"db" mapstructure:"db"`
			User string `toml:"user" mapstructure:"user"`
			Pwd  string `toml:"pwd" mapstructure:"pwd"`
		} `toml:"write"`
		Read struct {
			Host []string `toml:"host" mapstructure:"host"`
			Port []string `toml:"port" mapstructure:"port"`
			DB   []string `toml:"db" mapstructure:"db"`
			User []string `toml:"user" mapstructure:"user"`
			Pwd  []string `toml:"pwd" mapstructure:"pwd"`
		} `toml:"read"`
		// TODO:GG mysql?
		MaxOpenConn     int           `toml:"max_open_conn" mapstructure:"max_open_conn"`
		MaxIdleConn     int           `toml:"max_idle_conn" mapstructure:"max_idle_conn"`
		ConnMaxLifeTime time.Duration `toml:"conn_max_life_time" mapstructure:"conn_max_life_time"`
		// TODO:GG other
		Timeout    int    `toml:"timeout" mapstructure:"timeout"`
		TimeZone   string `toml:"timezone" mapstructure:"timezone"`
		SSLMode    string `toml:"ssl_mode" mapstructure:"ssl_mode"`
		MaxRetries int    `toml:"max_retries" mapstructure:"max_retries"`
		RetryDelay int    `toml:"retry_delay" mapstructure:"retry_delay"`
	}

	RedisConf struct {
		// TODO:GG 也需要支持集群模式？
		Host        string `toml:"host" mapstructure:"host"`
		Port        string `toml:"port" mapstructure:"port"`
		DB          string `toml:"db" mapstructure:"db"`
		Pwd         string `toml:"pwd" mapstructure:"pwd"`
		MaxRetries  int    `toml:"max_retries" mapstructure:"max_retries"`
		PoolSize    int    `toml:"pool_size" mapstructure:"pool_size"`
		MinIdleConn int    `toml:"min_idle_conn" mapstructure:"min_idle_conn"`
	}

	MongoDBConf struct {
		// TODO:GG
	}
)

func (m *Config) IsProd() bool {
	return m.Env == "pro"
}

func (m *Config) merge() {
	config.Account.ModuleConf = config.ModuleConf
	config.Client.ModuleConf = config.ModuleConf
	config.User.ModuleConf = config.ModuleConf
}

func InitConfig(confDir string) *Config {
	loadLocalConfig(confDir)
	//loadRemoteConfig()
	return config
}

func loadLocalConfig(confDir string) {
	viper.SetConfigType("toml")

	// 初始app目录的conf加载
	bs := [][]byte{fileAppInit, fileAppPri, fileAppPub}
	for _, b := range bs {
		//if err := viper.ReadConfig(bytes.NewReader(b)); err != nil {
		if err := viper.MergeConfig(bytes.NewReader(b)); err != nil {
			log.Fatalf("merge config failed: %v", err)
		}
	}

	// 找到指定env的dir下的所有文件，并集合到io.reader中
	envDir := filepath.Join(confDir, viper.GetString(envKey))
	files, err := os.ReadDir(envDir)
	if err != nil {
		log.Fatalf("failed to read directory: %v", err)
	}

	// 读取并加载，重复会覆盖app.conf
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(envDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatalf("failed to read file: %v", err)
			}
			if err := viper.MergeConfig(bytes.NewReader(content)); err != nil {
				log.Fatalf("merge env.config failed: %v", err)
			}
		}
	}

	// 打印所有conf的kv
	settings := viper.AllSettings()
	log.Printf("config.settings init suceess: %v", settings)

	// 解析到config结构体
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into config_1, %v", err)
	}
	// 设置默认值
	config.merge()
	// 再次解析,覆盖默认值
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into config_2, %v", err)
	}

	log.Printf("config: %v", config)

	viper.OnConfigChange(func(e fsnotify.Event) {
		// TODO:GG 配置会被更新，这里要做一些相关的re_init操作
		fmt.Printf("config file changed name:%s\n", e.Name)
		//reloadConfig()
	})
	viper.WatchConfig()
}

// loadRemoteConfig 加载远程配置
// TODO:GG 加载远程配置
func loadRemoteConfig() {
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
