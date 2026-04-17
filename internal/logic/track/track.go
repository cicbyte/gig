package track

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/models"
)

type TrackConfig struct {
	FilePath string
	Yes      bool
}

type TrackResult struct {
	FilePath       string
	AlreadyTracked bool
	NotIgnored      bool
	SourceFile     string
	LineNumber     string
	Pattern        string
	ExceptionRule  string
}

type TrackProcessor struct {
	config    *TrackConfig
	appConfig *models.Config
}

func NewTrackProcessor(config *TrackConfig, appConfig *models.Config) *TrackProcessor {
	return &TrackProcessor{config: config, appConfig: appConfig}
}

// Prepare checks state and returns what action is needed (without executing git add)
func (p *TrackProcessor) Prepare(ctx context.Context) (*TrackResult, error) {
	filePath := p.config.FilePath

	if common.IsFileTracked(filePath) {
		return &TrackResult{FilePath: filePath, AlreadyTracked: true}, nil
	}

	if !common.IsPathIgnored(filePath) {
		return &TrackResult{FilePath: filePath, NotIgnored: true}, nil
	}

	source, line, pattern, err := common.GetIgnoringRule(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法确定 %s 的忽略规则：%w", filePath, err)
	}

	// 否定规则只取文件名，如 !node_modules/ 而非 !web/admin/node_modules
	baseName := filepath.Base(filePath)
	info, err := os.Stat(filePath)
	if err == nil && info.IsDir() {
		baseName = strings.TrimSuffix(baseName, string(filepath.Separator)) + "/"
	}

	return &TrackResult{
		FilePath:      filePath,
		SourceFile:    source,
		LineNumber:    line,
		Pattern:       pattern,
		ExceptionRule: "!" + baseName,
	}, nil
}

// AddExceptionRule appends the exception rule to the .gitignore that contains the matching rule
func (p *TrackProcessor) AddExceptionRule(sourceFile string, rule string) error {
	// sourceFile 来自 git check-ignore -v，是相对于 git root 的路径，需要转为绝对路径
	root, err := common.GetGitRoot()
	if err != nil {
		return err
	}
	absPath := filepath.Join(root, sourceFile)
	return appendRuleToGitignore(absPath, rule)
}

func appendRuleToGitignore(targetPath string, rule string) error {
	var content []byte

	if _, err := os.Stat(targetPath); err == nil {
		content, err = os.ReadFile(targetPath)
		if err != nil {
			return err
		}
	}

	if len(content) > 0 && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}

	content = append(content, []byte(rule+"\n")...)
	return os.WriteFile(targetPath, content, 0660)
}
