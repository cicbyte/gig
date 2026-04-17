package doctor

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/gitignore"
	"github.com/cicbyte/gig/internal/models"
	"github.com/cicbyte/gig/internal/utils"
)

type AuditConfig struct {
	Source string // "local", "github", "ai", "api"
}

type AuditResultData struct {
	ProjectTypes []string
	Issues       []gitignore.AuditIssue
	AIReport     string // AI 模式下的原始文本
	Content      string // .gitignore 原始内容（供 --fix 使用）
	Streamed     bool   // AI 模式是否已流式输出
}

type AuditProcessor struct {
	config    *AuditConfig
	appConfig *models.Config
}

func PeekProjectTypes(p *AuditProcessor) []string {
	return utils.DetectionUtil.DetectProjectTypes(p.appConfig.Detection.FileMap)
}

func NewAuditProcessor(config *AuditConfig, appConfig *models.Config) *AuditProcessor {
	return &AuditProcessor{config: config, appConfig: appConfig}
}

func (p *AuditProcessor) Execute(ctx context.Context) (*AuditResultData, error) {
	_, err := gitignore.InitLogger(gitignore.LogError)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告：初始化日志记录器失败：%v\n", err)
	}

	gitignorePath, err := common.GetGitignorePath()
	if err != nil {
		return nil, err
	}

	updater := gitignore.NewGitignoreUpdater(gitignorePath)
	content, err := updater.ReadContent()
	if err != nil {
		return nil, fmt.Errorf("读取 .gitignore 文件时出错：%w", err)
	}

	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("当前目录中未找到 .gitignore 文件或文件为空，请先使用 'gig add' 创建")
	}

	projectTypes := utils.DetectionUtil.DetectProjectTypes(p.appConfig.Detection.FileMap)

	// 规则引擎：始终运行
	analyzer := gitignore.NewRuleAnalyzer()
	auditResult := analyzer.AnalyzeContent(content)
	issues := convertIssues(auditResult)

	// 模板建议（非 AI 模式）
	if p.config.Source != "ai" {
		var missingIssues []gitignore.AuditIssue
		switch p.config.Source {
		case "github":
			tm, err := gitignore.NewTemplateManager()
			if err != nil {
				return nil, fmt.Errorf("创建模板管理器时出错：%w", err)
			}
			missingIssues = tm.GetSuggestionsFromTemplates(content, projectTypes, "github")
		case "api":
			apiClient := gitignore.NewAPIClient()
			suggestions, err := apiClient.GetSuggestionsFromAPI(content, projectTypes)
			if err != nil {
				fmt.Println("从 API 获取建议失败：", err)
			}
			for _, s := range suggestions {
				missingIssues = append(missingIssues, gitignore.AuditIssue{
					Severity: "info",
					Category: "missing",
					Message:  s,
					Fixable:  true,
				})
			}
		default:
			tm, err := gitignore.NewTemplateManager()
			if err != nil {
				return nil, fmt.Errorf("创建模板管理器时出错：%w", err)
			}
			missingIssues = tm.GetSuggestionsFromTemplates(content, projectTypes, "local")
		}
		issues = append(issues, missingIssues...)
	}

	result := &AuditResultData{
		ProjectTypes: projectTypes,
		Issues:       issues,
		Content:      content,
	}

	// AI 模式：额外获取 AI 报告
	if p.config.Source == "ai" {
		osInfo := runtime.GOOS + "/" + runtime.GOARCH
		templateRules := p.collectTemplateRules(projectTypes)
		dirTree := scanProjectTree(".", 3)
		aiReport, err := utils.AI.StreamChat("audit", strings.Join(projectTypes, ", "), osInfo, content, templateRules, dirTree)
		if err != nil {
			return nil, fmt.Errorf("AI 审计失败：%w", err)
		}
		result.AIReport = strings.TrimSpace(aiReport)
		result.Streamed = true
	}

	return result, nil
}

// Fix 自动修复可修复的问题，返回修复后的内容
func (p *AuditProcessor) Fix(ctx context.Context, result *AuditResultData) (string, int, error) {
	lines := strings.Split(result.Content, "\n")
	fixedCount := 0

	// 1. 删除重复规则（保留第一次出现）
	seen := make(map[string]bool)
	var kept []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			kept = append(kept, line)
			continue
		}
		if seen[trimmed] {
			fixedCount++
			continue
		}
		seen[trimmed] = true
		kept = append(kept, line)
	}

	// 2. 修复性能问题（dir/* → dir）
	for i, line := range kept {
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, "/*") && len(strings.Split(trimmed, "/")) == 2 {
			kept[i] = strings.TrimSuffix(trimmed, "/*") + "/"
			fixedCount++
		}
	}

	// 3. 追加缺失规则
	var missingRules []string
	for _, issue := range result.Issues {
		if issue.Category == "missing" && issue.Fixable && issue.Pattern != "" {
			if !seen[issue.Pattern] {
				missingRules = append(missingRules, issue.Pattern)
			}
		}
	}
	if len(missingRules) > 0 {
		kept = append(kept, "")
		kept = append(kept, "# --- 由 gig doctor --fix 自动添加 ---")
		kept = append(kept, missingRules...)
		fixedCount += len(missingRules)
	}

	return strings.Join(kept, "\n"), fixedCount, nil
}

// convertIssues 将内部 Issue 转为统一的 AuditIssue
func convertIssues(result gitignore.AuditResult) []gitignore.AuditIssue {
	var issues []gitignore.AuditIssue

	for _, issue := range result.Danger {
		issues = append(issues, gitignore.AuditIssue{
			Severity: "error",
			Category: "danger",
			Message:  issue.Message,
			LineNum:  issue.LineNum,
			Pattern:  issue.Pattern,
			Fixable:  false,
		})
	}
	for _, issue := range result.Redundancy {
		issues = append(issues, gitignore.AuditIssue{
			Severity: "warning",
			Category: "redundancy",
			Message:  issue.Message,
			LineNum:  issue.LineNum,
			Pattern:  issue.Pattern,
			Fixable:  true,
		})
	}
	for _, issue := range result.Portability {
		issues = append(issues, gitignore.AuditIssue{
			Severity: "info",
			Category: "portability",
			Message:  issue.Message,
			LineNum:  issue.LineNum,
			Pattern:  issue.Pattern,
			Fixable:  false,
		})
	}
	for _, issue := range result.Performance {
		issues = append(issues, gitignore.AuditIssue{
			Severity: "info",
			Category: "performance",
			Message:  issue.Message,
			LineNum:  issue.LineNum,
			Pattern:  issue.Pattern,
			Fixable:  true,
		})
	}
	return issues
}

// collectTemplateRules 收集项目类型对应的模板规则
func (p *AuditProcessor) collectTemplateRules(projectTypes []string) string {
	templateManager, err := gitignore.NewTemplateManager()
	if err != nil {
		return ""
	}

	var builder strings.Builder
	for _, pt := range projectTypes {
		tmplContent, err := templateManager.GetTemplateContent(pt, "local")
		if err != nil {
			continue
		}
		rules := gitignore.ExtractRules(tmplContent)
		builder.WriteString(fmt.Sprintf("[%s]\n", pt))
		for rule := range rules {
			builder.WriteString(rule + "\n")
		}
	}
	return builder.String()
}

// scanProjectTree 扫描项目目录结构
func scanProjectTree(root string, maxDepth int) string {
	skipDirs := map[string]bool{
		"node_modules": true, "vendor": true, "third_party": true,
		"__pycache__": true, "venv": true, ".venv": true,
		"pkg": true, "target": true,
		".gradle": true, "gradle": true,
		".next": true, ".nuxt": true, ".cache": true, "dist": true,
		".vscode": true, ".idea": true, ".vs": true,
		"coverage": true, ".coverage": true, ".nyc_output": true,
		".git": true, ".svn": true, ".hg": true,
		".serverless": true, ".terraform": true,
	}

	skipExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".wasm": true, ".a": true, ".o": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".svg": true, ".ico": true, ".webp": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true,
		".pdf": true, ".doc": true, ".docx": true,
		".lock": true,
	}

	maxEntries := 80
	var entries []string
	truncated := false

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if truncated {
			return fs.SkipAll
		}

		if path == "." || path == root {
			return nil
		}

		rel := strings.TrimPrefix(path, root)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		depth := strings.Count(rel, string(filepath.Separator))

		if depth > maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if skipDirs[d.Name()] {
				return fs.SkipDir
			}
			entries = append(entries, "├── "+rel+"/")
			if len(entries) >= maxEntries {
				truncated = true
			}
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") && d.Name() != ".gitignore" && d.Name() != ".env" && d.Name() != ".env.example" {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if skipExts[ext] {
			return nil
		}

		entries = append(entries, "├── "+rel)
		if len(entries) >= maxEntries {
			truncated = true
		}
		return nil
	})

	sort.Strings(entries)
	result := strings.Join(entries, "\n")
	if truncated {
		result += fmt.Sprintf("\n... (已截断，仅显示前 %d 项)", maxEntries)
	}
	return result
}
