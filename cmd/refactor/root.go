package refactor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/logic/refactor"
	"github.com/spf13/cobra"
)

var refactorYesFlag bool
var refactorDryRunFlag bool

func GetRefactorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refactor",
		Short: "整理 .gitignore 文件：去重、分类、优化格式。",
		Long: `读取当前 .gitignore 文件，使用 AI 进行整理优化。

整理规则：
- 移除重复规则
- 按类别重新分组（依赖、构建产物、IDE、日志等）
- 优化规则格式
- 不新增或删除任何规则

此命令需要配置 AI。`,
		Run: func(cmd *cobra.Command, args []string) {
			processor := refactor.NewRefactorProcessor(common.AppConfigModel)

			result, err := processor.Execute(context.Background())
			if err != nil {
				fmt.Fprintln(os.Stderr, "重构失败：", err)
				return
			}

			fmt.Println("--- 整理后的内容 ---")
			fmt.Println(result.DiffText)
			fmt.Println("--- 内容结束 ---")

			if refactorDryRunFlag {
				fmt.Println("(dry-run 模式，未实际修改文件)")
				return
			}

			if refactorYesFlag {
				fmt.Println("\n已自动确认并应用更改 (--yes)")
			} else {
				fmt.Print("\n是否应用更改？(y/n): ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(input)) != "y" {
					fmt.Println("操作已取消。")
					return
				}
			}

			if err := processor.Apply(result.RefactoredContent); err != nil {
				fmt.Fprintln(os.Stderr, "写入 .gitignore 文件时出错：", err)
				return
			}

			fmt.Println(".gitignore 已整理完成。")
		},
	}
	cmd.Flags().BoolVar(&refactorYesFlag, "yes", false, "自动确认，无需人工输入")
	cmd.Flags().BoolVar(&refactorDryRunFlag, "dry-run", false, "仅显示整理结果，不实际修改")
	return cmd
}
