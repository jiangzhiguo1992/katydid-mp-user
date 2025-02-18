package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log/slog"
	"os"
	"path"
	"time"
)

const (
	// 颜色代码
	//colorRed    = "\x1b[31m"
	//colorGreen  = "\x1b[32m"
	//colorYellow = "\x1b[33m"
	//colorBlue   = "\x1b[34m"
	//colorPurple = "\x1b[35m"
	//colorCyan   = "\x1b[36m"
	//colorGray   = "\x1b[37m"

	colorReset = "\x1b[0m"

	// 背景色代码
	bgBlack  = "\x1b[40m"
	bgRed    = "\x1b[41m"
	bgGreen  = "\x1b[42m"
	bgYellow = "\x1b[43m"
	bgPurple = "\x1b[45m"
	bgGray   = "\x1b[47m"
)

// 日志级别映射
var levelColors = map[zapcore.Level]struct {
	bg   string
	text string
}{
	zapcore.DebugLevel:  {bgGray, "DEBUG"},
	zapcore.InfoLevel:   {bgGreen, "INFO"},
	zapcore.WarnLevel:   {bgYellow, "WARN"},
	zapcore.ErrorLevel:  {bgRed, "ERROR"},
	zapcore.DPanicLevel: {bgPurple, "DPANIC"},
	zapcore.PanicLevel:  {bgPurple, "PANIC"},
	zapcore.FatalLevel:  {bgBlack, "FATAL"},
}

type levelConfig struct {
	dir    string
	enable func(zapcore.Level) bool
}

var levelConfigs = map[int][]levelConfig{
	0: {{"fat", func(lv zapcore.Level) bool { return lv >= zapcore.FatalLevel }}},
	1: {{"pac", func(lv zapcore.Level) bool { return lv >= zapcore.DPanicLevel && lv < zapcore.FatalLevel }}},
	2: {{"err", func(lv zapcore.Level) bool { return lv >= zapcore.ErrorLevel && lv < zapcore.DPanicLevel }}},
	3: {{"warn", func(lv zapcore.Level) bool { return lv >= zapcore.WarnLevel && lv < zapcore.ErrorLevel }}},
	4: {{"info", func(lv zapcore.Level) bool { return lv >= zapcore.InfoLevel && lv < zapcore.WarnLevel }}},
	5: {{"debug", func(lv zapcore.Level) bool { return lv < zapcore.InfoLevel }}},
}

// Config 日志配置
type Config struct {
	OutEnable bool           // 是否启用输出
	OutDir    *string        // 输出目录
	OutLevel  *int           // 输出级别
	OutFormat *string        // 输出格式
	CheckInt  *time.Duration // 检查间隔
	MaxAge    *time.Duration // 最大时间
	MaxSize   *int64         // 最大大小
}

var (
	cfg     Config
	buffers []*BufferedWriteSyncer
	logger  *zap.Logger
)

// Init 初始化日志
func Init(config Config) {
	cfg = config

	// encoder
	encodeCfg := zap.NewProductionEncoderConfig()
	encodeCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	if cfg.OutEnable {
		encodeCfg.LevelKey = ""
	} else {
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
	var encoder zapcore.Encoder
	if cfg.OutEnable {
		encoder = zapcore.NewJSONEncoder(encodeCfg)
	} else {
		//encoder = &customConsoleEncoder{zapcore.NewConsoleEncoder(encodeCfg)}
		encoder = zapcore.NewConsoleEncoder(encodeCfg)
	}

	if cfg.OutEnable {
		// outfile
		var cores []zapcore.Core
		if cfg.OutEnable && (cfg.OutDir != nil) && (cfg.OutLevel != nil) {
			// cores
			for level := 0; level <= *cfg.OutLevel; level++ {
				if configs, ok := levelConfigs[level]; ok {
					for _, config := range configs {
						dir := path.Join(*cfg.OutDir, config.dir)
						if err := os.MkdirAll(dir, os.ModePerm); err != nil {
							panic(fmt.Errorf("failed to create dir %s: %w", config.dir, err))
						}
						// writer
						writer := NewDateWriteSyncer(
							&dir,
							cfg.OutFormat,
							cfg.CheckInt,
							cfg.MaxAge,
							cfg.MaxSize,
						)
						buffers = append(buffers, NewBufferedWriteSyncer(writer))
						//buffered := NewBufferedWriteSyncer(writer)
						//defer func(buffered *BufferedWriteSyncer) {
						//	_ = buffered.Close()
						//}(buffered)
						//writer := &DateWriteSyncer{outPath: dir, format: cfg.OutFormat}
						writeSyncer := zapcore.Lock(zapcore.AddSync(writer))
						core := zapcore.NewCore(encoder, writeSyncer, zap.LevelEnablerFunc(config.enable))
						cores = append(cores, core)
					}
				}
			}
		}

		// logger
		core := zapcore.NewTee(cores...)
		logger = zap.New(core)
	} else {
		// production config
		c := zap.NewProductionConfig()
		c.Development = false
		c.Encoding = "console"
		c.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		c.Sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
		c.DisableCaller = true
		c.EncoderConfig = encodeCfg

		// logger
		var err error
		logger, err = c.Build()
		if err != nil {
			panic(err)
		}
	}
}

// OnExit 退出日志
func OnExit() {
	if logger == nil {
		return
	}
	for _, v := range buffers {
		_ = v.Close()
	}
	_ = logger.Sync()
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

	if !cfg.OutEnable {
		if color, ok := levelColors[level]; ok {
			msg = fmt.Sprintf("%s%s%s", color.text, msg, colorReset)
		}
	}

	switch level {
	case zapcore.DebugLevel:
		logger.Debug(msg, fields...)
	case zapcore.InfoLevel:
		logger.Info(msg, fields...)
	case zapcore.WarnLevel:
		logger.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		logger.Error(msg, fields...)
	case zapcore.PanicLevel:
		logger.Panic(msg, fields...)
	case zapcore.FatalLevel:
		logger.Fatal(msg, fields...)
	default:
		panic("unhandled default case")
	}
}

func Debug(msg string, fields ...zap.Field) { log(zapcore.DebugLevel, msg, fields...) }
func Info(msg string, fields ...zap.Field)  { log(zapcore.InfoLevel, msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { log(zapcore.WarnLevel, msg, fields...) }
func Error(msg string, fields ...zap.Field) { log(zapcore.ErrorLevel, msg, fields...) }
func Panic(msg string, fields ...zap.Field) { log(zapcore.PanicLevel, msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { log(zapcore.FatalLevel, msg, fields...) }