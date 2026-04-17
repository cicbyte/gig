# gig config

管理 gig 的应用程序配置。支持查看、设置、编辑和重置配置项。

## 用法

```bash
gig config [key] [value]
gig config <subcommand>
```

## 示例

```bash
# 查看所有配置
gig config
gig config list

# 设置配置项
gig config set ai.api_key sk-xxx
gig config set ai.model deepseek-chat

# 用编辑器打开配置文件
gig config edit

# 重置单个配置项
gig config reset ai.api_key

# 重置所有配置项
gig config reset
```

## 子命令

| 子命令 | 说明 |
|--------|------|
| `list` | 以表格形式展示所有配置项 |
| `set <key> <value>` | 设置配置项的值 |
| `edit` | 用系统编辑器打开配置文件 |
| `reset [key]` | 重置配置项为默认值 |

## 可用配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `ai.api_key` | AI API 密钥 | 空 |
| `ai.url` | AI API 地址 | `https://api.deepseek.com/chat/completions` |
| `ai.model` | AI 模型名称 | `deepseek-chat` |

## 环境变量

所有配置项支持环境变量覆盖（前缀 `GIG_`，用 `_` 替代 `.`）：

| 环境变量 | 对应配置项 |
|----------|-----------|
| `GIG_AI_API_KEY` | `ai.api_key` |
| `GIG_AI_URL` | `ai.url` |
| `GIG_AI_MODEL` | `ai.model` |

## 配置文件位置

`~/.cicbyte/gig/config/config.yaml`
