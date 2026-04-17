# gig doctor

诊断 `.gitignore` 文件的健康状态，检测重复规则、危险模式、性能问题和缺失项。支持一键修复。

## 用法

```
gig doctor [flags]
```

## 示例

```bash
# 诊断当前项目
gig doctor

# 一键修复所有可修复的问题
gig doctor --fix

# 预览修复内容
gig doctor --fix --dry-run

# 自动确认修复
gig doctor --fix --yes
```

## 参数

| 参数 | 说明 |
|------|------|
| `--fix` | 自动修复可修复的问题 |
| `--dry-run` | 仅展示修复内容，不实际修改 |
| `--yes` | 自动确认修复，无需人工输入 |
| `-t, --type` | 诊断模式：`local` 或 `ai`（默认 `local`） |

## 注意

- 使用 `-t ai` 可获取额外的 AI 诊断建议，需要配置 API Key
- 使用 `--dry-run` 可预览修复内容
