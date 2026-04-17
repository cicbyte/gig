package common

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cicbyte/gig/internal/models"
)

// AppConfigModel 全局应用配置实例
var AppConfigModel *models.Config

// gitRoot 缓存 Git 仓库根目录
var gitRoot string

// CommandRunner 定义命令运行接口，便于测试 mock
type CommandRunner interface {
	Run(cmd *exec.Cmd) error
	Output(cmd *exec.Cmd) ([]byte, error)
}

// LiveRunner 是真实执行命令的实现
type LiveRunner struct{}

func (r LiveRunner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (r LiveRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

// GitCommandRunner 当前命令运行器，默认为真实执行
var GitCommandRunner CommandRunner = LiveRunner{}

// IsFileTracked 检查文件是否被 Git 追踪
func IsFileTracked(filePath string) bool {
	cmd := exec.Command("git", "ls-files", "--error-unmatch", filePath)
	err := GitCommandRunner.Run(cmd)
	return err == nil
}

// IsPathIgnored 检查路径是否被 .gitignore 忽略
func IsPathIgnored(filePath string) bool {
	cmd := exec.Command("git", "check-ignore", "-q", filePath)
	err := GitCommandRunner.Run(cmd)
	return err == nil
}

// GetIgnoringRule 获取忽略指定文件的规则
func GetIgnoringRule(filePath string) (string, string, string, error) {
	cmd := exec.Command("git", "check-ignore", "-v", filePath)
	output, err := GitCommandRunner.Output(cmd)
	if err != nil {
		return "", "", "", err
	}
	return parseGitCheckIgnoreOutput(output)
}

// parseGitCheckIgnoreOutput 解析 git check-ignore -v 的输出，返回第一条匹配
func parseGitCheckIgnoreOutput(output []byte) (string, string, string, error) {
	trimmedOutput := strings.TrimSpace(string(output))
	parts := strings.SplitN(trimmedOutput, "\t", 2)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("无法解析 git 输出")
	}

	details := strings.SplitN(parts[0], ":", 3)
	if len(details) < 3 {
		return "", "", "", fmt.Errorf("无法解析 git 输出详情")
	}

	return details[0], details[1], details[2], nil
}

// EnsureRemoteTemplatesAreCloned 检查远程模板是否已克隆，否则执行克隆
func EnsureRemoteTemplatesAreCloned() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("找不到主目录：%w", err)
	}
	remoteDir := filepath.Join(home, ".cicbyte", "gig", "template_github")

	if _, err := os.Stat(remoteDir); os.IsNotExist(err) {
		fmt.Printf("正在克隆 GitHub 官方模板...\n")
		cmd := exec.Command("git", "clone", remoteTemplateURL, remoteDir)
		if err := cmd.Run(); err != nil {
			fmt.Printf("\n克隆失败：%v\n", err)
			fmt.Println("你可以手动完成以下操作：")
			fmt.Printf("  1. 克隆仓库：git clone %s\n", remoteTemplateURL)
			fmt.Printf("  2. 将仓库内容放置到：%s\n", remoteDir)
			fmt.Println("  3. 重新运行当前命令")
			return "", fmt.Errorf("克隆远程仓库失败：%w", err)
		}
		fmt.Println("成功克隆官方模板。")
		return remoteDir, nil
	}

	// 已克隆，执行 pull
	fmt.Println("正在更新 GitHub 官方模板...")
	cmd := exec.Command("git", "-C", remoteDir, "pull")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("更新 GitHub 模板失败：%w", err)
	}
	fmt.Println("已更新官方模板。")
	return remoteDir, nil
}

// ResetRemoteTemplates 删除并重新克隆远程模板
func ResetRemoteTemplates() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("找不到主目录：%w", err)
	}
	remoteDir := filepath.Join(home, ".cicbyte", "gig", "template_github")

	if _, err := os.Stat(remoteDir); err == nil {
		fmt.Println("正在删除旧的 GitHub 模板...")
		if err := os.RemoveAll(remoteDir); err != nil {
			return "", fmt.Errorf("删除旧模板失败：%w", err)
		}
	}

	fmt.Println("正在重新克隆 GitHub 官方模板...")
	cmd := exec.Command("git", "clone", remoteTemplateURL, remoteDir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("\n克隆失败：%v\n", err)
		fmt.Println("你可以手动完成以下操作：")
		fmt.Printf("  1. 克隆仓库：git clone %s\n", remoteTemplateURL)
		fmt.Printf("  2. 将仓库内容放置到：%s\n", remoteDir)
		fmt.Println("  3. 重新运行当前命令")
		return "", fmt.Errorf("重新克隆失败：%w", err)
	}
	fmt.Println("成功重新克隆官方模板。")
	return remoteDir, nil
}

// PromptUserForConfirmation 提示用户确认
func PromptUserForConfirmation(message string) bool {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(input)) == "y"
}

const remoteTemplateURL = "https://github.com/github/gitignore.git"

// GetGitRoot 获取 Git 仓库根目录，未找到则报错
func GetGitRoot() (string, error) {
	if gitRoot != "" {
		return gitRoot, nil
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("当前目录不在 Git 仓库中，无法执行此操作")
	}
	gitRoot = strings.TrimSpace(string(output))
	return gitRoot, nil
}

// GetGitignorePath 获取仓库根目录下的 .gitignore 路径
func GetGitignorePath() (string, error) {
	root, err := GetGitRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".gitignore"), nil
}

// GetNearestGitignorePath 从指定目录向上查找最近的 .gitignore，找不到则返回根目录的 .gitignore
func GetNearestGitignorePath(dir string) (string, error) {
	root, err := GetGitRoot()
	if err != nil {
		return "", err
	}

	current, err := filepath.Abs(dir)
	if err != nil {
		current = dir
	}

	for {
		candidate := filepath.Join(current, ".gitignore")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		if current == root {
			break
		}
		current = filepath.Dir(current)
	}

	return filepath.Join(root, ".gitignore"), nil
}

// ResetGitRootCache 重置缓存（测试用）
func ResetGitRootCache() {
	gitRoot = ""
}
