# gig completion

生成 Shell 自动补全脚本，支持 Bash、Zsh、Fish 和 PowerShell。

## 用法

```
gig completion [shell]
```

## 示例

```bash
# Bash
gig completion bash > /etc/bash_completion.d/gig

# Zsh
gig completion zsh > "${fpath[1]}/_gig"

# Fish
gig completion fish > ~/.config/fish/completions/gig.fish

# PowerShell
gig completion powershell | Out-String | Invoke-Expression
```

## 支持的 Shell

| Shell | 参数 |
|-------|------|
| Bash | `bash` |
| Zsh | `zsh` |
| Fish | `fish` |
| PowerShell | `powershell` |
