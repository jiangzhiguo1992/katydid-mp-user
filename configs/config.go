package configs

type (
	Config struct {
		Env     string `toml:"env" mapstructure:"env"`
		EnvName string `toml:"env_name" mapstructure:"env_name"`

		RemoteConf `toml:"remote_conf" mapstructure:"remote_conf"`

		LogConf  `toml:"log" mapstructure:"log"`
		LangConf `toml:"lang" mapstructure:"lang"`

		AppConf `mapstructure:",squash"`

		Auth   AuthConf   `toml:"auth" mapstructure:"auth"`
		Client ClientConf `toml:"client" mapstructure:"client"`
		User   UserConf   `toml:"user" mapstructure:"user"`
	}

	RemoteConf struct {
		Enabled         bool   `toml:"enabled" mapstructure:"enabled"`
		Provider        string `toml:"provider" mapstructure:"provider"` // etcd, consul, firestore
		Endpoint        string `toml:"endpoint" mapstructure:"endpoint"`
		Path            string `toml:"path" mapstructure:"path"`
		SecretKey       string `toml:"secret_key" mapstructure:"secret_key"`
		RefreshInterval int    `toml:"refresh_interval" mapstructure:"refresh_interval"`
	}

	LogConf struct {
		ConLevels     []int  `toml:"con_levels" mapstructure:"con_levels"`
		OutLevels     []int  `toml:"out_levels" mapstructure:"out_levels"`
		OutFormat     string `toml:"out_format" mapstructure:"out_format"`
		CheckInterval int    `toml:"check_interval" mapstructure:"check_interval"`
		FileMaxAge    int    `toml:"file_max_age" mapstructure:"file_max_age"`
		FileMaxSize   int64  `toml:"file_max_size" mapstructure:"file_max_size"`
	}

	LangConf struct {
		Default string `toml:"default" mapstructure:"default"`
	}

	AppConf struct {
		Server Server `toml:"server" mapstructure:"server"`

		ModuleConf `mapstructure:",squash"`
	}

	Server struct {
		ApiDomain    string `toml:"api_domain" mapstructure:"api_domain"`
		ApiHttpPort  string `toml:"api_http_port" mapstructure:"api_http_port"`
		ApiHttpsPort string `toml:"api_https_port" mapstructure:"api_https_port"`
	}

	AuthConf struct {
		ModuleConf `mapstructure:",squash"`
	}

	ClientConf struct {
		ModuleConf `mapstructure:",squash"`
	}

	UserConf struct {
		ModuleConf `mapstructure:",squash"`
	}

	ModuleConf struct {
		Enable bool       `toml:"enable" mapstructure:"enable"`
		PgSql  *PgSqlConf `toml:"pgsql" mapstructure:"pgsql"`
		Redis  *RedisConf `toml:"redis"  mapstructure:"redis"`
		Mongo  *MongoConf `toml:"mongo"  mapstructure:"mongo"`
		// TODO:GG 其他诸如tidb等数据库配置
	}

	PgSqlConf struct {
		Write struct {
			Host   string `toml:"host" mapstructure:"host"`
			Port   int    `toml:"port" mapstructure:"port"`
			DBName string `toml:"db_name" mapstructure:"db_name"`
			User   string `toml:"user" mapstructure:"user"`
			Pwd    string `toml:"pwd" mapstructure:"pwd"`
		} `mapstructure:",squash"`
		// cluster
		Reads *struct {
			Host   []string            `toml:"host" mapstructure:"host"`
			Port   []int               `toml:"port" mapstructure:"port"`
			User   []string            `toml:"user" mapstructure:"user"`
			Pwd    []string            `toml:"pwd" mapstructure:"pwd"`
			Weight []int               `toml:"weight" mapstructure:"weight"`
			Params []map[string]string `toml:"params" mapstructure:"params"`
		} `toml:"reads"`
		// retry
		MaxRetries    int `toml:"max_retries" mapstructure:"max_retries"`
		RetryDelay    int `toml:"retry_delay" mapstructure:"retry_delay"`
		RetryMaxDelay int `toml:"retry_max_delay" mapstructure:"retry_max_delay"`
		// pool
		MaxOpen    int `toml:"max_open" mapstructure:"max_open"`
		MaxIdle    int `toml:"max_idle" mapstructure:"max_idle"`
		MaxLifeMin int `toml:"max_life_min" mapstructure:"max_life_min"`
		MaxIdleMin int `toml:"max_idle_min" mapstructure:"max_idle_min"`
		// health
		HealthCheckInterval int  `toml:"health_check_interval" mapstructure:"health_check_interval"`
		AutoReconnect       bool `toml:"auto_reconnect" mapstructure:"auto_reconnect"`
		QueryTimeout        int  `toml:"query_timeout" mapstructure:"query_timeout"`
		// extra
		TimeZone string `toml:"timezone" mapstructure:"timezone"`
		SSLMode  string `toml:"ssl_mode" mapstructure:"ssl_mode"`
	}

	RedisConf struct {
		Host       string   `toml:"host" mapstructure:"host"`
		Port       int      `toml:"port" mapstructure:"port"`
		DBName     string   `toml:"db_name" mapstructure:"db_name"`
		Pwd        string   `toml:"pwd" mapstructure:"pwd"`
		MaxRetries int      `toml:"max_retries" mapstructure:"max_retries"`
		PoolSize   int      `toml:"pool_size" mapstructure:"pool_size"`
		MinIdle    int      `toml:"min_idle" mapstructure:"min_idle"`
		Clusters   []string `toml:"clusters" mapstructure:"clusters"`
	}

	MongoConf struct {
		// TODO:GG
	}
)

// IsDebug 调试模式 (本地开发)
func (m *Config) IsDebug() bool {
	return (m.Env == "dev") || (m.Env == "fat")
}

// IsTest 测试模式 (云测试环境)
func (m *Config) IsTest() bool {
	return m.Env == "uat"
}

// IsProd 生产模式 (云线上环境)
func (m *Config) IsProd() bool {
	return m.Env == "pro"
}

// merge 主要是为了覆盖默认配置，不好自动化处理，手动配置
func (m *Config) merge() {
	m.Auth.ModuleConf = m.AppConf.ModuleConf
	m.Client.ModuleConf = m.AppConf.ModuleConf
	m.User.ModuleConf = m.AppConf.ModuleConf
}
