package log

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"time"
)

var (
	logger *zap.Logger

	fileNameFormat = "2006-01-02"
)

// Init 初始化日志
func Init(prod bool, logDir, fNameFormat string) {
	fileNameFormat = fNameFormat
	// encoder
	encodeCfg := zap.NewProductionEncoderConfig()
	encodeCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	if prod {
		encodeCfg.LevelKey = ""
	} else {
		encodeCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encodeCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			switch l {
			default:
			case zapcore.DebugLevel:
				enc.AppendString("\x1b[34mDEBUG\x1b[0m") // 蓝色
			case zapcore.InfoLevel:
				enc.AppendString("\x1b[36mINFO\x1b[0m") // 青色
			case zapcore.WarnLevel:
				enc.AppendString("\x1b[33mWARN\x1b[0m") // 黄色
			case zapcore.ErrorLevel:
				enc.AppendString("\x1b[31mERROR\x1b[0m") // 红色
			case zapcore.DPanicLevel:
				enc.AppendString("\x1b[35mDPANIC\x1b[0m") // 紫色
			case zapcore.PanicLevel:
				enc.AppendString("\x1b[35mPANIC\x1b[0m") // 紫色
			case zapcore.FatalLevel:
				enc.AppendString("\x1b[31mFATAL\x1b[0m") // 红色
			}
		}
	}
	var encoder zapcore.Encoder
	if prod {
		encoder = zapcore.NewJSONEncoder(encodeCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encodeCfg)
	}

	if prod {
		// writer
		infoDir := path.Join(logDir, "info")
		warnDir := path.Join(logDir, "warn")
		errDir := path.Join(logDir, "err")
		pacDir := path.Join(logDir, "pac")
		fatDir := path.Join(logDir, "fat")
		if err := os.MkdirAll(infoDir, os.ModePerm); err != nil {
			panic(errors.New(fmt.Sprintf("failed to create log_info_dir %s: %s", infoDir, err)))
		}
		if err := os.MkdirAll(warnDir, os.ModePerm); err != nil {
			panic(errors.New(fmt.Sprintf("failed to create log_warn_dir %s: %s", warnDir, err)))
		}
		if err := os.MkdirAll(errDir, os.ModePerm); err != nil {
			panic(errors.New(fmt.Sprintf("failed to create log_err_dir %s: %s", errDir, err)))
		}
		if err := os.MkdirAll(pacDir, os.ModePerm); err != nil {
			panic(errors.New(fmt.Sprintf("failed to create log_pac_dir %s: %s", pacDir, err)))
		}
		if err := os.MkdirAll(fatDir, os.ModePerm); err != nil {
			panic(errors.New(fmt.Sprintf("failed to create log_fat_dir %s: %s", fatDir, err)))
		}
		infoWriteSyncer := &dateWriteSyncer{outPath: infoDir}
		warnWriteSyncer := &dateWriteSyncer{outPath: warnDir}
		errWriteSyncer := &dateWriteSyncer{outPath: errDir}
		pacWriteSyncer := &dateWriteSyncer{outPath: pacDir}
		fatWriteSyncer := &dateWriteSyncer{outPath: fatDir}

		// core
		core := zapcore.NewTee(
			// info
			zapcore.NewCore(encoder, zapcore.AddSync(infoWriteSyncer), zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
				return lv < zapcore.WarnLevel
			})),
			// warn
			zapcore.NewCore(encoder, zapcore.AddSync(warnWriteSyncer), zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
				return (lv >= zapcore.WarnLevel) && (lv < zapcore.ErrorLevel)
			})),
			// error
			zapcore.NewCore(encoder, zapcore.AddSync(errWriteSyncer), zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
				return (lv >= zapcore.ErrorLevel) && (lv < zapcore.DPanicLevel)
			})),
			// panic
			zapcore.NewCore(encoder, zapcore.AddSync(pacWriteSyncer), zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
				return (lv >= zapcore.DPanicLevel) && (lv < zapcore.FatalLevel)
			})),
			// fatal
			zapcore.NewCore(encoder, zapcore.AddSync(fatWriteSyncer), zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
				return lv >= zapcore.FatalLevel
			})),
		)

		// logger
		logger = zap.New(core)
	} else {
		// production config
		cfg := zap.NewProductionConfig()
		cfg.Development = false
		cfg.Encoding = "console"
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		cfg.Sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
		cfg.DisableCaller = true
		cfg.EncoderConfig = encodeCfg

		// logger
		var err error
		logger, err = cfg.Build()
		if err != nil {
			panic(err)
		}
	}
}

// dateWriteSyncer 按日期写入日志
type dateWriteSyncer struct {
	file    *os.File
	outPath string
}

func (d *dateWriteSyncer) Write(p []byte) (n int, err error) {
	fileName := filepath.Join(d.outPath, time.Now().Format(fileNameFormat)+".log")
	if d.file == nil || d.file.Name() != fileName {
		if d.file != nil {
			_ = d.file.Close()
		}
		d.file, err = os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return 0, err
		}
	}
	return d.file.Write(p)
}

func (d *dateWriteSyncer) Sync() error {
	if d.file != nil {
		return d.file.Sync()
	}
	return nil
}

// LogExit 退出日志
func LogExit() {
	if logger == nil {
		return
	}
	_ = logger.Sync()
}

func Debug(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Debug(msg)
		return
	}
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Info(msg)
		return
	}
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Warn(msg)
		return
	}
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Error(msg)
		return
	}
	logger.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Error(msg)
		return
	}
	logger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Error(msg)
		return
	}
	logger.Fatal(msg, fields...)
}
