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

var logger *zap.Logger

// Init 初始化日志
func Init(prod bool, logDir string, outLevel *int, outFormat *string) {
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
		var levels []string
		var enables []func(lv zapcore.Level) bool
		if outLevel != nil {
			if *outLevel >= 0 {
				levels = append(levels, "fat")
				enables = append(enables, func(lv zapcore.Level) bool {
					return lv >= zapcore.FatalLevel
				})
			}
			if *outLevel >= 1 {
				levels = append(levels, "pac")
				enables = append(enables, func(lv zapcore.Level) bool {
					return (lv >= zapcore.DPanicLevel) && (lv < zapcore.FatalLevel)
				})
			}
			if *outLevel >= 2 {
				levels = append(levels, "err")
				enables = append(enables, func(lv zapcore.Level) bool {
					return (lv >= zapcore.ErrorLevel) && (lv < zapcore.DPanicLevel)
				})
			}
			if *outLevel >= 3 {
				levels = append(levels, "warn")
				enables = append(enables, func(lv zapcore.Level) bool {
					return (lv >= zapcore.WarnLevel) && (lv < zapcore.ErrorLevel)
				})
			}
			if *outLevel >= 4 {
				levels = append(levels, "info")
				enables = append(enables, func(lv zapcore.Level) bool {
					return (lv >= zapcore.InfoLevel) && (lv < zapcore.WarnLevel)
				})
			}
			if *outLevel >= 5 {
				levels = append(levels, "debug")
				enables = append(enables, func(lv zapcore.Level) bool {
					return lv < zapcore.InfoLevel
				})
			}
		}

		// cores
		var cores []zapcore.Core
		for i, v := range levels {
			dir := path.Join(logDir, v)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				panic(errors.New(fmt.Sprintf("failed to create dir %s: %s", v, err)))
			}
			writer := &dateWriteSyncer{outPath: dir, format: outFormat}
			core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
				return enables[i](lv)
			}))
			cores = append(cores, core)
		}

		// logger
		core := zapcore.NewTee(cores...)
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
	format  *string
	outPath string
}

func (d *dateWriteSyncer) Write(p []byte) (n int, err error) {
	format := "06-01-02"
	if (d.format != nil) && (len(*d.format) > 0) {
		format = *d.format
	}
	fileName := filepath.Join(d.outPath, time.Now().Format(format)+".log")
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

// OnExit 退出日志
func OnExit() {
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
