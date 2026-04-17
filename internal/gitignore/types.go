package gitignore

// Rule 表示一条.gitignore规则
type Rule struct {
	Pattern string // 规则模式
	LineNum int    // 行号
	Raw     string // 原始行内容
}

// Issue 表示一个规则问题（内部使用）
type Issue struct {
	Category string // 问题类别: danger, redundancy, portability, performance
	Message  string // 问题描述
	LineNum  int    // 行号
	Pattern  string // 相关规则
}

// AuditIssue 统一的审计问题（面向输出）
type AuditIssue struct {
	Severity string // "error", "warning", "info"
	Category string // "danger", "redundancy", "portability", "performance", "missing"
	Message  string
	LineNum  int    // 0 表示无具体行号
	Pattern  string
	Fixable  bool
}

// AuditResult 保存分类后的审计结果
type AuditResult struct {
	Danger      []Issue
	Redundancy  []Issue
	Portability []Issue
	Performance []Issue
}

// ReportFormat 定义报告格式
type ReportFormat string

const (
	FormatConsole ReportFormat = "console"
	FormatJSON    ReportFormat = "json"
	FormatHTML    ReportFormat = "html"
)
