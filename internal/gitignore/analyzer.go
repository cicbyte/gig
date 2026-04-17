package gitignore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/gig/internal/utils"
)

// RuleAnalyzer 提供规则分析功能
type RuleAnalyzer struct {
	GlobalRules map[string]bool
}

// NewRuleAnalyzer 创建规则分析器
func NewRuleAnalyzer() *RuleAnalyzer {
	return &RuleAnalyzer{
		GlobalRules: getGlobalGitignoreRules(),
	}
}

// AnalyzeContent 分析gitignore内容并返回所有问题
func (a *RuleAnalyzer) AnalyzeContent(content string) AuditResult {
	logger := GetLogger()
	logger.Debugf("开始分析gitignore内容，长度: %d字节", len(content))

	rules := ParseRules(content)
	logger.Debugf("解析出 %d 条规则", len(rules))

	result := AuditResult{}

	// 检查重复规则
	logger.Debug("检查重复规则")
	duplicateIssues := a.CheckDuplicates(rules)
	for _, issue := range duplicateIssues {
		result.Redundancy = append(result.Redundancy, issue)
	}

	// 检查危险模式
	logger.Debug("检查危险模式")
	dangerIssues := a.CheckDangerousPatterns(rules)
	for _, issue := range dangerIssues {
		result.Danger = append(result.Danger, issue)
	}

	// 检查性能优化
	logger.Debug("检查性能优化")
	perfIssues := a.CheckPerformance(rules)
	for _, issue := range perfIssues {
		result.Performance = append(result.Performance, issue)
	}

	// 检查全局规则冲突
	logger.Debug("检查全局规则冲突")
	portIssues := a.CheckGlobalRules(rules)
	for _, issue := range portIssues {
		result.Portability = append(result.Portability, issue)
	}

	// 检查否定模式有效性
	logger.Debug("检查否定模式有效性")
	negationIssues := a.CheckNegationPatterns(rules)
	for _, issue := range negationIssues {
		result.Danger = append(result.Danger, issue)
	}

	totalIssues := len(result.Redundancy) + len(result.Danger) +
		len(result.Portability) + len(result.Performance)
	logger.Infof("分析完成，共发现 %d 个问题", totalIssues)

	return result
}

// CheckDuplicates 检查重复规则
func (a *RuleAnalyzer) CheckDuplicates(rules []Rule) []Issue {
	var issues []Issue
	seenRules := make(map[string]int)

	for _, rule := range rules {
		if line, ok := seenRules[rule.Pattern]; ok {
			issues = append(issues, Issue{
				Category: "redundancy",
				Message:  fmt.Sprintf("规则 '%s' 在第 %d 行，与第 %d 行重复", rule.Pattern, rule.LineNum, line),
				LineNum:  rule.LineNum,
				Pattern:  rule.Pattern,
			})
		}
		seenRules[rule.Pattern] = rule.LineNum
	}

	return issues
}

// CheckDangerousPatterns 检查危险模式
func (a *RuleAnalyzer) CheckDangerousPatterns(rules []Rule) []Issue {
	var issues []Issue

	for _, rule := range rules {
		p := rule.Pattern
		if p == "*" || p == "!/" {
			issues = append(issues, Issue{
				Category: "danger",
				Message:  fmt.Sprintf("模式 '%s' 在第 %d 行过于宽泛", p, rule.LineNum),
				LineNum:  rule.LineNum,
				Pattern:  p,
			})
		}
		if p == ".gitignore" {
			issues = append(issues, Issue{
				Category: "danger",
				Message:  fmt.Sprintf("不应忽略 .gitignore 本身（第 %d 行），这会导致 Git 无法读取规则", rule.LineNum),
				LineNum:  rule.LineNum,
				Pattern:  p,
			})
		}
		if p == ".git" {
			issues = append(issues, Issue{
				Category: "danger",
				Message:  fmt.Sprintf("无需忽略 .git 目录（第 %d 行），Git 已自动处理", rule.LineNum),
				LineNum:  rule.LineNum,
				Pattern:  p,
			})
		}
	}

	return issues
}

// CheckPerformance 检查性能优化
func (a *RuleAnalyzer) CheckPerformance(rules []Rule) []Issue {
	var issues []Issue

	for _, rule := range rules {
		if strings.HasSuffix(rule.Pattern, "/*") && len(strings.Split(rule.Pattern, "/")) == 2 {
			issues = append(issues, Issue{
				Category: "performance",
				Message:  fmt.Sprintf("模式 '%s' 在第 %d 行可以优化为 '%s'", rule.Pattern, rule.LineNum, strings.TrimSuffix(rule.Pattern, "/*")),
				LineNum:  rule.LineNum,
				Pattern:  rule.Pattern,
			})
		}
	}

	return issues
}

// CheckGlobalRules 检查全局规则冲突
func (a *RuleAnalyzer) CheckGlobalRules(rules []Rule) []Issue {
	var issues []Issue

	for _, rule := range rules {
		if a.GlobalRules[rule.Pattern] {
			issues = append(issues, Issue{
				Category: "portability",
				Message:  fmt.Sprintf("规则 '%s' 在第 %d 行已被全局 .gitignore 覆盖", rule.Pattern, rule.LineNum),
				LineNum:  rule.LineNum,
				Pattern:  rule.Pattern,
			})
		}
	}

	return issues
}

// CheckNegationPatterns 检查否定模式（!规则）是否因父目录被忽略而失效
func (a *RuleAnalyzer) CheckNegationPatterns(rules []Rule) []Issue {
	var issues []Issue

	// 收集所有非否定的忽略规则
	ignoredDirs := make(map[string]int)
	for _, rule := range rules {
		pattern := rule.Pattern
		if strings.HasPrefix(pattern, "!") {
			continue
		}
		// 记录以 / 结尾的目录忽略规则
		if strings.HasSuffix(pattern, "/") {
			ignoredDirs[strings.TrimSuffix(pattern, "/")] = rule.LineNum
		}
	}

	// 检查每个否定规则
	for _, rule := range rules {
		if !strings.HasPrefix(rule.Pattern, "!") {
			continue
		}
		negationPath := strings.TrimPrefix(rule.Pattern, "!")

		// 检查否定路径的每一级父目录是否被忽略
		parts := strings.Split(strings.Trim(negationPath, "/"), "/")
		for i := 1; i < len(parts); i++ {
			parent := strings.Join(parts[:i], "/")
			if line, ok := ignoredDirs[parent]; ok {
				issues = append(issues, Issue{
					Category: "danger",
					Message:  fmt.Sprintf("否定模式 '%s' 在第 %d 行可能无效：父目录 '%s' 在第 %d 行已被忽略", rule.Pattern, rule.LineNum, parent, line),
					LineNum:  rule.LineNum,
					Pattern:  rule.Pattern,
				})
				break
			}
		}
	}

	return issues
}

// getGlobalGitignoreRules 获取全局gitignore规则
func getGlobalGitignoreRules() map[string]bool {
	globalRules := make(map[string]bool)

	// 尝试读取全局gitignore文件
	home, err := utils.GetUserHomeDir()
	if err != nil {
		return globalRules
	}

	// 检查常见的全局gitignore位置
	globalPaths := []string{
		filepath.Join(home, ".gitignore_global"),
		filepath.Join(home, ".gitignore"),
	}

	for _, path := range globalPaths {
		if _, err := os.Stat(path); err == nil {
			content, err := utils.ReadFile(path)
			if err != nil {
				continue
			}

			scanner := bufio.NewScanner(strings.NewReader(string(content)))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && !strings.HasPrefix(line, "#") {
					globalRules[line] = true
				}
			}

			break
		}
	}

	return globalRules
}
