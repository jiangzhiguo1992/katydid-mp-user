package tools

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DirCreate 创建目录
func DirCreate(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return errors.New(fmt.Sprintf("failed to create dir %s: %s", path, err))
	}
	return nil
}

// FileCreate 创建文件
func FileCreate(path string) (*os.File, error) {
	// dir
	err := DirCreate(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	// file
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to open file %s: %s", path, err))
	}
	return file, nil
}

// GetAllFiles 获取目录下的所有文件
func GetAllFiles(dir string, suffix *string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if suffix == nil {
				files = append(files, path)
			} else if (suffix != nil) && (filepath.Ext(info.Name()) == *suffix) {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// FileSizeFormat 文件大小格式化
func FileSizeFormat(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
