# gig ignore

将文件或文件夹添加到 `.gitignore`，同时智能处理 Git 跟踪状态。

## 用法

```
gig ignore <path>... [flags]
```

## 示例

```bash
# 忽略目录
gig ignore build dist node_modules

# 忽略已跟踪的文件（自动执行 git rm --cached）
gig ignore .env

# 忽略通配符文件
gig ignore *.log

# 原样写入，不做路径检测
gig ignore --raw "!keep.me"

# 自动确认所有操作
gig ignore build dist --yes
```

## 参数

| 参数 | 说明 |
|------|------|
| `path...` | 文件或文件夹路径，支持多个 |
| `--raw` | 原样写入，不做路径检测和格式化 |
| `--yes` | 自动确认所有交互（如取消跟踪确认） |

## 智能行为

- **文件未被跟踪** → 直接添加忽略规则
- **文件已被跟踪** → 添加忽略规则后，询问是否执行 `git rm --cached` 取消跟踪
- **规则已存在** → 跳过，但如果文件仍被跟踪则询问是否取消跟踪

## 注意

- 目录会自动追加 `/`（如 `build` → `build/`）
- 含通配符的模式原样添加（如 `*.log`）
- 使用 `--raw` 可写入任意模式，包括否定规则
