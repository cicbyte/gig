package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetTemplateContent reads the content of a template file from a given directory.
// Supports case-insensitive matching (e.g. "python" matches "Python.gitignore").
func GetTemplateContent(templateName, templateDir string) (string, error) {
	// 路径穿越校验
	if strings.ContainsAny(templateName, `/\`) || strings.Contains(templateName, "..") {
		return "", fmt.Errorf("非法模板名称: %s", templateName)
	}

	// 精确匹配
	templatePath := filepath.Join(templateDir, templateName+".gitignore")
	content, err := os.ReadFile(templatePath)
	if err == nil {
		return string(content), nil
	}

	// 大小写不敏感匹配
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return "", fmt.Errorf("模板目录不存在或无法读取")
	}
	lowerName := strings.ToLower(templateName)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".gitignore") {
			if strings.ToLower(strings.TrimSuffix(e.Name(), ".gitignore")) == lowerName {
				content, err := os.ReadFile(filepath.Join(templateDir, e.Name()))
				if err != nil {
					return "", err
				}
				return string(content), nil
			}
		}
	}

	return "", fmt.Errorf("找不到模板 '%s'", templateName)
}
