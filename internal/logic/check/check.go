package check

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cicbyte/gig/internal/models"
)

type CheckConfig struct {
	FilePath string
}

type CheckMatch struct {
	SourceFile string
	LineNumber string
	Pattern    string
}

type CheckResult struct {
	Ignored bool
	FilePath string
	Matches []CheckMatch // 所有匹配的规则（按优先级排序）
}

type CheckProcessor struct {
	config    *CheckConfig
	appConfig *models.Config
}

func NewCheckProcessor(config *CheckConfig, appConfig *models.Config) *CheckProcessor {
	return &CheckProcessor{config: config, appConfig: appConfig}
}

func (p *CheckProcessor) Execute(ctx context.Context) (*CheckResult, error) {
	filePath := strings.TrimRight(p.config.FilePath, string(filepath.Separator))

	cmd := exec.Command("git", "check-ignore", "-v", filePath)
	output, err := cmd.Output()

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return &CheckResult{Ignored: false, FilePath: filePath}, nil
		}
		return nil, fmt.Errorf("运行 git 时发生错误：%w", err)
	}

	matches, err := parseCheckIgnoreOutput(string(output))
	if err != nil {
		return nil, err
	}

	return &CheckResult{
		Ignored: true,
		FilePath: filePath,
		Matches: matches,
	}, nil
}

func parseCheckIgnoreOutput(output string) ([]CheckMatch, error) {
	var matches []CheckMatch
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		details := strings.SplitN(parts[0], ":", 3)
		if len(details) < 3 {
			continue
		}
		matches = append(matches, CheckMatch{
			SourceFile: details[0],
			LineNumber: details[1],
			Pattern:    details[2],
		})
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("无法解析 git check-ignore 的输出")
	}
	return matches, nil
}
