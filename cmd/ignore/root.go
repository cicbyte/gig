package ignore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/gitignore"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var ignoreYesFlag bool

func GetIgnoreCommand() *cobra.Command {
	var raw bool
	cmd := &cobra.Command{
		Use:   "ignore <path>...",
		Short: "将文件或文件夹添加到 .gitignore",
		Long: `将文件或文件夹添加到 .gitignore，同时处理 Git 跟踪状态。

智能行为：
  - 文件已被跟踪 → 自动执行 git rm --cached，然后添加忽略规则
  - 文件未被跟踪 → 直接添加忽略规则
  - 规则已存在且 Git 确实忽略 → 跳过
  - 存在否定规则覆盖 → 自动移除否定规则

默认行为：
  - 自动检测路径是否存在，不存在则报错
  - 目录自动追加 /（如 build → build/）
  - 含通配符的模式原样添加

使用 --raw 跳过路径检测，原样写入任意模式。

示例:
  gig ignore build/              # 目录自动补 /
  gig ignore dist node_modules   # 多个路径
  gig ignore *.log               # 通配符原样添加
  gig ignore --raw "*.exe"       # 原样写入，不做任何检测
  gig ignore .env                # 已跟踪的文件自动取消跟踪`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, arg := range args {
				handleIgnore(arg, raw)
			}
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "原样写入，不做路径检测和格式化")
	cmd.Flags().BoolVar(&ignoreYesFlag, "yes", false, "自动确认所有交互，无需人工输入")
	return cmd
}

func handleIgnore(input string, raw bool) {
	fmt.Printf("--- 正在处理：%s ---\n", input)

	rule, err := resolveRule(input, raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		fmt.Println("-------------------------")
		return
	}
	if rule == "" {
		fmt.Println("跳过。")
		fmt.Println("-------------------------")
		return
	}

	// 用 git check-ignore 判断路径是否已被实际忽略
	alreadyIgnored := common.IsPathIgnored(input)

	if alreadyIgnored {
		fmt.Printf("'%s' 已被 Git 忽略，无需操作。\n", rule)

		// 即使已被忽略，仍需清理可能存在的否定规则（避免影响其他目录）
		negationRule := "!" + rule
		if removed, _ := findAndRemoveNegationRule(negationRule); removed {
			fmt.Printf("[OK] 已清理否定规则 '%s'，'%s' 仍保持忽略状态。\n", negationRule, rule)
		}

		// 已忽略但文件仍被跟踪，询问是否取消跟踪
		if common.IsFileTracked(input) {
			if confirmGitRm(input) {
				if err := gitRmCached(input); err != nil {
					fmt.Fprintf(os.Stderr, "取消跟踪失败：%v\n", err)
				} else {
					fmt.Printf("[OK] 已取消跟踪 '%s'。\n", input)
				}
			}
		}
		fmt.Println("-------------------------")
		return
	}

	// 路径未被忽略，在所有 .gitignore 文件中搜索并清理否定规则
	negationRule := "!" + rule
	if removed, removedFrom := findAndRemoveNegationRule(negationRule); removed {
		// 移除后再次检查是否已被忽略
		if common.IsPathIgnored(input) {
			fmt.Printf("[OK] '%s' 现在已被 Git 忽略。\n", rule)
			fmt.Println("-------------------------")
			return
		}
		// 移除了否定规则但仍未被忽略（可能缺少对应的正向规则），继续添加正向规则
		_ = removedFrom
	}

	// 获取写入目标 .gitignore（最近的）
	gitignorePath, err := common.GetNearestGitignorePath(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		fmt.Println("-------------------------")
		return
	}

	existingContent, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "读取 .gitignore 失败：%v\n", err)
		fmt.Println("-------------------------")
		return
	}

	// 写入忽略规则
	var content []byte
	if len(existingContent) > 0 {
		content = existingContent
		if content[len(content)-1] != '\n' {
			content = append(content, '\n')
		}
	} else {
		content = []byte("# --- 由 gig ignore 添加 ---\n")
	}
	content = append(content, []byte(rule+"\n")...)

	if err := os.WriteFile(gitignorePath, content, 0660); err != nil {
		fmt.Fprintf(os.Stderr, "写入 .gitignore 失败：%v\n", err)
		fmt.Println("-------------------------")
		return
	}

	relPath, _ := filepath.Rel(".", gitignorePath)
	fmt.Printf("[OK] 已添加 '%s' 到 %s\n", rule, relPath)

	// 如果文件被 Git 跟踪，询问是否取消跟踪
	if common.IsFileTracked(input) {
		if confirmGitRm(input) {
			if err := gitRmCached(input); err != nil {
				fmt.Fprintf(os.Stderr, "取消跟踪失败：%v\n", err)
			} else {
				fmt.Printf("[OK] 已取消跟踪 '%s'。\n", input)
				fmt.Println("下一步：请提交此更改。")
			}
		}
	}

	fmt.Println("-------------------------")
}

// removeLine 从 content 中移除包含指定文本的行
func removeLine(content, target string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		if strings.TrimSpace(line) == target {
			continue
		}
		filtered = append(filtered, line)
	}
	// 清理尾部连续空行
	for len(filtered) > 0 && strings.TrimSpace(filtered[len(filtered)-1]) == "" {
		filtered = filtered[:len(filtered)-1]
	}
	return strings.Join(filtered, "\n") + "\n"
}

func confirmGitRm(filePath string) bool {
	if ignoreYesFlag {
		return true
	}
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("文件 '%s' 已被 Git 跟踪，是否取消跟踪（保留本地文件）？", filePath),
		IsConfirm: true,
	}
	_, err := prompt.Run()
	return err == nil
}

func gitRmCached(filePath string) error {
	cmd := exec.Command("git", "rm", "--cached", "--", filePath)
	return cmd.Run()
}

// resolveRule 解析并验证规则，返回标准化后的规则或错误
func resolveRule(input string, raw bool) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil
	}

	// raw 模式：原样写入
	if raw {
		return input, nil
	}

	// 含通配符的模式：原样添加
	if strings.ContainsAny(input, "*?[") {
		return input, nil
	}

	// 检查路径是否存在
	info, err := os.Stat(input)
	if err != nil {
		return "", fmt.Errorf("'%s' 不存在，请检查路径。如需添加自定义模式请使用 --raw", input)
	}

	// 目录：只取目录名，追加 /
	if info.IsDir() {
		dirName := filepath.Base(input)
		return dirName + "/", nil
	}

	// 文件：只取文件名
	fileName := filepath.Base(input)
	return fileName, nil
}

// findAndRemoveNegationRule 从当前目录向上遍历所有 .gitignore 文件，查找并移除指定的否定规则。
// 返回是否成功移除，以及来源文件路径。
func findAndRemoveNegationRule(negationRule string) (bool, string) {
	root, err := common.GetGitRoot()
	if err != nil {
		return false, ""
	}

	current, err := filepath.Abs(".")
	if err != nil {
		return false, ""
	}

	for {
		candidate := filepath.Join(current, ".gitignore")
		content, err := os.ReadFile(candidate)
		if err == nil && len(content) > 0 {
			existingRules := gitignore.ExtractRules(string(content))
			if existingRules[negationRule] {
				relPath, relErr := filepath.Rel(".", candidate)
				if relErr != nil || relPath == "" {
					relPath = candidate
				}
				fmt.Printf("检测到否定规则 '%s'（%s），正在移除...\n", negationRule, relPath)
				newContent := removeLine(string(content), negationRule)
				if err := os.WriteFile(candidate, []byte(newContent), 0660); err != nil {
					fmt.Fprintf(os.Stderr, "移除否定规则失败：%v\n", err)
					return false, ""
				}
				fmt.Printf("[OK] 已移除否定规则 '%s'\n", negationRule)
				return true, candidate
			}
		}

		if current == root {
			break
		}
		current = filepath.Dir(current)
	}

	return false, ""
}
