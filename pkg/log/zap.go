package log

import (
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

var (
	logger *Logger
	once   sync.Once
)

// Logger 封装日志实例
type Logger struct {
	zap    *zap.Logger
	config *Config
	syncs  []*zapcore.BufferedWriteSyncer
	mu     sync.RWMutex
}

// Config 日志配置
type Config struct {
	OutLevel  int           // 输出级别
	OutDir    string        // 输出目录
	OutFormat string        // 输出格式
	CheckInt  time.Duration // 检查间隔
	MaxAge    time.Duration // 最大时间
	MaxSize   int64         // 最大大小
}

func NewDefaultConfig(outLevel int) Config {
	return Config{
		OutLevel:  outLevel,
		OutDir:    defaultOutPath,
		OutFormat: defaultFormat,
		CheckInt:  defaultCheckInterval,
		MaxAge:    defaultMaxAge,
		MaxSize:   defaultMaxSize,
	}
}

// Init 初始化日志
func Init(cfg Config) {
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

	var errs []error
	for _, buffer := range logger.syncs {
		if err := buffer.Stop(); err != nil {
			errs = append(errs, err)
		}
	}

	if err := logger.zap.Sync(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("■ ■ Log ■ ■ failed to close logger: %v", errs)
	}
	return nil
}

// newLogger 创建新的日志实例
func newLogger(cfg Config) *Logger {
	l := &Logger{config: &cfg}
	l.initialize()
	return l
}

func (c *Config) outEnable() bool {
	return c.OutLevel > 0
}

func (l *Logger) initialize() {
	// encoder
	var encoder zapcore.Encoder
	encoderCfg := createEncoderConfig(l.config)
	if l.config.outEnable() {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	if l.config.outEnable() {
		// cores
		var cores []zapcore.Core
		for level := 1; level <= l.config.OutLevel; level++ {
			if configs, ok := levelConfigs[level]; ok {
				for _, config := range configs {
					core, syncer := createCore(encoder, l.config, config)
					l.syncs = append(l.syncs, syncer)
					cores = append(cores, core)
				}
			}
		}

		// logger
		core := zapcore.NewTee(cores...)
		l.zap = zap.New(
			core,
			zap.AddStacktrace(zap.ErrorLevel),
			zap.AddCallerSkip(2),
		)
	} else {
		// production config
		c := zap.NewProductionConfig()
		c.Development = true
		c.Encoding = "console"
		c.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		c.Sampling = nil                   // 禁用采样
		c.OutputPaths = []string{"stdout"} // 及时打印
		c.DisableStacktrace = true
		c.DisableCaller = true
		c.EncoderConfig = encoderCfg

		// logger
		var err error
		l.zap, err = c.Build(
			zap.WithClock(zapcore.DefaultClock), // 使用默认时钟
			zap.AddStacktrace(zap.ErrorLevel),   // 只在 error 级别添加堆栈
			zap.AddCallerSkip(2),
		)
		if err != nil {
			panic(err)
		}
	}
}

func createEncoderConfig(config *Config) zapcore.EncoderConfig {
	encodeCfg := zap.NewProductionEncoderConfig()
	encodeCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	if config.outEnable() {
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

func createCore(encoder zapcore.Encoder, config *Config, lConf levelConfig) (zapcore.Core, *zapcore.BufferedWriteSyncer) {
	// path
	dir := path.Join(config.OutDir, lConf.dir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Errorf("■ ■ Log ■ ■ failed to create dir %s: %w", lConf.dir, err))
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

func log(level zapcore.Level, msg string, fields ...zap.Field) {
	if logger == nil {
		switch level {
		case zapcore.DebugLevel:
			slog.Debug(msg)
		case zapcore.InfoLevel:
			slog.Info(msg)
		case zapcore.WarnLevel:
			slog.Warn(msg)
		default:
			slog.Error(msg)
		}
		return
	}

	if !logger.config.outEnable() {
		if color, ok := levelColors[level]; ok {
			msg = fmt.Sprintf("%s%s%s", color.fg, msg, colorReset)
		}
	}

	switch level {
	case zapcore.DebugLevel:
		logger.zap.Debug(msg, fields...)
	case zapcore.InfoLevel:
		logger.zap.Info(msg, fields...)
	case zapcore.WarnLevel:
		logger.zap.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		logger.zap.Error(msg, fields...)
	case zapcore.PanicLevel:
		logger.zap.Panic(msg, fields...)
	case zapcore.FatalLevel:
		logger.zap.Fatal(msg, fields...)
	default:
		panic("unhandled default case")
	}
}

func Debug(msg string, fields ...Field) {
	log(zapcore.DebugLevel, msg, toZapFields(fields)...)
}

func Info(msg string, fields ...Field) {
	log(zapcore.InfoLevel, msg, toZapFields(fields)...)
}

func Warn(msg string, fields ...Field) {
	log(zapcore.WarnLevel, msg, toZapFields(fields)...)
}

func Error(msg string, fields ...Field) {
	log(zapcore.ErrorLevel, msg, toZapFields(fields)...)
}

func Panic(msg string, fields ...Field) {
	log(zapcore.PanicLevel, msg, toZapFields(fields)...)
}

func Fatal(msg string, fields ...Field) {
	log(zapcore.FatalLevel, msg, toZapFields(fields)...)
}
