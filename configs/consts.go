package configs

const (
	// ConfDir configs根目录
	ConfDir = "./configs"

	// LogDir logs输出目录
	LogDir = "logs"
)

var (
	// ConfigIgnoreDir 忽略的配置目录 TODO:GG 需要配置
	ConfigIgnoreDir = []string{"perm"}

	// LangDirs 本地化文件目录
	LangDirs = []string{"./assets/i18n", "./assets/i18n/locales"}
)
