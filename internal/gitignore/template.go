package gitignore

import (
	"fmt"
	"path/filepath"

	"github.com/cicbyte/gig/internal/utils"
)

// TemplateManager 管理gitignore模板
type TemplateManager struct {
	LocalDir  string
	RemoteDir string
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager() (*TemplateManager, error) {
	home, err := utils.GetUserHomeDir()
	if err != nil {
		return nil, err
	}

	return &TemplateManager{
		LocalDir:  filepath.Join(home, ".cicbyte", "gig", "template"),
		RemoteDir: filepath.Join(home, ".cicbyte", "gig", "template_github"),
	}, nil
}

// GetTemplateContent 获取指定语言的模板内容
func (tm *TemplateManager) GetTemplateContent(language string, source string) (string, error) {
	var templateDir string
	if source == "github" {
		templateDir = tm.RemoteDir
	} else {
		templateDir = tm.LocalDir
	}

	return utils.GetTemplateContent(language, templateDir)
}

// CompareWithTemplate 比较当前内容与模板，找出缺失的规则
func (tm *TemplateManager) CompareWithTemplate(currentContent string, language string, source string) []AuditIssue {
	templateContent, err := tm.GetTemplateContent(language, source)
	if err != nil {
		return nil
	}

	existingRules := ExtractRules(currentContent)
	templateRules := ParseRules(templateContent)

	var issues []AuditIssue
	for _, rule := range templateRules {
		if !existingRules[rule.Pattern] {
			issues = append(issues, AuditIssue{
				Severity: "info",
				Category: "missing",
				Message:  fmt.Sprintf("考虑添加 '%s' (来自 %s 模板)。", rule.Pattern, language),
				Pattern:  rule.Pattern,
				Fixable:  true,
			})
		}
	}

	return issues
}

// GetSuggestionsFromTemplates 从多个模板获取建议
func (tm *TemplateManager) GetSuggestionsFromTemplates(content string, projectTypes []string, source string) []AuditIssue {
	logger := GetLogger()
	logger.Debugf("从%s模板获取建议，项目类型: %v", source, projectTypes)

	var issues []AuditIssue

	for _, projectType := range projectTypes {
		logger.Debugf("处理项目类型: %s", projectType)
		templateIssues := tm.CompareWithTemplate(content, projectType, source)
		logger.Debugf("项目类型 %s 找到 %d 条建议", projectType, len(templateIssues))
		issues = append(issues, templateIssues...)
	}

	logger.Debugf("总共生成 %d 条建议", len(issues))
	return issues
}
