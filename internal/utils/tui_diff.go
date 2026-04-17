package utils

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// DiffViewer 创建一个TUI差异对比界面
type DiffViewer struct {
	app       *tview.Application
	confirmed bool
	cancelled bool
}

// NewDiffViewer 创建新的差异查看器
func NewDiffViewer() *DiffViewer {
	return &DiffViewer{
		app: tview.NewApplication(),
	}
}

// ShowDiff 显示差异对比界面，返回用户是否确认应用更改
func (dv *DiffViewer) ShowDiff(original, updated string) (bool, error) {
	// 创建左右两个文本视图
	leftView := tview.NewTextView()
	leftView.SetDynamicColors(true)
	leftView.SetTitle("原始版本")
	leftView.SetBorder(true)
	leftView.SetBorderPadding(0, 0, 1, 1)

	rightView := tview.NewTextView()
	rightView.SetDynamicColors(true)
	rightView.SetTitle("更新版本")
	rightView.SetBorder(true)
	rightView.SetBorderPadding(0, 0, 1, 1)

	// 按行级diff
	dmp := diffmatchpatch.New()
	a, b, lineArray := dmp.DiffLinesToChars(original, updated)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	var leftBuilder, rightBuilder strings.Builder
	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		for i, line := range lines {
			if i == len(lines)-1 && line == "" {
				continue
			}
			switch diff.Type {
			case diffmatchpatch.DiffInsert:
				rightBuilder.WriteString("[black:green]+ " + line + "[white:black]\n")
			case diffmatchpatch.DiffDelete:
				leftBuilder.WriteString("[black:red]- " + line + "[white:black]\n")
			case diffmatchpatch.DiffEqual:
				leftBuilder.WriteString("  " + line + "\n")
				rightBuilder.WriteString("  " + line + "\n")
			}
		}
	}
	leftView.SetText(leftBuilder.String())
	rightView.SetText(rightBuilder.String())

	// 创建帮助信息
	helpView := tview.NewTextView()
	helpView.SetDynamicColors(true)
	helpView.SetTitle("操作说明")
	helpView.SetBorder(true)
	helpView.SetBorderPadding(0, 0, 1, 1)
	helpView.SetText("[yellow]Tab/Shift+Tab: 切换视图  [green]Enter/y: 确认应用  [red]Esc/q/n: 取消")

	// 创建网格布局
	grid := tview.NewGrid().
		SetRows(0, 3).
		SetColumns(0, 0).
		SetBorders(false)

	// 添加主要内容到网格
	grid.AddItem(leftView, 0, 0, 1, 1, 0, 0, false).
		AddItem(rightView, 0, 1, 1, 1, 0, 0, false).
		AddItem(helpView, 1, 0, 1, 2, 0, 0, false)

	// 设置焦点循环
	focusables := []tview.Primitive{leftView, rightView}
	currentFocus := 0

	// 键盘事件处理
	dv.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			currentFocus = (currentFocus + 1) % len(focusables)
			dv.app.SetFocus(focusables[currentFocus])
			return nil
		case tcell.KeyBacktab:
			currentFocus = (currentFocus - 1 + len(focusables)) % len(focusables)
			dv.app.SetFocus(focusables[currentFocus])
			return nil
		case tcell.KeyEnter:
			dv.confirmed = true
			dv.app.Stop()
			return nil
		case tcell.KeyEscape:
			dv.cancelled = true
			dv.app.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'y', 'Y':
				dv.confirmed = true
				dv.app.Stop()
				return nil
			case 'n', 'N', 'q', 'Q':
				dv.cancelled = true
				dv.app.Stop()
				return nil
			}
		}
		return event
	})

	// 设置初始焦点
	dv.app.SetFocus(leftView)

	// 运行应用
	if err := dv.app.SetRoot(grid, true).Run(); err != nil {
		return false, fmt.Errorf("TUI应用运行失败: %v", err)
	}

	return dv.confirmed && !dv.cancelled, nil
}

// ShowGitignoreDiff 专门用于.gitignore文件差异显示的便捷方法
func ShowGitignoreDiff(original, updated string) (bool, error) {
	viewer := NewDiffViewer()
	return viewer.ShowDiff(original, updated)
}
