package template

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/utils"
	"github.com/spf13/cobra"
)

func GetTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "管理 .gitignore 模板",
		Long: `管理本地和 GitHub 源的 .gitignore 模板。

示例:
  gig template local list              # 列出所有本地模板
  gig template local list -n Go        # 查看 Go 模板内容
  gig template local search py         # 搜索本地模板
  gig template local add ./f           # 添加本地模板
  gig template local remove -n Go      # 删除本地模板
  gig template local edit              # 打开模板目录
  gig template local edit -n Go        # 编辑 Go 模板
  gig template github sync             # 克隆/更新 GitHub 模板
  gig template github reset            # 重置 GitHub 模板
  gig template github list             # 列出 GitHub 模板
  gig template github search py        # 搜索 GitHub 模板`,
	}

	cmd.AddCommand(getLocalCommand())
	cmd.AddCommand(getGithubCommand())

	return cmd
}

// ==================== local ====================

func getLocalCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local",
		Short: "管理本地模板",
	}

	cmd.AddCommand(getLocalListCommand())
	cmd.AddCommand(getLocalViewCommand())
	cmd.AddCommand(getLocalSearchCommand())
	cmd.AddCommand(getLocalAddCommand())
	cmd.AddCommand(getLocalCopyCommand())
	cmd.AddCommand(getLocalRemoveCommand())
	cmd.AddCommand(getLocalEditCommand())

	return cmd
}

func getLocalListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有本地模板",
		Run: func(cmd *cobra.Command, args []string) {
			names := collectTemplateNames(localTemplateDir())
			if len(names) == 0 {
				fmt.Println("没有本地模板。")
				return
			}
			sort.Strings(names)
			fmt.Printf("本地模板 (%d):\n  %s\n", len(names), strings.Join(names, ", "))
		},
	}
}

func getLocalViewCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "view",
		Short: "查看本地模板内容",
		Long: `查看指定本地模板的完整内容。

示例:
  gig template local view -n Go
  gig template local view -n Python`,
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				fmt.Println("请指定模板名称：gig template local view -n <name>")
				return
			}
			content, err := utils.GetTemplateContent(name, localTemplateDir())
			if err != nil {
				fmt.Printf("找不到模板 '%s'：%v\n", name, err)
				return
			}
			fmt.Println(content)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "要查看的模板名称")
	return cmd
}

func getLocalSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search <keyword>",
		Short: "搜索本地模板",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			matches := searchTemplates(localTemplateDir(), args[0])
			if len(matches) == 0 {
				fmt.Printf("没有找到与 '%s' 匹配的本地模板。\n", args[0])
				return
			}
			sort.Strings(matches)
			fmt.Printf("搜索 '%s' 的结果 (%d):\n  %s\n", args[0], len(matches), strings.Join(matches, "\n  "))
		},
	}
}

func getLocalAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <file>",
		Short: "添加本地模板文件",
		Long: `将一个 .gitignore 模板文件复制到本地模板目录。

示例:
  gig template local add ./my-custom.gitignore
  gig template local add ~/Downloads/Python.gitignore`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			srcPath := args[0]
			absSrc, err := filepath.Abs(srcPath)
			if err != nil {
				fmt.Printf("路径解析失败：%v\n", err)
				return
			}

			info, err := os.Stat(absSrc)
			if err != nil {
				fmt.Printf("文件不存在：%s\n", absSrc)
				return
			}
			if info.IsDir() {
				fmt.Println("错误：请指定一个文件，不是目录。")
				return
			}

			content, err := utils.ReadFile(absSrc)
			if err != nil {
				fmt.Printf("读取文件失败：%v\n", err)
				return
			}

			baseName := filepath.Base(absSrc)
			if !strings.HasSuffix(baseName, ".gitignore") {
				baseName = baseName + ".gitignore"
			}

			tmplDir := localTemplateDir()
			os.MkdirAll(tmplDir, os.ModePerm)
			destPath := filepath.Join(tmplDir, baseName)

			if _, err := os.Stat(destPath); err == nil {
				fmt.Printf("模板 '%s' 已存在。\n", baseName)
				return
			}

			if err := utils.WriteFile(destPath, content); err != nil {
				fmt.Printf("写入模板失败：%v\n", err)
				return
			}

			fmt.Printf("已添加本地模板：%s\n", baseName)
		},
	}
}

func getLocalCopyCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "将 GitHub 模板复制到本地，便于二次编辑",
		Long: `从 GitHub 官方模板中复制一个到本地模板目录，方便自定义修改。
需要先运行 gig template github sync 同步 GitHub 模板。

示例:
  gig template local copy -n Python
  gig template local copy -n Go`,
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				fmt.Println("请指定模板名称：gig template local copy -n <name>")
				return
			}

			githubDir := githubTemplateDir()
			if _, err := os.Stat(githubDir); err != nil {
				fmt.Println("GitHub 模板尚未同步。请先运行：gig template github sync")
				return
			}

			content, err := utils.GetTemplateContent(name, githubDir)
			if err != nil {
				fmt.Printf("GitHub 模板 '%s' 不存在：%v\n", name, err)
				return
			}

			tmplDir := localTemplateDir()
			os.MkdirAll(tmplDir, os.ModePerm)

			fileName := name
			if !strings.HasSuffix(fileName, ".gitignore") {
				fileName = fileName + ".gitignore"
			}
			destPath := filepath.Join(tmplDir, fileName)

			if _, err := os.Stat(destPath); err == nil {
				fmt.Printf("本地模板 '%s' 已存在，跳过。\n", name)
				return
			}

			if err := utils.WriteFile(destPath, content); err != nil {
				fmt.Printf("复制失败：%v\n", err)
				return
			}

			fmt.Printf("已将 GitHub 模板 '%s' 复制到本地。\n", name)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "要复制的 GitHub 模板名称")
	return cmd
}

func getLocalRemoveCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "删除本地模板",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				fmt.Println("请指定要删除的模板名称：gig template local remove -n <name>")
				return
			}

			fileName := name
			if !strings.HasSuffix(fileName, ".gitignore") {
				fileName = fileName + ".gitignore"
			}

			templatePath := filepath.Join(localTemplateDir(), fileName)
			if _, err := os.Stat(templatePath); err != nil {
				fmt.Printf("本地模板 '%s' 不存在。\n", name)
				return
			}

			if err := os.Remove(templatePath); err != nil {
				fmt.Printf("删除模板失败：%v\n", err)
				return
			}

			fmt.Printf("已删除本地模板：%s\n", name)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "要删除的模板名称")
	return cmd
}

func getLocalEditCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "用编辑器打开模板，-n 指定模板",
		Run: func(cmd *cobra.Command, args []string) {
			editor := os.Getenv("VISUAL")
			if editor == "" {
				editor = os.Getenv("EDITOR")
			}

			var target string
			if name != "" {
				fileName := name
				if !strings.HasSuffix(fileName, ".gitignore") {
					fileName = fileName + ".gitignore"
				}
				target = filepath.Join(localTemplateDir(), fileName)
				if _, err := os.Stat(target); err != nil {
					fmt.Printf("本地模板 '%s' 不存在。\n", name)
					return
				}
			} else {
				target = localTemplateDir()
				if _, err := os.Stat(target); err != nil {
					fmt.Println("本地模板目录不存在。")
					return
				}
			}

			var cmdExec *exec.Cmd
			if editor != "" {
				parts := strings.Fields(editor)
				if len(parts) == 0 {
					switch runtime.GOOS {
					case "darwin":
						cmdExec = exec.Command("open", target)
					case "windows":
						cmdExec = exec.Command("explorer", target)
					default:
						cmdExec = exec.Command("xdg-open", target)
					}
				} else {
					cmdExec = exec.Command(parts[0], append(parts[1:], target)...)
				}
			} else {
				switch runtime.GOOS {
				case "darwin":
					cmdExec = exec.Command("open", target)
				case "windows":
					cmdExec = exec.Command("explorer", target)
				default:
					cmdExec = exec.Command("xdg-open", target)
				}
			}

			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			if err := cmdExec.Run(); err != nil {
				fmt.Printf("打开编辑器失败：%v\n", err)
			}
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "指定要编辑的模板名称")
	return cmd
}

// ==================== github ====================

func getGithubCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: "管理 GitHub 官方模板",
	}

	cmd.AddCommand(getGithubSyncCommand())
	cmd.AddCommand(getGithubResetCommand())
	cmd.AddCommand(getGithubListCommand())
	cmd.AddCommand(getGithubViewCommand())
	cmd.AddCommand(getGithubSearchCommand())

	return cmd
}

func getGithubSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "克隆或拉取最新 GitHub 官方模板",
		Long: `首次运行克隆 github/gitignore 仓库，之后每次运行执行 git pull 更新。`,
		Run: func(cmd *cobra.Command, args []string) {
			remoteDir, err := common.EnsureRemoteTemplatesAreCloned()
			if err != nil {
				fmt.Printf("同步失败：%v\n", err)
				return
			}
			names := collectTemplateNames(remoteDir)
			sort.Strings(names)
			fmt.Printf("共 %d 个 GitHub 模板。\n", len(names))
		},
	}
}

func getGithubResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "删除并重新克隆 GitHub 官方模板",
		Long: `删除本地缓存的 GitHub 模板目录，重新克隆整个仓库。
适用于仓库损坏或需要完整重置的场景。`,
		Run: func(cmd *cobra.Command, args []string) {
			remoteDir, err := common.ResetRemoteTemplates()
			if err != nil {
				fmt.Printf("重置失败：%v\n", err)
				return
			}
			names := collectTemplateNames(remoteDir)
			sort.Strings(names)
			fmt.Printf("已重置，共 %d 个 GitHub 模板。\n", len(names))
		},
	}
}

func getGithubListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出 GitHub 模板",
		Run: func(cmd *cobra.Command, args []string) {
			githubDir := githubTemplateDir()
			if _, err := os.Stat(githubDir); err != nil {
				fmt.Println("GitHub 模板尚未同步。请先运行：gig template github sync")
				return
			}

			names := collectTemplateNames(githubDir)
			if len(names) == 0 {
				fmt.Println("没有 GitHub 模板。")
				return
			}
			sort.Strings(names)
			fmt.Printf("GitHub 模板 (%d):\n  %s\n", len(names), strings.Join(names, ", "))
		},
	}
}

func getGithubViewCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "view",
		Short: "查看 GitHub 模板内容",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				fmt.Println("请指定模板名称：gig template github view -n <name>")
				return
			}
			githubDir := githubTemplateDir()
			if _, err := os.Stat(githubDir); err != nil {
				fmt.Println("GitHub 模板尚未同步。请先运行：gig template github sync")
				return
			}
			content, err := utils.GetTemplateContent(name, githubDir)
			if err != nil {
				fmt.Printf("找不到模板 '%s'：%v\n", name, err)
				return
			}
			fmt.Println(content)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "要查看的模板名称")
	return cmd
}

func getGithubSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search <keyword>",
		Short: "搜索 GitHub 模板",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			githubDir := githubTemplateDir()
			if _, err := os.Stat(githubDir); err != nil {
				fmt.Println("GitHub 模板尚未同步。请先运行：gig template github sync")
				return
			}

			matches := searchTemplates(githubDir, args[0])
			if len(matches) == 0 {
				fmt.Printf("没有找到与 '%s' 匹配的 GitHub 模板。\n", args[0])
				return
			}
			sort.Strings(matches)
			fmt.Printf("搜索 '%s' 的结果 (%d):\n  %s\n", args[0], len(matches), strings.Join(matches, "\n  "))
		},
	}
}

// ==================== helpers ====================

func localTemplateDir() string {
	home, _ := utils.GetUserHomeDir()
	return filepath.Join(home, ".cicbyte", "gig", "template")
}

func githubTemplateDir() string {
	home, _ := utils.GetUserHomeDir()
	return filepath.Join(home, ".cicbyte", "gig", "template_github")
}

func collectTemplateNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".gitignore") {
			names = append(names, strings.TrimSuffix(name, ".gitignore"))
		}
	}
	return names
}

func searchTemplates(dir string, keyword string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var matches []string
	lowerKeyword := strings.ToLower(keyword)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".gitignore") {
			templateName := strings.TrimSuffix(name, ".gitignore")
			if strings.Contains(strings.ToLower(templateName), lowerKeyword) {
				matches = append(matches, templateName)
			}
		}
	}
	return matches
}
