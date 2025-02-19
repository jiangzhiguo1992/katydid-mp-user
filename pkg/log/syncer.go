package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DateWriteSyncer 按日期写入日志
type DateWriteSyncer struct {
	sync.Mutex
	file      *os.File
	outPath   string        // 输出目录路径
	format    string        // 文件名格式
	checkInt  time.Duration // 定期清理时间间隔
	maxAge    time.Duration // 最大保存时间
	maxSize   int64         // 最大文件大小
	lastClean time.Time     // 上次清理时间
}

// NewDateWriteSyncer 创建日志写入器
func NewDateWriteSyncer(
	outPath string,
	format string,
	checkInt time.Duration,
	maxAge time.Duration,
	maxSize int64,
) *DateWriteSyncer {
	return &DateWriteSyncer{
		outPath:   outPath,
		format:    format,
		checkInt:  checkInt,
		maxAge:    maxAge,
		maxSize:   maxSize,
		lastClean: time.Now(),
	}
}

func (d *DateWriteSyncer) Write(p []byte) (int, error) {
	d.Lock()
	defer d.Unlock()

	// 定期清理，避免频繁检查
	if time.Since(d.lastClean) > d.checkInt {
		d.triggerCleanup()
	}

	fileName := d.getCurrentFileName()
	if err := d.rotateIfNeeded(fileName); err != nil {
		return 0, fmt.Errorf("rotate log file failed: %w", err)
	}

	return d.file.Write(p)
}

func (d *DateWriteSyncer) triggerCleanup() {
	d.lastClean = time.Now()
	go func() {
		if err := d.cleanOldLogs(); err != nil {
			// 使用标准库log记录清理错误，避免循环依赖
			fmt.Printf("clean old logs failed: %v\n", err)
		}
	}()
}

func (d *DateWriteSyncer) cleanOldLogs() error {
	if d.maxAge <= 0 {
		return nil
	}

	cutoff := time.Now().Add(-d.maxAge)
	return filepath.Walk(d.outPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if info.ModTime().Before(cutoff) {
			return os.Remove(path)
		}
		return nil
	})
}

func (d *DateWriteSyncer) rotateIfNeeded(fileName string) error {
	if d.file != nil {
		if d.file.Name() == fileName {
			if d.maxSize > 0 {
				if stat, err := d.file.Stat(); err == nil {
					if stat.Size() >= d.maxSize {
						return d.rotate(fileName)
					}
				}
			}
			return nil
		}
		_ = d.file.Close()
	}

	return d.createNewFile(fileName)
}

// getCurrentFileName 获取当前日志文件名
func (d *DateWriteSyncer) getCurrentFileName() string {
	return filepath.Join(d.outPath, time.Now().Format(d.format)+".log")
}

// rotate 轮转日志文件
func (d *DateWriteSyncer) rotate(fileName string) error {
	// 关闭当前文件
	if d.file != nil {
		if err := d.file.Close(); err != nil {
			return err
		}
	}

	// 重命名当前文件，加上时间戳后缀
	if err := os.Rename(fileName, fileName+"."+time.Now().Format("150405")); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 创建新文件
	return d.createNewFile(fileName)
}

// createNewFile 创建新的日志文件
func (d *DateWriteSyncer) createNewFile(fileName string) error {
	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	d.file = f
	return nil
}

// Sync 实现 zapcore.WriteSyncer 接口
func (d *DateWriteSyncer) Sync() error {
	d.Lock()
	defer d.Unlock()

	if d.file != nil {
		return d.file.Sync()
	}
	return nil
}