# gig track

强制 Git 跟踪一个当前被忽略的文件。通过向匹配的 `.gitignore` 添加否定规则（`!`）实现。

自动定位到包含匹配忽略规则的 `.gitignore` 文件（可能是子目录的），在其中添加否定规则。

## 用法

```
gig track <file-path>... [flags]
```

## 示例

```bash
# 强制跟踪被忽略的文件
gig track important.log

# 自动确认
gig track debug.log --yes
```

## 参数

| 参数 | 说明 |
|------|------|
| `file-path...` | 要跟踪的文件路径，支持多个 |
| `--yes` | 自动确认所有交互 |

## 注意

- 否定规则会添加到包含匹配规则的 `.gitignore` 中，可能是子目录的
- 否定规则使用文件名（如 `!node_modules/`），而非完整路径
- 如果文件未被忽略或已被跟踪，会给出相应提示
