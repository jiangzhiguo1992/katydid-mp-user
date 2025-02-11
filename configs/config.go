package configs

import (
	"bytes"
	_ "embed"
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
	confDir = "./configs" // configs根目录
	envKey  = "env"       // 环境key
)

type (
	Config struct {
		*ModuleConf `mapstructure:",squash"`
		*ServerConf `mapstructure:",squash"`
		PgSql       *PgSqlConf `toml:"pgsql" mapstructure:"pgsql"`
		Redis       *RedisConf `toml:"redis" mapstructure:"redis"`

		account *AccountConf `toml:"account" mapstructure:"account"`
		client  *ClientConf  `toml:"client" mapstructure:"client"`
		user    *UserConf    `toml:"user" mapstructure:"user"`
	}

	ModuleConf struct {
		Enable  bool   `toml:"enable" mapstructure:"enable"`
		Env     string `toml:"env" mapstructure:"env"`
		EnvName string `toml:"env_name" mapstructure:"env_name"`
	}

	ServerConf struct {
	}

	AccountConf struct {
		*ModuleConf `mapstructure:",squash"`
		*ServerConf `mapstructure:",squash"`
		PgSql       *PgSqlConf `toml:"pgsql" mapstructure:"pgsql"`
		Redis       *RedisConf `toml:"redis" mapstructure:"redis"`
	}

	ClientConf struct {
		*ModuleConf `mapstructure:",squash"`
		*ServerConf `mapstructure:",squash"`
		PgSql       *PgSqlConf `toml:"pgsql" mapstructure:"pgsql"`
		Redis       *RedisConf `toml:"redis" mapstructure:"redis"`
	}

	UserConf struct {
		*ModuleConf `mapstructure:",squash"`
		*ServerConf `mapstructure:",squash"`
		PgSql       *PgSqlConf `toml:"pgsql" mapstructure:"pgsql"`
		Redis       *RedisConf `toml:"redis" mapstructure:"redis"`
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
)

func (m *ModuleConf) IsProd() bool {
	return m.Env == "pro"
}

func LoadConfig() *Config {
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
	log.Printf("config: %v", settings)

	// 解析到config结构体
	err = viper.Unmarshal(&config.ModuleConf)
	if err != nil {
		log.Fatalf("unable to decode into mudule, %v", err)
	}
	err = viper.Unmarshal(&config.ServerConf)
	if err != nil {
		log.Fatalf("unable to decode into server, %v", err)
	}
	err = viper.UnmarshalKey("pgsql", &config.PgSql)
	if err != nil {
		log.Fatalf("unable to decode into pgsql, %v", err)
	}
	err = viper.UnmarshalKey("pgsql.write", &config.PgSql.Write)
	if err != nil {
		log.Fatalf("unable to decode into pgsql.write, %v", err)
	}
	err = viper.UnmarshalKey("pgsql.read", &config.PgSql.Read)
	if err != nil {
		log.Fatalf("unable to decode into pgsql.read, %v", err)
	}
	err = viper.UnmarshalKey("redis", &config.Redis)
	if err != nil {
		log.Fatalf("unable to decode into redis, %v", err)
	}

	// 解析到config.account结构体
	config.account = &AccountConf{}
	err = viper.Unmarshal(&config.account.ModuleConf)
	if err != nil {
		log.Fatalf("unable to decode into account.mudule, %v", err)
	}
	if config.account.ModuleConf == nil {
		config.account.ModuleConf = config.ModuleConf
	}
	err = viper.Unmarshal(&config.account.ServerConf)
	if err != nil {
		log.Fatalf("unable to decode into account.server, %v", err)
	}
	if config.account.ServerConf == nil {
		config.account.ServerConf = config.ServerConf
	}
	err = viper.UnmarshalKey("account.pgsql", &config.account.PgSql)
	if err != nil {
		log.Fatalf("unable to decode into account.pgsql, %v", err)
	}
	if config.account.PgSql == nil {
		config.account.PgSql = config.PgSql
	}
	err = viper.UnmarshalKey("account.pgsql.write", &config.account.PgSql.Write)
	if err != nil {
		log.Fatalf("unable to decode into account.pgsql.write, %v", err)
	}
	if &config.account.PgSql.Write == nil {
		config.account.PgSql.Write = config.PgSql.Write
	}
	err = viper.UnmarshalKey("account.pgsql.read", &config.account.PgSql.Read)
	if err != nil {
		log.Fatalf("unable to decode into account.pgsql.read, %v", err)
	}
	if &config.account.PgSql.Read == nil {
		config.account.PgSql.Read = config.PgSql.Read
	}
	err = viper.UnmarshalKey("account.redis", &config.account.Redis)
	if err != nil {
		log.Fatalf("unable to decode into account.redis, %v", err)
	}
	if config.account.Redis == nil {
		config.account.Redis = config.Redis
	}

	// 解析到config.client结构体
	config.client = &ClientConf{}
	err = viper.Unmarshal(&config.client.ModuleConf)
	if err != nil {
		log.Fatalf("unable to decode into client.mudule, %v", err)
	}
	if config.client.ModuleConf == nil {
		config.client.ModuleConf = config.ModuleConf
	}
	err = viper.Unmarshal(&config.client.ServerConf)
	if err != nil {
		log.Fatalf("unable to decode into client.server, %v", err)
	}
	if config.client.ServerConf == nil {
		config.client.ServerConf = config.ServerConf
	}
	err = viper.UnmarshalKey("client.pgsql", &config.client.PgSql)
	if err != nil {
		log.Fatalf("unable to decode into client.pgsql, %v", err)
	}
	if config.client.PgSql == nil {
		config.client.PgSql = config.PgSql
	}
	err = viper.UnmarshalKey("client.pgsql.write", &config.client.PgSql.Write)
	if err != nil {
		log.Fatalf("unable to decode into client.pgsql.write, %v", err)
	}
	if &config.client.PgSql.Write == nil {
		config.client.PgSql.Write = config.PgSql.Write
	}
	err = viper.UnmarshalKey("client.pgsql.read", &config.client.PgSql.Read)
	if err != nil {
		log.Fatalf("unable to decode into client.pgsql.read, %v", err)
	}
	if &config.client.PgSql.Read == nil {
		config.client.PgSql.Read = config.PgSql.Read
	}
	err = viper.UnmarshalKey("client.redis", &config.client.Redis)
	if err != nil {
		log.Fatalf("unable to decode into client.redis, %v", err)
	}
	if config.client.Redis == nil {
		config.client.Redis = config.Redis
	}

	// 解析到config.user结构体
	config.user = &UserConf{}
	err = viper.Unmarshal(&config.user.ModuleConf)
	if err != nil {
		log.Fatalf("unable to decode into user.mudule, %v", err)
	}
	if config.user.ModuleConf == nil {
		config.user.ModuleConf = config.ModuleConf
	}
	err = viper.Unmarshal(&config.user.ServerConf)
	if err != nil {
		log.Fatalf("unable to decode into user.server, %v", err)
	}
	if config.user.ServerConf == nil {
		config.user.ServerConf = config.ServerConf
	}
	err = viper.UnmarshalKey("user.pgsql", &config.user.PgSql)
	if err != nil {
		log.Fatalf("unable to decode into user.pgsql, %v", err)
	}
	if config.user.PgSql == nil {
		config.user.PgSql = config.PgSql
	}
	err = viper.UnmarshalKey("user.pgsql.write", &config.user.PgSql.Write)
	if err != nil {
		log.Fatalf("unable to decode into user.pgsql.write, %v", err)
	}
	if &config.user.PgSql.Write == nil {
		config.user.PgSql.Write = config.PgSql.Write
	}
	err = viper.UnmarshalKey("user.pgsql.read", &config.user.PgSql.Read)
	if err != nil {
		log.Fatalf("unable to decode into user.pgsql.read, %v", err)
	}
	if &config.user.PgSql.Read == nil {
		config.user.PgSql.Read = config.PgSql.Read
	}
	err = viper.UnmarshalKey("user.redis", &config.user.Redis)
	if err != nil {
		log.Fatalf("unable to decode into user.redis, %v", err)
	}
	if config.user.Redis == nil {
		config.account.Redis = config.Redis
	}

	log.Printf("config: %v", config)

	// TODO:GG 
	//viper.RemoteConfig.WatchChannel()

	return config
}
