package track

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/logic/track"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var trackYesFlag bool

func GetTrackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track [file-path...]",
		Short: "强制 Git 跟踪一个当前被忽略的文件。",
		Long: `通过向匹配的 .gitignore 添加否定规则，强制 Git 跟踪被忽略的文件。

自动定位到包含匹配规则的 .gitignore 文件（可能是子目录的），
在其中添加否定规则。`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, filePath := range args {
				fmt.Printf("--- 正在处理：%s ---\n", filePath)
				handleTrack(filePath)
				fmt.Println("-------------------------")
			}
		},
	}
	cmd.Flags().BoolVar(&trackYesFlag, "yes", false, "自动确认所有交互，无需人工输入")
	return cmd
}

func handleTrack(filePath string) {
	config := &track.TrackConfig{FilePath: filePath, Yes: trackYesFlag}
	processor := track.NewTrackProcessor(config, common.AppConfigModel)
	result, err := processor.Prepare(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		return
	}

	if result.AlreadyTracked {
		fmt.Printf("信息：文件 '%s' 已被 Git 跟踪。无需操作。\n", filePath)
		return
	}

	if result.NotIgnored {
		fmt.Printf("信息：文件 '%s' 未被忽略。正在将其添加到 Git...\n", filePath)
	} else {
		fmt.Printf("信息：文件被 %s（行 %s）的规则 '%s' 忽略。\n", result.SourceFile, result.LineNumber, result.Pattern)
		fmt.Printf("将在 %s 中添加否定规则 '%s'\n", result.SourceFile, result.ExceptionRule)

		relSource, _ := filepath.Rel(".", result.SourceFile)
		if relSource != result.SourceFile {
			fmt.Printf("（相对路径：%s）\n", relSource)
		}

		if err := processor.AddExceptionRule(result.SourceFile, result.ExceptionRule); err != nil {
			fmt.Printf("更新 .gitignore 时出错：%v\n", err)
			return
		}
		fmt.Printf("已添加否定规则。\n")
	}

	if !trackYesFlag {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("这将运行 'git add %s'。是否继续？", filePath),
			IsConfirm: true,
		}
		if _, err := prompt.Run(); err != nil {
			fmt.Println("操作已取消。")
			return
		}
	}

	fmt.Printf("已暂存 '%s'。\n", filePath)
	fmt.Println("\n下一步：请运行 'git commit' 来保存您的更改。")
}
