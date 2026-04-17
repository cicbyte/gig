package defaults

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/gig/internal/models"
	"github.com/cicbyte/gig/prompts"
	"github.com/cicbyte/gig/template"
)

// loadBundledPrompts 从嵌入的 prompts FS 中读取所有 .md 文件
func loadBundledPrompts() map[string]string {
	result := make(map[string]string)
	files, err := prompts.PromptsFS.ReadDir(".")
	if err != nil {
		return result
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
			name := strings.TrimSuffix(f.Name(), ".md")
			data, err := prompts.PromptsFS.ReadFile(f.Name())
			if err != nil {
				continue
			}
			result[name] = string(data)
		}
	}
	return result
}

// GetDefaultConfig 返回应用的默认配置
func GetDefaultConfig() models.Config {
	return models.Config{
		AI: models.AIConfig{
			URL:   "https://api.deepseek.com/chat/completions",
			Model: "deepseek-chat",
		},
		Prompts: loadBundledPrompts(),
		Detection: models.DetectionConfig{
			FileMap: map[string]string{
				"go.mod":           "go",
				"package.json":     "node",
				"requirements.txt": "python",
				"pom.xml":          "java",
				"Cargo.toml":       "rust",
				"composer.json":    "php",
				"Gemfile":          "ruby",
				"Makefile":         "c/c++",
				"CMakeLists.txt":   "c/c++",
				"build.gradle":     "java",
				"Podfile":          "ios",
				"Package.swift":    "swift",
			},
		},
	}
}

// CleanupAIOutput 清理 AI 返回的 markdown 代码块包裹
func CleanupAIOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	start := -1
	end := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if start == -1 {
				start = i + 1
			} else {
				end = i
				break
			}
		}
	}

	if start != -1 && end != -1 {
		return strings.TrimSpace(strings.Join(lines[start:end], "\n"))
	}
	return strings.TrimSpace(raw)
}

// InitBundledTemplates 初始化内置模板到用户目录
func InitBundledTemplates(userTemplateDir string) {
	files, err := os.ReadDir(userTemplateDir)
	if err != nil || len(files) > 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "正在初始化模板于", userTemplateDir)
	fs.WalkDir(template.TemplatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".gitignore") {
			return copyEmbeddedTemplate(userTemplateDir, path)
		}
		return nil
	})
}

func copyEmbeddedTemplate(dstDir, srcPath string) error {
	dstPath := filepath.Join(dstDir, srcPath)

	data, err := template.TemplatesFS.ReadFile(srcPath)
	if err != nil {
		return err
	}

	return os.WriteFile(dstPath, data, 0644)
}
