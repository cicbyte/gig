# gig add

向 `.gitignore` 文件添加模板规则。支持本地模板、GitHub 官方模板和 AI 生成三种来源。如果 `.gitignore` 不存在会自动创建。

## 用法

```
gig add [languages...] [flags]
```

## 示例

```bash
# 从本地模板添加（默认）
gig add go python node

# 从 GitHub 官方模板添加（首次自动克隆）
gig add Go Python -t github

# 用 AI 生成规则
gig add go,python -t ai

# 将 AI 生成的规则保存为本地模板
gig add rust -t ai --save

# 交互式向导
gig add -i

# 预览模式，不实际修改
gig add go python --dry-run
```

## 参数

| 参数 | 说明 |
|------|------|
| `languages...` | 语言名称，支持逗号分隔 |
| `-t, --type` | 模板源：`local`、`github`、`ai`（默认 `local`） |
| `-i, --interactive` | 进入交互式向导 |
| `-u, --update` | 强制更新 GitHub 模板 |
| `--save` | 将 AI 生成的规则保存为本地模板 |
| `--tui` | 使用 TUI 界面显示差异对比 |
| `--dry-run` | 仅预览，不实际修改 |
| `--yes` | 自动确认所有交互 |

## 注意

- 使用 `-t github` 时，如果本地未克隆 GitHub 模板仓库，会自动执行克隆
- AI 模式需要配置 API Key，参见 `gig config`
