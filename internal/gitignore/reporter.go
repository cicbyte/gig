package gitignore

import (
	"fmt"
	"strings"
)

// Reporter 生成报告的接口
type Reporter interface {
	GenerateReport(result AuditResult, suggestions []string) string
}

// ConsoleReporter 控制台报告生成器
type ConsoleReporter struct{}

// GenerateReport 生成控制台格式的报告
func (r *ConsoleReporter) GenerateReport(result AuditResult, suggestions []string) string {
	var sb strings.Builder
	
	totalIssues := len(result.Danger) + len(result.Redundancy) + 
		len(result.Portability) + len(result.Performance)
	
	if totalIssues == 0 && len(suggestions) == 0 {
				sb.WriteString("[OK] 您的 .gitignore 文件看起来很棒！未发现任何问题或建议。\n")
		return sb.String()
	}
	
	// 显示危险问题
	if len(result.Danger) > 0 {
		sb.WriteString(fmt.Sprintf("[!DANGER] 严重问题 (%d)\n", len(result.Danger)))
		sb.WriteString("   这些模式可能会导致严重问题：\n")
		for _, issue := range result.Danger {
			sb.WriteString(fmt.Sprintf("   • %s\n", issue.Message))
		}
		sb.WriteString("\n")
	}
	
	// 显示冗余问题
	if len(result.Redundancy) > 0 {
		sb.WriteString(fmt.Sprintf("[!REDUNDANT] 冗余问题 (%d)\n", len(result.Redundancy)))
		sb.WriteString("   重复或不必要的规则：\n")
		for _, issue := range result.Redundancy {
			sb.WriteString(fmt.Sprintf("   • %s\n", issue.Message))
		}
		sb.WriteString("\n")
	}
	
	// 显示可移植性问题
	if len(result.Portability) > 0 {
		sb.WriteString(fmt.Sprintf("[!PORTABLE] 可移植性问题 (%d)\n", len(result.Portability)))
		sb.WriteString("   可能与全局设置冲突的规则：\n")
		for _, issue := range result.Portability {
			sb.WriteString(fmt.Sprintf("   • %s\n", issue.Message))
		}
		sb.WriteString("\n")
	}
	
	// 显示性能问题
	if len(result.Performance) > 0 {
		sb.WriteString(fmt.Sprintf("[!PERF] 性能优化 (%d)\n", len(result.Performance)))
		sb.WriteString("   可以更高效的模式：\n")
		for _, issue := range result.Performance {
			sb.WriteString(fmt.Sprintf("   • %s\n", issue.Message))
		}
		sb.WriteString("\n")
	}
	
	
	return sb.String()
}

// 辅助函数：按模板分组建议
func groupSuggestionsByTemplate(suggestions []string) map[string][]string {
	result := make(map[string][]string)
	
	for _, suggestion := range suggestions {
		if strings.Contains(suggestion, "(from ") {
			// 提取模板名称
			start := strings.Index(suggestion, "(from ") + 6
			end := strings.Index(suggestion[start:], " template)")
			if end > 0 {
				templateName := suggestion[start : start+end]
				pattern := strings.Split(suggestion, "'")[1]
				result[templateName] = append(result[templateName], pattern)
			}
		} else {
			result["other"] = append(result["other"], suggestion)
		}
	}
	
	return result
}

// 辅助函数：计算非空类别数量
func countNonEmptyCategories(result AuditResult) int {
	count := 0
	if len(result.Danger) > 0 { count++ }
	if len(result.Redundancy) > 0 { count++ }
	if len(result.Portability) > 0 { count++ }
	if len(result.Performance) > 0 { count++ }
	return count
}