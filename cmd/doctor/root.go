package doctor

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/gitignore"
	"github.com/cicbyte/gig/internal/logic/doctor"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var doctorFixFlag bool
var doctorYesFlag bool
var doctorDryRunFlag bool

func GetDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "诊断 .gitignore 文件，发现潜在问题。",
		Long: `诊断 .gitignore 文件，对照最佳实践和项目上下文进行分析。

使用规则引擎检查重复、危险模式、性能问题和全局冲突。
结合 -t ai 可额外获取 AI 诊断建议。
使用 --fix 可自动修复可修复的问题。`,
		Run: func(cmd *cobra.Command, args []string) {
			source, _ := cmd.Flags().GetString("type")
			config := &doctor.AuditConfig{Source: source}
			processor := doctor.NewAuditProcessor(config, common.AppConfigModel)

			// AI 模式：先打印标题，再流式输出
			if source == "ai" {
				projectTypes := doctor.PeekProjectTypes(processor)
				fmt.Printf("项目类型：%s\n", strings.Join(projectTypes, ", "))
				fmt.Println(strings.Repeat("─", 50))
			}

			result, err := processor.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("诊断失败：%v\n", err)
				return
			}

			// 非 AI 模式：打印标题
			if !result.Streamed {
				fmt.Printf("项目类型：%s\n", strings.Join(result.ProjectTypes, ", "))
				fmt.Println(strings.Repeat("─", 50))
			}

			// 渲染规则引擎问题表格
			renderIssueTable(result.Issues)

			// AI 模式：追加 AI 报告
			if result.AIReport != "" {
				fmt.Println(strings.Repeat("─", 50))
				fmt.Println("AI 诊断：")
				fmt.Println(result.AIReport)
			}

			// --fix 模式
			if doctorFixFlag {
				fixableCount := 0
				for _, issue := range result.Issues {
					if issue.Fixable {
						fixableCount++
					}
				}
				if fixableCount == 0 {
					fmt.Println("\n没有可自动修复的问题。")
					return
				}

				fixedContent, fixedCount, err := processor.Fix(cmd.Context(), result)
				if err != nil {
					fmt.Printf("修复失败：%v\n", err)
					return
				}

				// 展示 diff
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(result.Content, fixedContent, false)
				diffText := dmp.DiffPrettyText(diffs)

				fmt.Printf("\n将修复 %d 个问题，变更如下：\n", fixedCount)
				fmt.Println(diffText)

				if doctorDryRunFlag {
					fmt.Println("(dry-run 模式，未实际修改)")
					return
				}

				if doctorYesFlag {
					fmt.Println("已自动确认修复 (--yes)")
				} else {
					fmt.Print("是否应用修复？(y/n): ")
					reader := bufio.NewReader(os.Stdin)
					input, _ := reader.ReadString('\n')
					if strings.TrimSpace(strings.ToLower(input)) != "y" {
						fmt.Println("操作已取消。")
						return
					}
				}

				gitignorePath, err := common.GetGitignorePath()
				if err != nil {
					fmt.Fprintf(os.Stderr, "错误：%v\n", err)
					return
				}
				updater := gitignore.NewGitignoreUpdater(gitignorePath)
				if err := updater.ApplyChanges(fixedContent); err != nil {
					fmt.Fprintf(os.Stderr, "写入失败：%v\n", err)
					return
				}
				fmt.Printf("已修复 %d 个问题。\n", fixedCount)
			}
		},
	}
	cmd.Flags().BoolVar(&doctorFixFlag, "fix", false, "自动修复可修复的问题")
	cmd.Flags().BoolVar(&doctorYesFlag, "yes", false, "自动确认修复，无需人工输入")
	cmd.Flags().BoolVar(&doctorDryRunFlag, "dry-run", false, "仅展示修复内容，不实际修改")
	return cmd
}

// renderIssueTable 渲染问题表格
func renderIssueTable(issues []gitignore.AuditIssue) {
	if len(issues) == 0 {
		fmt.Println("[OK] 未发现问题")
		return
	}

	severityLabel := map[string]string{
		"error":   "严重",
		"warning": "警告",
		"info":    "提示",
	}

	fixableCount := 0
	for _, issue := range issues {
		label := severityLabel[issue.Severity]
		lineInfo := ""
		if issue.LineNum > 0 {
			lineInfo = fmt.Sprintf("[%d]", issue.LineNum)
		}
		fixTag := ""
		if issue.Fixable {
			fixTag = "  可修复"
			fixableCount++
		}
		fmt.Printf("  %s %s %-40s%s\n", label, lineInfo, truncateStr(issue.Message, 40), fixTag)
	}

	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("共 %d 个问题", len(issues))
	if fixableCount > 0 {
		fmt.Printf("，其中 %d 个可通过 --fix 自动修复", fixableCount)
	}
	fmt.Println()
}

func truncateStr(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}
