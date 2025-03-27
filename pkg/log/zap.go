package log

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"
)

const (
	// 颜色代码
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorPurple = "\x1b[35m"
	colorGray   = "\x1b[37m"

	colorReset = "\x1b[0m"

	// 背景色代码
	bgRed    = "\x1b[41m"
	bgGreen  = "\x1b[42m"
	bgYellow = "\x1b[43m"
	bgPurple = "\x1b[45m"
	bgGray   = "\x1b[47m"

	// writer默认conf
	defaultOutPath             = "logs"             // 默认输出目录
	defaultFormat              = "06-01-02"         // 默认文件名格式
	defaultCheckInterval       = time.Hour          // 定期清理时间间隔
	defaultMaxAge              = 7 * 24 * time.Hour // 7天
	defaultMaxSize       int64 = 100 << 20          // 100MB

	// buffer默认配置
	defaultBatchSize    = 1024 * 4        // 缓冲区大小设置为 4kb
	defaultFlushTimeout = 5 * time.Second // 缓冲区刷新间隔设置为 5 秒
)

// 日志级别映射
var levelColors = map[zapcore.Level]struct {
	fg   string
	bg   string
	text string
}{
	zapcore.DebugLevel:  {colorGray, bgGray, "DEBUG"},
	zapcore.InfoLevel:   {colorGreen, bgGreen, "INFO"},
	zapcore.WarnLevel:   {colorYellow, bgYellow, "WARN"},
	zapcore.ErrorLevel:  {colorRed, bgRed, "ERROR"},
	zapcore.DPanicLevel: {colorPurple, bgPurple, "DPANIC"},
	zapcore.PanicLevel:  {colorPurple, bgPurple, "PANIC"},
	zapcore.FatalLevel:  {colorRed, colorPurple, "FATAL"},
}

type levelConfig struct {
	dir    string
	enable func(zapcore.Level) bool
}

var levelConfigs = map[int][]levelConfig{
	1: {{"fat", func(lv zapcore.Level) bool { return lv >= zapcore.FatalLevel }}},
	2: {{"pac", func(lv zapcore.Level) bool { return lv >= zapcore.DPanicLevel && lv < zapcore.FatalLevel }}},
	3: {{"err", func(lv zapcore.Level) bool { return lv >= zapcore.ErrorLevel && lv < zapcore.DPanicLevel }}},
	4: {{"warn", func(lv zapcore.Level) bool { return lv >= zapcore.WarnLevel && lv < zapcore.ErrorLevel }}},
	5: {{"info", func(lv zapcore.Level) bool { return lv >= zapcore.InfoLevel && lv < zapcore.WarnLevel }}},
	6: {{"debug", func(lv zapcore.Level) bool { return lv < zapcore.InfoLevel }}},
}

// 定义日志级别映射常量
const (
	LevelFatal = 1
	LevelPanic = 2
	LevelError = 3
	LevelWarn  = 4
	LevelInfo  = 5
	LevelDebug = 6
)

// 日志级别映射表
var levelMapping = map[zapcore.Level]int{
	zapcore.FatalLevel:  LevelFatal,
	zapcore.PanicLevel:  LevelPanic,
	zapcore.DPanicLevel: LevelPanic,
	zapcore.ErrorLevel:  LevelError,
	zapcore.WarnLevel:   LevelWarn,
	zapcore.InfoLevel:   LevelInfo,
	zapcore.DebugLevel:  LevelDebug,
}

var (
	logger *Logger
	once   sync.Once
)

// Logger 封装日志实例
type Logger struct {
	console *zap.Logger // 控制台
	output  *zap.Logger // 写文件

	config *Config
	syncs  []*zapcore.BufferedWriteSyncer
	mu     sync.RWMutex
}

// Config 日志配置
type Config struct {
	ConLevels []int         // 打印级别
	OutLevels []int         // 输出级别
	OutDir    string        // 输出目录
	OutFormat string        // 输出格式
	CheckInt  time.Duration // 检查间隔
	MaxAge    time.Duration // 最大时间
	MaxSize   int64         // 最大大小
}

func NewDefaultConfig(conLevels, outLevels []int) Config {
	return Config{
		ConLevels: conLevels,
		OutLevels: outLevels,
		OutDir:    defaultOutPath,
		OutFormat: defaultFormat,
		CheckInt:  defaultCheckInterval,
		MaxAge:    defaultMaxAge,
		MaxSize:   defaultMaxSize,
	}
}

// Init 初始化日志
func Init(cfg Config) {
	marshal, _ := json.MarshalIndent(cfg, "", "\t")
	slog.Info(fmt.Sprintf("■ ■ Log ■ ■ 配置 ---> %s", marshal))

	once.Do(func() {
		logger = newLogger(cfg)
	})
}

// Close 优雅退出
func Close() error {
	if logger == nil {
		return nil
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()

	// 关闭所有缓冲写入器
	var errs []error
	for _, buffer := range logger.syncs {
		if err := buffer.Stop(); err != nil {
			errs = append(errs, err)
		}
	}

	// 同步日志
	if err := logger.console.Sync(); err != nil {
		errs = append(errs, err)
	}
	if err := logger.output.Sync(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("■ ■ Log ■ ■ 关闭失败: %v", errs)
	}
	return nil
}

// newLogger 创建新的日志实例
func newLogger(cfg Config) *Logger {
	l := &Logger{config: &cfg}
	l.initializeConsole()
	l.initializeOutput()
	return l
}

func (l *Logger) initializeConsole() {
	c := zap.NewProductionConfig()
	c.Development = true
	c.Encoding = "console"
	c.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	c.Sampling = nil                   // 禁用采样
	c.OutputPaths = []string{"stdout"} // 及时打印
	c.DisableStacktrace = true
	c.DisableCaller = true
	c.EncoderConfig = createEncoderConfig(false)

	// logger
	var err error
	l.console, err = c.Build(
		zap.WithClock(zapcore.DefaultClock), // 使用默认时钟
		zap.AddStacktrace(zap.ErrorLevel),   // 只在 error 级别添加堆栈
		zap.AddCallerSkip(2),
	)
	if err != nil {
		panic(err)
	}
}

func (l *Logger) initializeOutput() {
	// encoder
	encoderCfg := createEncoderConfig(true)
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	// cores
	var cores []zapcore.Core
	for _, configs := range levelConfigs {
		for _, config := range configs {
			core, syncer := createOutputCore(encoder, l.config, config)
			l.syncs = append(l.syncs, syncer)
			cores = append(cores, core)
		}
	}

	// logger
	core := zapcore.NewTee(cores...)
	l.output = zap.New(
		core,
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(2),
	)
}

func createEncoderConfig(output bool) zapcore.EncoderConfig {
	encodeCfg := zap.NewProductionEncoderConfig()
	encodeCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	if output {
		encodeCfg.LevelKey = ""
	} else {
		encodeCfg.EncodeDuration = zapcore.StringDurationEncoder
		encodeCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("\x1b[20m%s\x1b[0m", t.Format("2006-01-02 15:04:05.000")))
		}
		//encodeCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encodeCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			if level, ok := levelColors[l]; ok {
				enc.AppendString(fmt.Sprintf("%s%s%s", level.bg, level.text, colorReset))
			}
		}
	}
	return encodeCfg
}

func createOutputCore(encoder zapcore.Encoder, config *Config, lConf levelConfig) (zapcore.Core, *zapcore.BufferedWriteSyncer) {
	// path
	dir := path.Join(config.OutDir, lConf.dir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Errorf("■ ■ Log ■ ■ 创建文件夹失败 %s: %w", lConf.dir, err))
	}
	// writer
	writer := NewDateWriteSyncer(
		dir,
		config.OutFormat,
		config.CheckInt,
		config.MaxAge,
		config.MaxSize,
	)

	// syncer
	syncer := &zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(writer),
		Size:          defaultBatchSize,
		FlushInterval: defaultFlushTimeout,
	}
	// core
	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(syncer),
		zap.LevelEnablerFunc(lConf.enable),
	)
	return core, syncer
}

func log(level zapcore.Level, output *bool, msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}

	if output == nil {
		matchLevel, ok := levelMapping[level]
		if !ok {
			return // 不支持的日志级别
		}

		// 检查是否需要输出到文件
		for _, v := range logger.config.OutLevels {
			if v == matchLevel {
				shouldOutput := true
				output = &shouldOutput
				break
			}
		}

		// 如果不输出到文件，检查是否需要输出到控制台
		if output == nil {
			for _, v := range logger.config.ConLevels {
				if v == matchLevel {
					shouldOutput := false
					output = &shouldOutput
					break
				}
			}
		}

		// 不需要输出
		if output == nil {
			return
		}
	}

	// 使用适当的日志记录器
	lg := logger.console
	if *output {
		lg = logger.output
	} else {
		if color, ok := levelColors[level]; ok {
			msg = fmt.Sprintf("%s%s%s", color.fg, msg, colorReset)
		}
	}

	// 记录日志
	switch level {
	case zapcore.DebugLevel:
		lg.Debug(msg, fields...)
	case zapcore.InfoLevel:
		lg.Info(msg, fields...)
	case zapcore.WarnLevel:
		lg.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		lg.Error(msg, fields...)
	case zapcore.DPanicLevel:
		lg.DPanic(msg, fields...)
	case zapcore.PanicLevel:
		lg.Panic(msg, fields...)
	case zapcore.FatalLevel:
		lg.Fatal(msg, fields...)
	default:
		panic("■ ■ Log ■ ■ 打印级别找不到")
	}
}

func Debug(msg string, fields ...Field) {
	log(zapcore.DebugLevel, nil, msg, toZapFields(fields)...)
}

func DebugOutput(msg string, output bool, fields ...Field) {
	log(zapcore.DebugLevel, &output, msg, toZapFields(fields)...)
}

func DebugFmt(msg string, a ...any) {
	log(zapcore.DebugLevel, nil, fmt.Sprintf(msg, a...))
}

func DebugFmtOutput(msg string, output bool, a ...any) {
	log(zapcore.DebugLevel, &output, fmt.Sprintf(msg, a...))
}

func Info(msg string, fields ...Field) {
	log(zapcore.InfoLevel, nil, msg, toZapFields(fields)...)
}

func InfoOutput(msg string, output bool, fields ...Field) {
	log(zapcore.InfoLevel, &output, msg, toZapFields(fields)...)
}

func InfoFmt(msg string, a ...any) {
	log(zapcore.InfoLevel, nil, fmt.Sprintf(msg, a...))
}

func InfoFmtOutput(msg string, output bool, a ...any) {
	log(zapcore.InfoLevel, &output, fmt.Sprintf(msg, a...))
}

func Warn(msg string, fields ...Field) {
	log(zapcore.WarnLevel, nil, msg, toZapFields(fields)...)
}

func WarnOutput(msg string, output bool, fields ...Field) {
	log(zapcore.WarnLevel, &output, msg, toZapFields(fields)...)
}

func WarnFmt(msg string, a ...any) {
	log(zapcore.WarnLevel, nil, fmt.Sprintf(msg, a...))
}

func WarnFmtOutput(msg string, output bool, a ...any) {
	log(zapcore.WarnLevel, &output, fmt.Sprintf(msg, a...))
}

func Error(msg string, fields ...Field) {
	log(zapcore.ErrorLevel, nil, msg, toZapFields(fields)...)
}

func ErrorOutput(msg string, output bool, fields ...Field) {
	log(zapcore.ErrorLevel, &output, msg, toZapFields(fields)...)
}

func ErrorFmt(msg string, a ...any) {
	log(zapcore.ErrorLevel, nil, fmt.Sprintf(msg, a...))
}

func ErrorFmtOutput(msg string, output bool, a ...any) {
	log(zapcore.ErrorLevel, &output, fmt.Sprintf(msg, a...))
}

func DPanic(msg string, fields ...Field) {
	log(zapcore.DPanicLevel, nil, msg, toZapFields(fields)...)
}

func DPanicOutput(msg string, output bool, fields ...Field) {
	log(zapcore.DPanicLevel, &output, msg, toZapFields(fields)...)
}

func DPanicFmt(msg string, a ...any) {
	log(zapcore.DPanicLevel, nil, fmt.Sprintf(msg, a...))
}

func DPanicFmtOutput(msg string, output bool, a ...any) {
	log(zapcore.DPanicLevel, &output, fmt.Sprintf(msg, a...))
}

func Panic(msg string, fields ...Field) {
	log(zapcore.PanicLevel, nil, msg, toZapFields(fields)...)
}

func PanicOutput(msg string, output bool, fields ...Field) {
	log(zapcore.PanicLevel, &output, msg, toZapFields(fields)...)
}

func PanicFmt(msg string, a ...any) {
	log(zapcore.PanicLevel, nil, fmt.Sprintf(msg, a...))
}

func PanicFmtOutput(msg string, output bool, a ...any) {
	log(zapcore.PanicLevel, &output, fmt.Sprintf(msg, a...))
}

func Fatal(msg string, fields ...Field) {
	log(zapcore.FatalLevel, nil, msg, toZapFields(fields)...)
}

func FatalOutput(msg string, output bool, fields ...Field) {
	log(zapcore.FatalLevel, &output, msg, toZapFields(fields)...)
}

func FatalFmt(msg string, a ...any) {
	log(zapcore.FatalLevel, nil, fmt.Sprintf(msg, a...))
}

func FatalFmtOutput(msg string, output bool, a ...any) {
	log(zapcore.FatalLevel, &output, fmt.Sprintf(msg, a...))
}
