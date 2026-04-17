package view

import (
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/spf13/cobra"
)

func GetViewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "查看当前项目的 .gitignore 文件",
		Long:  "查看当前目录下的 .gitignore 文件。如果当前目录没有，则向上查找至 Git 根目录。都没有则报错。",
		Run: func(cmd *cobra.Command, args []string) {
			gitignorePath, err := common.GetNearestGitignorePath(".")
			if err != nil {
				fmt.Fprintln(os.Stderr, "错误：不在 Git 仓库中")
				return
			}

			data, err := os.ReadFile(gitignorePath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintln(os.Stderr, "错误：未找到 .gitignore 文件")
				} else {
					fmt.Fprintf(os.Stderr, "错误：读取 .gitignore 失败: %v\n", err)
				}
				return
			}

			fmt.Println("文件:", gitignorePath)
			fmt.Println(strings.Repeat("─", 40))
			fmt.Print(string(data))
		},
	}
}
