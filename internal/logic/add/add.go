package add

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/defaults"
	"github.com/cicbyte/gig/internal/gitignore"
	"github.com/cicbyte/gig/internal/models"
	"github.com/cicbyte/gig/internal/utils"
)

type AddConfig struct {
	Languages []string
	Source    string
	Update    bool
	Yes       bool
	TUI       bool
	Save      bool
}

type AddResult struct {
	OriginalContent string
	FinalContent    string
	NewRulesCount   int
	DiffText        string
}

type AddProcessor struct {
	config    *AddConfig
	appConfig *models.Config
}

func NewAddProcessor(config *AddConfig, appConfig *models.Config) *AddProcessor {
	return &AddProcessor{config: config, appConfig: appConfig}
}

func (p *AddProcessor) Execute(ctx context.Context) (*AddResult, error) {
	var newContent string
	var err error

	switch p.config.Source {
	case "ai":
		result, err := utils.AI.StreamChat("add", strings.Join(p.config.Languages, ", "))
		if err != nil {
			return nil, fmt.Errorf("AI 生成失败：%w", err)
		}
		newContent = defaults.CleanupAIOutput(result)
	case "github":
		newContent, err = p.getTemplatesContent("github")
		if err != nil {
			return nil, err
		}
	case "api":
		apiClient := gitignore.NewAPIClient()
		newContent, err = apiClient.GetTemplate(p.config.Languages)
		if err != nil {
			return nil, fmt.Errorf("从 API 获取 .gitignore 失败：%w", err)
		}
	default:
		newContent, err = p.getTemplatesContent("local")
		if err != nil {
			return nil, err
		}
	}

	// --save: 将生成的内容保存为本地模板
	if p.config.Save {
		if err := p.saveTemplates(newContent); err != nil {
			fmt.Fprintf(os.Stderr, "保存模板失败：%v\n", err)
		}
	}

	return p.prepareMerge(newContent)
}

func (p *AddProcessor) getTemplatesContent(source string) (string, error) {
	if source == "github" {
		// 检查 GitHub 模板是否已克隆，未克隆则自动克隆
		home, _ := os.UserHomeDir()
		githubDir := filepath.Join(home, ".cicbyte", "gig", "template_github")
		if _, err := os.Stat(githubDir); os.IsNotExist(err) {
			if _, err := common.EnsureRemoteTemplatesAreCloned(); err != nil {
				return "", fmt.Errorf("GitHub 模板同步失败：%w", err)
			}
		}
	}

	templateManager, err := gitignore.NewTemplateManager()
	if err != nil {
		return "", fmt.Errorf("创建模板管理器时出错：%w", err)
	}

	var contentBuilder strings.Builder
	for _, lang := range p.config.Languages {
		templateContent, err := templateManager.GetTemplateContent(lang, source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告：找不到 '%s' 的模板。正在跳过。\n", lang)
			continue
		}
		contentBuilder.WriteString(templateContent)
		contentBuilder.WriteString("\n")
	}

	return contentBuilder.String(), nil
}

func (p *AddProcessor) prepareMerge(newContent string) (*AddResult, error) {
	gitignorePath, err := common.GetGitignorePath()
	if err != nil {
		return nil, err
	}

	updater := gitignore.NewGitignoreUpdater(gitignorePath)
	originalContent, err := updater.ReadContent()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("读取现有的 .gitignore 文件时出错：%w", err)
	}

	finalContent, newRulesCount, err := updater.MergeContent(newContent)
	if err != nil {
		return nil, fmt.Errorf("合并内容时出错：%w", err)
	}

	diffText := ""
	if newRulesCount > 0 {
		diffText = updater.ShowDiff(originalContent, finalContent)
	}

	return &AddResult{
		OriginalContent: originalContent,
		FinalContent:    finalContent,
		NewRulesCount:   newRulesCount,
		DiffText:        diffText,
	}, nil
}

func (p *AddProcessor) Apply(finalContent string) error {
	gitignorePath, err := common.GetGitignorePath()
	if err != nil {
		return err
	}

	updater := gitignore.NewGitignoreUpdater(gitignorePath)
	return updater.ApplyChanges(finalContent)
}

func (p *AddProcessor) saveTemplates(content string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取主目录失败：%w", err)
	}
	templateDir := filepath.Join(home, ".cicbyte", "gig", "template")

	for _, lang := range p.config.Languages {
		templatePath := filepath.Join(templateDir, lang+".gitignore")
		if utils.FileExists(templatePath) {
			fmt.Printf("模板 '%s' 已存在，跳过保存。\n", lang)
			continue
		}
		if err := utils.WriteFile(templatePath, content); err != nil {
			return fmt.Errorf("保存模板 '%s' 失败：%w", lang, err)
		}
		fmt.Printf("[OK] 已保存模板 '%s' 于 %s\n", lang, templatePath)
	}
	return nil
}
