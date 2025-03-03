package configs

//import _ "embed"

const (
	// ConfDir configs根目录
	ConfDir = "./configs"

	// LogDir logs输出目录
	LogDir = "logs"
)

var (
	//// ConfAppFiles 初始化加载的文件
	//ConfAppFiles = [][]byte{fileAppInit, fileAppPub, fileAppPri}
	////go:embed app/init.toml
	//fileAppInit []byte
	////go:embed app/public.toml
	//fileAppPub []byte
	////go:embed app/private.toml
	//fileAppPri []byte

	// LangDirs 本地化文件目录
	LangDirs = []string{"./assets/i18n", "./assets/i18n/locales"}
)
