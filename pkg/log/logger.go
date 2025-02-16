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
	prod   bool
	logger *zap.Logger
)

// Init 初始化日志
func Init(p bool, logDir string, outLevel *int, outFormat *string) {
	prod = p
	// encoder
	encodeCfg := zap.NewProductionEncoderConfig()
	encodeCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	if prod {
		encodeCfg.LevelKey = ""
	} else {
		encodeCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("\x1b[20m%s\x1b[0m", t.Format("2006-01-02 15:04:05.000")))
		}
		//encodeCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encodeCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			switch l {
			default:
			case zapcore.DebugLevel:
				enc.AppendString("\x1b[47mDEBUG\x1b[0m") // 灰色背景
			case zapcore.InfoLevel:
				enc.AppendString("\x1b[42mINFO\x1b[0m") // 绿色背景
			case zapcore.WarnLevel:
				enc.AppendString("\x1b[43mWARN\x1b[0m") // 黄色背景
			case zapcore.ErrorLevel:
				enc.AppendString("\x1b[41mERROR\x1b[0m") // 红色背景
			case zapcore.DPanicLevel:
				enc.AppendString("\x1b[45mDPANIC\x1b[0m") // 紫色背景
			case zapcore.PanicLevel:
				enc.AppendString("\x1b[45mPANIC\x1b[0m") // 紫色背景
			case zapcore.FatalLevel:
				enc.AppendString("\x1b[40mFATAL\x1b[0m") // 黑色背景
			}
		}
	}
	var encoder zapcore.Encoder
	if prod {
		encoder = zapcore.NewJSONEncoder(encodeCfg)
	} else {
		//encoder = &customConsoleEncoder{zapcore.NewConsoleEncoder(encodeCfg)}
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

//// customConsoleEncoder 自定义的控制台编码器
//type customConsoleEncoder struct {
//	zapcore.Encoder
//}
//
//func (c *customConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
//	fmt.Println("EncodeEntry called")
//	buf, err := c.Encoder.EncodeEntry(entry, fields)
//	if err != nil {
//		return nil, err
//	}
//	// 将消息内容变为绿色
//	buf.AppendString(fmt.Sprintf("\x1b[32m%s\x1b[0m", entry.Message))
//	return buf, nil
//}
//
//func (c *customConsoleEncoder) AddByteString(key string, value []byte) {
//	logger.Warn("AddByteString is not implemented")
//}
//
//func (c *customConsoleEncoder) AddString(key, value string) {
//	logger.Warn("AddByteString is not implemented")
//}

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
	if prod {
		logger.Debug(msg, fields...)
	} else {
		logger.Debug(fmt.Sprintf("\x1b[37m%s\x1b[0m", msg), fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Info(msg)
		return
	}
	if prod {
		logger.Info(msg, fields...)
	} else {
		logger.Info(fmt.Sprintf("\x1b[32m%s\x1b[0m", msg), fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Warn(msg)
		return
	}
	if prod {
		logger.Warn(msg, fields...)
	} else {
		logger.Warn(fmt.Sprintf("\x1b[33m%s\x1b[0m", msg), fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Error(msg)
		return
	}
	if prod {
		logger.Error(msg, fields...)
	} else {
		logger.Error(fmt.Sprintf("\x1b[31m%s\x1b[0m", msg), fields...)
	}
}

func Panic(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Error(msg)
		return
	}
	if prod {
		logger.Panic(msg, fields...)
	} else {
		logger.Panic(fmt.Sprintf("\x1b[35m%s\x1b[0m", msg), fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if logger == nil {
		slog.Error(msg)
		return
	}
	if prod {
		logger.Fatal(msg, fields...)
	} else {
		logger.Panic(fmt.Sprintf("\x1b[30m%s\x1b[0m", msg), fields...)
	}
}
