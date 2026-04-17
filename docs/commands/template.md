# gig template

管理 `.gitignore` 模板，分为本地模板和 GitHub 模板两个子组。

## 用法

```
gig template <subcommand> [flags]
```

## 子命令

### local — 本地模板管理

```bash
# 列出所有本地模板
gig template local list

# 查看指定模板内容
gig template local view -n Go

# 搜索本地模板
gig template local search py

# 添加本地模板文件
gig template local add ./my-template.gitignore

# 复制 GitHub 模板到本地，便于二次编辑
gig template local copy -n Go

# 删除本地模板
gig template local remove -n Go

# 用编辑器打开模板目录
gig template local edit

# 编辑指定模板
gig template local edit -n Go
```

| 参数 | 说明 |
|------|------|
| `-n` | 指定模板名称（大部分命令必需） |

### github — GitHub 官方模板管理

```bash
# 克隆/更新 GitHub 官方模板（首次自动克隆）
gig template github sync

# 重置 GitHub 模板（删除并重新克隆）
gig template github reset

# 列出 GitHub 模板
gig template github list

# 查看 GitHub 模板内容
gig template github view -n Go

# 搜索 GitHub 模板
gig template github search py
```

## 本地模板目录

`~/.cicbyte/gig/template/`

## GitHub 模板目录

`~/.cicbyte/gig/template_github/`（从 github/gitignore 仓库克隆）
