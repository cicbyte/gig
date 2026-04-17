package gitignore

import (
	"bufio"
	"os"
	"strings"

	"github.com/cicbyte/gig/internal/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// GitignoreUpdater 处理.gitignore文件更新
type GitignoreUpdater struct {
	Path string
}

// NewGitignoreUpdater 创建更新器
func NewGitignoreUpdater(path string) *GitignoreUpdater {
	return &GitignoreUpdater{
		Path: path,
	}
}

// ReadContent 读取.gitignore内容
func (u *GitignoreUpdater) ReadContent() (string, error) {
	content, err := utils.ReadFile(u.Path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return string(content), nil
}

// MergeContent 合并新内容到现有.gitignore
func (u *GitignoreUpdater) MergeContent(newContent string) (string, int, error) {
	originalContent, err := u.ReadContent()
	if err != nil {
		return "", 0, err
	}

	existingRules := ExtractRules(originalContent)

	var newRulesBuilder strings.Builder
	newRulesCount := 0
	lastLineEmpty := false
	seenInNewBatch := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(newContent))
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" && !existingRules[trimmedLine] && !seenInNewBatch[trimmedLine] {
			existingRules[trimmedLine] = true
			seenInNewBatch[trimmedLine] = true
			newRulesBuilder.WriteString(line + "\n")
			newRulesCount++
			lastLineEmpty = false
		} else if trimmedLine == "" && !lastLineEmpty {
			newRulesBuilder.WriteString("\n")
			lastLineEmpty = true
		}
	}

	var finalContentBuilder strings.Builder
	finalContentBuilder.WriteString(originalContent)
	// 如果原始内容不是以空行结尾，先补一行空行
	if len(originalContent) > 0 && !strings.HasSuffix(originalContent, "\n\n") {
		if !strings.HasSuffix(originalContent, "\n") {
			finalContentBuilder.WriteString("\n")
		}
		finalContentBuilder.WriteString("\n")
	}
	finalContentBuilder.WriteString(newRulesBuilder.String())

	return finalContentBuilder.String(), newRulesCount, nil
}

// ShowDiff 显示差异
func (u *GitignoreUpdater) ShowDiff(originalContent, newContent string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(originalContent, newContent, false)
	return dmp.DiffPrettyText(diffs)
}

// ApplyChanges 应用更改
func (u *GitignoreUpdater) ApplyChanges(content string) error {
	return utils.WriteFile(u.Path, content)
}
