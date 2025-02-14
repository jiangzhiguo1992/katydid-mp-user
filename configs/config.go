package configs

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	//go:embed app/init.toml
	fileAppInit []byte

	//go:embed app/private.toml
	fileAppPri []byte

	//go:embed app/public.toml
	fileAppPub []byte

	config = new(Config)
)

const (
	envKey = "env" // 环境key

	reloadMaxRetries = 3               // 重试次数
	reloadInterval   = 2 * time.Second // 重试间隔
)

type (
	Config struct {
		Env     string `toml:"env" mapstructure:"env"`
		EnvName string `toml:"env_name" mapstructure:"env_name"`

		LogFileNameFormat string `toml:"log_file_name_format" mapstructure:"log_file_name_format"`
		LogFileLevel      int    `toml:"log_file_level" mapstructure:"log_file_level"`

		DefLang string `toml:"def_lang" mapstructure:"def_lang"`

		AppConf `mapstructure:",squash"`

		Account AccountConf `toml:"account" mapstructure:"account"`
		Client  ClientConf  `toml:"client" mapstructure:"client"`
		User    UserConf    `toml:"user" mapstructure:"user"`
	}

	AppConf struct {
		ModuleConf `mapstructure:",squash"`
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
		Server  Server      `toml:"server" mapstructure:"server"`
		PgSql   PgSqlConf   `toml:"pgsql" mapstructure:"pgsql"`
		Redis   RedisConf   `toml:"redis"  mapstructure:"redis"`
		MongoDB MongoDBConf `toml:"mongo_db"  mapstructure:"mongo_db"`
	}

	Server struct {
		ApiDomain    string `toml:"api_domain" mapstructure:"api_domain"`
		ApiHttpPort  string `toml:"api_http_port" mapstructure:"api_http_port"`
		ApiHttpsPort string `toml:"api_https_port" mapstructure:"api_https_port"`
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

func Get() *Config {
	return config
}

func Init(confDir string, reload func() bool) *Config {
	loadLocalConfig(confDir, reload)
	loadRemoteConfig()
	return config
}

func loadLocalConfig(confDir string, reload func() bool) {
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
	log.Printf("config.settings init OK: %v", settings)

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

	log.Printf("config init OK: %v", config)

	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("config file changed name:%s\n", e.Name)
		if reload != nil {
			ctx := context.Background()
			for i := 0; i < reloadMaxRetries; i++ {
				if reload() {
					break
				}
				log.Printf("reload config failed, retry: %d", i)
				//time.Sleep(retryInterval)
				if i >= reloadMaxRetries-1 {
					log.Fatalf("reload config failed, max retries: %d", reloadMaxRetries)
				}
				select {
				case <-time.After(reloadInterval):
				case <-ctx.Done():
					break
				}
			}
		}
	})
	viper.WatchConfig()
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
