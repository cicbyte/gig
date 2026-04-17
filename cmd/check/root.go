package check

import (
	"context"
	"fmt"
	"os"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/logic/check"
	"github.com/spf13/cobra"
)

func GetCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check <file-path>",
		Short: "检查文件是否被 Git 忽略",
		Long: `检查指定文件是否被 .gitignore 忽略，显示所有匹配的规则。

Git 按优先级匹配：子目录 .gitignore > 父目录 .gitignore > 根目录 .gitignore。
结果按优先级从高到低排列。

示例:
  gig check .env
  gig check dist/bundle.js`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config := &check.CheckConfig{FilePath: args[0]}
			processor := check.NewCheckProcessor(config, common.AppConfigModel)
			result, err := processor.Execute(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误：%v\n", err)
				return
			}

			if !result.Ignored {
				fmt.Printf("[OK] 文件 '%s' 未被忽略。\n", result.FilePath)
				return
			}

			fmt.Printf("[X] 文件 '%s' 被忽略，共 %d 条规则匹配：\n", result.FilePath, len(result.Matches))
			for _, m := range result.Matches {
				fmt.Printf("  %s:%s  %s\n", m.SourceFile, m.LineNumber, m.Pattern)
			}
		},
	}
}
