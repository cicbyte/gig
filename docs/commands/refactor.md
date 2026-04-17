# gig refactor

使用 AI 整理优化现有的 `.gitignore` 文件。重新分类、去重、优化格式，不新增或删除任何规则。

## 用法

```
gig refactor [flags]
```

## 示例

```bash
# 整理 .gitignore
gig refactor

# 预览整理结果
gig refactor --dry-run

# 自动确认并应用
gig refactor --yes
```

## 参数

| 参数 | 说明 |
|------|------|
| `--dry-run` | 仅显示整理结果，不实际修改 |
| `--yes` | 自动确认，无需人工输入 |

## 注意

- 需要配置 AI API Key，参见 `gig config`
- 整理规则：移除重复、按类别分组、优化格式
- 不会新增或删除任何规则
