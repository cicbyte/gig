# gig

[![Go Reference](https://pkg.go.dev/badge/github.com/cicbyte/gig.svg)](https://pkg.go.dev/github.com/cicbyte/gig)
[![Release](https://img.shields.io/github/v/release/cicbyte/gig?style=flat-square)](https://github.com/cicbyte/gig/releases)
[![License](https://img.shields.io/github/license/cicbyte/gig?style=flat-square)](LICENSE)

**[English](README.en.md)** | 简体中文

> 智能 `.gitignore` 全生命周期管理工具

**gig** 是一个 Go CLI 工具，提供 `.gitignore` 文件的创建、诊断、整理、修复等全生命周期管理。支持本地模板、GitHub 官方模板和 AI 生成三种模板源。

## 功能特性

- **模板添加** (`gig add`) — 从本地、GitHub 或 AI 获取规则，合并去重后写入
- **直接忽略** (`gig ignore`) — 将文件/文件夹添加到 `.gitignore`，已跟踪的文件自动取消跟踪
- **健康诊断** (`gig doctor`) — 检测重复规则、危险模式、性能问题，支持一键修复
- **忽略检查** (`gig check`) — 检查文件是否被忽略，显示匹配规则和来源
- **AI 整理** (`gig refactor`) — 用 AI 重新分类、去重、优化 `.gitignore` 格式
- **强制跟踪** (`gig track`) — 向匹配的 `.gitignore` 添加否定规则，跟踪被忽略的文件
- **查看文件** (`gig view`) — 查看当前项目的 `.gitignore`，自动向上查找
- **模板管理** (`gig template local/github`) — 管理本地和 GitHub 模板

## 安装

需要 Go 1.23+。

```bash
go install github.com/cicbyte/gig@latest
```

也可从 [GitHub Releases](https://github.com/cicbyte/gig/releases) 下载预编译二进制文件。

## 快速开始

```bash
# 自动检测项目类型并添加本地模板
gig add

# 从 GitHub 官方仓库添加模板（首次自动克隆）
gig add Go Python -t github

# 用 AI 生成并添加规则
gig add go,python -t ai

# 直接忽略文件或文件夹
gig ignore build dist node_modules

# 检查文件是否被忽略
gig check .env

# 诊断 .gitignore 健康状态
gig doctor

# 一键修复可修复的问题
gig doctor --fix

# 用 AI 整理优化 .gitignore
gig refactor

# 强制跟踪被忽略的文件
gig track important.log

# 查看当前项目的 .gitignore
gig view
```

## 命令参考

### 项目操作

| 命令 | 说明 |
| :--- | :--- |
| `gig add [lang...]` | 添加模板规则到 `.gitignore` |
| `gig ignore <path>...` | 将文件/文件夹添加到 `.gitignore`（已跟踪的自动取消跟踪） |
| `gig check <file>` | 检查文件是否被 Git 忽略 |
| `gig view` | 查看当前项目的 `.gitignore` |
| `gig doctor` | 诊断 `.gitignore` 健康状态 |
| `gig refactor` | 用 AI 整理优化 `.gitignore` |
| `gig track <file>` | 强制跟踪被忽略的文件（添加否定规则） |

### 模板管理

| 命令 | 说明 |
| :--- | :--- |
| `gig template local list` | 列出所有本地模板 |
| `gig template local view -n <name>` | 查看指定模板内容 |
| `gig template local search <keyword>` | 搜索本地模板 |
| `gig template local add <file>` | 添加本地模板 |
| `gig template local copy -n <name>` | 复制 GitHub 模板到本地 |
| `gig template local remove -n <name>` | 删除本地模板 |
| `gig template local edit [-n <name>]` | 用编辑器打开模板目录或指定模板 |
| `gig template github sync` | 克隆/更新 GitHub 官方模板 |
| `gig template github reset` | 重置 GitHub 模板 |
| `gig template github list` | 列出 GitHub 模板 |
| `gig template github view -n <name>` | 查看 GitHub 模板内容 |
| `gig template github search <keyword>` | 搜索 GitHub 模板 |

### 配置与工具

| 命令 | 说明 |
| :--- | :--- |
| `gig config` | 显示当前配置 |
| `gig config set <key> <value>` | 设置配置项 |
| `gig config reset [key]` | 重置配置项 |
| `gig config edit` | 用编辑器打开配置文件 |
| `gig version` | 显示版本信息 |
| `gig completion` | 生成 Shell 补全脚本 |

### 全局参数

| 参数 | 说明 |
| :--- | :--- |
| `--config` | 指定配置文件路径 |
| `-t, --type` | 模板源：`local`、`github`、`ai`（默认 `local`） |
| `--yes` | 跳过确认提示 |

## 配置

配置目录：`~/.cicbyte/gig/config/`

| 文件 | 说明 |
| :--- | :--- |
| `config.yaml` | AI 配置（API Key、URL、Model） |
| `prompts/*.md` | AI 提示词模板，每类一个 `.md` 文件，可自定义 |
| `detection.json` | 文件标记到项目类型的映射（如 `go.mod` → `go`） |

### AI 配置

```bash
# 交互式配置
gig config

# 直接设置
gig config set ai.api_key sk-xxx
gig config set ai.url https://api.deepseek.com/chat/completions
gig config set ai.model deepseek-chat
```

也支持环境变量覆盖：

| 环境变量 | 对应配置项 |
| :--- | :--- |
| `GIG_AI_API_KEY` | `ai.api_key` |
| `GIG_AI_URL` | `ai.url` |
| `GIG_AI_MODEL` | `ai.model` |

默认使用 DeepSeek API（`deepseek-chat`），兼容所有 OpenAI 接口格式。为了获得更好的效果，建议选择能力较强的模型（如 `deepseek-reasoner`、`claude-sonnet-4-20250514` 等）。

## 数据目录结构

```
~/.cicbyte/gig/
├── config/
│   ├── config.yaml          # AI 配置
│   ├── detection.json       # 项目类型检测规则
│   └── prompts/             # AI 提示词模板
│       ├── add.md
│       ├── refactor.md
│       └── ...
├── template/                # 本地用户模板
│   ├── Go.gitignore
│   ├── Python.gitignore
│   └── ...
└── template_github/         # GitHub 官方模板（git clone）
```

## 多 .gitignore 支持

gig 会智能处理项目中存在多个 `.gitignore` 的场景：

- **`gig ignore`** — 写入当前目录最近的 `.gitignore`（从当前目录向上查找）
- **`gig check`** — 显示所有匹配的规则及来源文件，按优先级排序
- **`gig track`** — 定位到包含匹配规则的 `.gitignore` 并写入否定规则
- **`gig view`** — 从当前目录向上查找最近的 `.gitignore`

优先级：子目录 `.gitignore` > 父目录 `.gitignore` > 根目录 `.gitignore`

## 技术栈

- Go 1.23
- [Cobra](https://github.com/spf13/cobra) — CLI 框架
- [Viper](https://github.com/spf13/viper) — 配置管理
- [go-diff](https://github.com/sergi/go-diff) — Diff 计算
- [promptui](https://github.com/manifoldco/promptui) — 交互式提示
- [Zap](https://github.com/uber-go/zap) — 结构化日志

## License

[MIT](LICENSE)
