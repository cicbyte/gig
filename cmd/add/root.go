package add

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/logic/add"
	"github.com/cicbyte/gig/internal/models"
	"github.com/cicbyte/gig/internal/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var addYesFlag bool
var addUpdateFlag bool
var addTUIFlag bool
var addSaveFlag bool
var addDryRunFlag bool
var addInteractiveFlag bool

func GetAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [languages...]",
		Short: "向 .gitignore 文件添加规则。",
		Long: `向 .gitignore 文件添加特定语言的规则。

您可以直接指定语言名称，或使用 -i 进入交互式向导。
该命令将从模板中获取规则，合并去重后写入 .gitignore。
如果 .gitignore 不存在，会自动创建。

示例 (直接指定):
  gig add go python node
  gig add Go Python -t github
  gig add go,python -t ai
  gig add rust -t ai --save

示例 (交互式):
  gig add -i

示例 (预览):
  gig add go python --dry-run`,
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var languages []string

			if addInteractiveFlag {
				// 交互模式
				detectedTypes := utils.DetectionUtil.DetectProjectTypes(common.AppConfigModel.Detection.FileMap)
				projectTemplates, err := promptForProjectTemplates(detectedTypes)
				if err != nil {
					fmt.Fprintf(os.Stderr, "错误：%v\n", err)
					return
				}
				envTemplates, err := promptForEnvTemplates()
				if err != nil {
					fmt.Fprintf(os.Stderr, "错误：%v\n", err)
					return
				}
				languages = append(projectTemplates, envTemplates...)
			} else if len(args) == 0 {
				// 无参数且非交互：自动检测并确认
				detected := utils.DetectionUtil.DetectProjectTypes(models.AppConfig.Detection.FileMap)
				if len(detected) > 0 {
					languages = promptForTypes(detected, addYesFlag)
				}
			} else {
				languages = args
			}

			if len(languages) == 0 {
				fmt.Println("未指定任何模板。使用 gig add -i 进入交互模式，或直接指定语言，例如：gig add python")
				return
			}

			source, _ := cmd.Flags().GetString("type")
			config := &add.AddConfig{
				Languages: languages,
				Source:    source,
				Update:    addUpdateFlag,
				Yes:       addYesFlag,
				TUI:       addTUIFlag,
				Save:      addSaveFlag,
			}
			processor := add.NewAddProcessor(config, common.AppConfigModel)
			result, err := processor.Execute(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误：%v\n", err)
				return
			}

			if result.NewRulesCount == 0 {
				fmt.Println("没有新的规则需要添加。.gitignore 文件已是最新。")
				return
			}

			if addDryRunFlag {
				fmt.Println("(dry-run 模式，未实际修改文件)")
				fmt.Println(result.DiffText)
				return
			}

			if addYesFlag {
				fmt.Println("已自动确认并应用更改 (--yes)")
			} else if addTUIFlag {
				confirmed, err := utils.ShowGitignoreDiff(result.OriginalContent, result.FinalContent)
				if err != nil {
					fmt.Printf("显示差异界面失败：%v\n", err)
					fmt.Println("将对 .gitignore 文件进行以下更改：")
					fmt.Println(result.DiffText)
					fmt.Print("您要应用这些更改吗？ (y/n): ")
					reader := bufio.NewReader(os.Stdin)
					input, _ := reader.ReadString('\n')
					if strings.TrimSpace(strings.ToLower(input)) != "y" {
						fmt.Println("操作已取消。")
						return
					}
				} else if !confirmed {
					fmt.Println("操作已取消。")
					return
				}
			} else {
				fmt.Println("将对 .gitignore 文件进行以下更改：")
				fmt.Println(result.DiffText)
				fmt.Print("您要应用这些更改吗？ (y/n): ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(input)) != "y" {
					fmt.Println("操作已取消。")
					return
				}
			}

			if err := processor.Apply(result.FinalContent); err != nil {
				fmt.Fprintln(os.Stderr, "写入 .gitignore 文件时出错：", err)
				return
			}

			fmt.Println(".gitignore 文件已成功更新。")
		},
	}
	cmd.Flags().BoolVar(&addYesFlag, "yes", false, "自动确认所有交互，无需人工输入")
	cmd.Flags().BoolVarP(&addUpdateFlag, "update", "u", false, "强制更新github模板")
	cmd.Flags().BoolVar(&addTUIFlag, "tui", false, "使用TUI界面显示差异对比")
	cmd.Flags().BoolVar(&addSaveFlag, "save", false, "将AI生成的规则保存为本地模板")
	cmd.Flags().BoolVar(&addDryRunFlag, "dry-run", false, "仅显示将要添加的内容，不实际修改")
	cmd.Flags().BoolVarP(&addInteractiveFlag, "interactive", "i", false, "进入交互式向导，引导选择项目模板和OS/IDE模板")
	return cmd
}

func promptForTypes(detected []string, yes bool) []string {
	if yes {
		fmt.Printf("检测到以下项目类型：%s\n", strings.Join(detected, ", "))
		fmt.Println("已自动确认项目类型 (--yes)")
		return detected
	}
	fmt.Printf("检测到以下项目类型：%s\n", strings.Join(detected, ", "))
	fmt.Print("是否确认？（直接回车确认，或输入新的类型列表，如：go,python）：")
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return detected
	}
	var result []string
	for _, t := range strings.Split(input, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

func promptForProjectTemplates(detected []string) ([]string, error) {
	fmt.Printf("检测到的项目类型：%s\n", strings.Join(detected, ", "))
	prompt := promptui.Prompt{
		Label:   "输入项目模板（逗号分隔，如 go,node）",
		Default: strings.Join(detected, ","),
	}
	result, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	return strings.Split(result, ","), nil
}

func promptForEnvTemplates() ([]string, error) {
	prompt := promptui.Prompt{
		Label: "输入操作系统/IDE 模板（逗号分隔，如 macos,vscode，直接回车跳过）",
	}
	result, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	if result == "" {
		return []string{}, nil
	}
	return strings.Split(result, ","), nil
}
