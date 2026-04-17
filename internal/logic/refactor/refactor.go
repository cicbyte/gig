package refactor

import (
	"context"
	"fmt"
	"os"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/defaults"
	"github.com/cicbyte/gig/internal/models"
	"github.com/cicbyte/gig/internal/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type RefactorResult struct {
	OriginalContent   string
	RefactoredContent string
	DiffText          string
}

type RefactorProcessor struct {
	appConfig *models.Config
}

func NewRefactorProcessor(appConfig *models.Config) *RefactorProcessor {
	return &RefactorProcessor{appConfig: appConfig}
}

func (p *RefactorProcessor) Execute(ctx context.Context) (*RefactorResult, error) {
	gitignorePath, err := common.GetGitignorePath()
	if err != nil {
		return nil, err
	}

	originalContent, err := utils.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("当前目录中未找到 .gitignore 文件")
		}
		return nil, fmt.Errorf("读取 .gitignore 文件时出错：%w", err)
	}

	refactoredContent, err := utils.AI.StreamChat("refactor", originalContent)
	if err != nil {
		return nil, fmt.Errorf("AI 整理失败：%w", err)
	}
	refactoredContent = defaults.CleanupAIOutput(refactoredContent)

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(originalContent, refactoredContent, false)
	diffText := dmp.DiffPrettyText(diffs)

	return &RefactorResult{
		OriginalContent:   originalContent,
		RefactoredContent: refactoredContent,
		DiffText:          diffText,
	}, nil
}

func (p *RefactorProcessor) Apply(content string) error {
	gitignorePath, err := common.GetGitignorePath()
	if err != nil {
		return err
	}

	return utils.WriteFile(gitignorePath, content)
}
