package utils

import (
	"os"
	"path/filepath"
)

// 判断文件或目录是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// 读取文件内容为字符串
func ReadFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// 写入字符串到文件
func WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0660)
}

// 获取用户主目录
func GetUserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// 遍历目录，带自定义处理函数
func WalkDir(root string, fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return fn(path, info)
	})
}
