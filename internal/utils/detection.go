package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// 全局工具实例
var DetectionUtil = Detection{}

// Detection 工具结构体
type Detection struct{}

// 自动识别项目类型（递归遍历，仅根据配置文件匹配，特殊目录优化）
func (d Detection) DetectProjectTypes(fileMap map[string]string) []string {
	if len(fileMap) == 0 {
		return nil
	}

	langSet := make(map[string]struct{})

	// 特殊目录与类型的映射
	specialDirs := map[string]string{
		"node_modules": "node",
		".venv":        "python",
		"venv":         "python",
		"vendor":       "go",
		"target":       "rust",
	}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			// 跳过隐藏目录（如.git、.idea等）
			if strings.HasPrefix(base, ".") && path != "." {
				return filepath.SkipDir
			}
			// 特殊目录：遇到直接判定类型并跳过递归
			if lang, ok := specialDirs[base]; ok {
				langSet[lang] = struct{}{}
				return filepath.SkipDir
			}
			// 目录后缀匹配：.xcodeproj -> ios
			if strings.HasSuffix(base, ".xcodeproj") || strings.HasSuffix(base, ".xcworkspace") {
				langSet["ios"] = struct{}{}
			}
			return nil
		}

		name := info.Name()
		if lang, ok := fileMap[name]; ok {
			langSet[lang] = struct{}{}
		}
		return nil
	})

	var langs []string
	for lang := range langSet {
		langs = append(langs, lang)
	}
	return langs
}
