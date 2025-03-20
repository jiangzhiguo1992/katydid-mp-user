package file

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Exists 检查文件或目录是否存在
func Exists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

// IsDir 检查路径是否为目录
func IsDir(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile 检查路径是否为文件
func IsFile(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// CreateDir 创建目录，如果不存在
func CreateDir(path string) error {
	if path == "" {
		return errors.New("path is empty")
	}
	return os.MkdirAll(path, os.ModePerm)
}

// CreateFile 创建文件
func CreateFile(path string) (*os.File, error) {
	err := CreateDir(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	return file, err
}

// DeleteFile 删除文件
func DeleteFile(path string) error {
	if !Exists(path) {
		return errors.New("file does not exist")
	}
	return os.Remove(path)
}

// DeleteDir 删除目录及其下所有文件
func DeleteDir(path string) error {
	if !Exists(path) {
		return errors.New("file does not exist")
	}
	return os.RemoveAll(path)
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

// GetFilesBySuffix 获取指定目录下的特定后缀文件
func GetFilesBySuffix(dir string, suffix string) ([]string, error) {
	if !Exists(dir) {
		return nil, errors.New("file does not exist")
	}

	if !IsDir(dir) {
		return nil, errors.New("path is not a directory")
	}

	var result []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), suffix) {
			result = append(result, path)
		}
		return nil
	})

	return result, err
}

// GetSize 获取文件大小
func GetSize(path string) (int64, error) {
	if !Exists(path) {
		return 0, errors.New("file does not exist")
	}

	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatSize 文件大小格式化
func FormatSize(size int64) string {
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

// ReadLines 按行读取文件内容
func ReadLines(path string) ([]string, error) {
	if !Exists(path) {
		return nil, errors.New("file does not exist")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// WriteFile 写入内容到文件
func WriteFile(path string, data []byte, append bool) error {
	dir := filepath.Dir(path)
	if !Exists(dir) {
		if err := CreateDir(dir); err != nil {
			return err
		}
	}

	flag := os.O_CREATE | os.O_WRONLY
	if append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	file, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	if !Exists(src) {
		return errors.New("file does not exist")
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dir := filepath.Dir(dst)
	if !Exists(dir) {
		if err := CreateDir(dir); err != nil {
			return err
		}
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
