package configs

import (
	_ "embed"
	_ "github.com/spf13/viper/remote"
	"time"
)

var (
	//go:embed app/init.toml
	fileAppInit []byte

	//go:embed app/public.toml
	fileAppPub []byte

	//go:embed app/private.toml
	fileAppPri []byte

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

		DefLang string `toml:"def_lang" mapstructure:"def_lang"`

		LogConf `toml:"log" mapstructure:"log"`

		AppConf `mapstructure:",squash"`

		Account AccountConf `toml:"account" mapstructure:"account"`
		Client  ClientConf  `toml:"client" mapstructure:"client"`
		User    UserConf    `toml:"user" mapstructure:"user"`
	}

	LogConf struct {
		OutEnable         bool   `toml:"out_enable" mapstructure:"out_enable"`
		OutLevel          int    `toml:"out_level" mapstructure:"out_level"`
		OutFormat         string `toml:"out_format" mapstructure:"out_format"`
		FileCheckInterval int    `toml:"file_check_interval" mapstructure:"file_check_interval"`
		FileMaxAge        int    `toml:"file_max_age" mapstructure:"file_max_age"`
		FileMaxSize       int64  `toml:"file_max_size" mapstructure:"file_max_size"`
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

func (m *Config) IsDebug() bool {
	return (m.Env == "dev") || (m.Env == "fat")
}

func (m *Config) IsTest() bool {
	return m.Env == "uat"
}

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
