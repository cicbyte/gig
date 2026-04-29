# gig

[![Go Reference](https://pkg.go.dev/badge/github.com/cicbyte/gig.svg)](https://pkg.go.dev/github.com/cicbyte/gig)
[![Release](https://img.shields.io/github/v/release/cicbyte/gig?style=flat-square)](https://github.com/cicbyte/gig/releases)
[![License](https://img.shields.io/github/license/cicbyte/gig?style=flat-square)](LICENSE)

> Smart full-lifecycle `.gitignore` management tool

**gig** is a Go CLI tool that provides full-lifecycle management for `.gitignore` files — creation, diagnosis, refactoring, and repair. It supports three template sources: local templates, GitHub official templates, and AI generation.

## Features

- **Template Add** (`gig add`) — Fetch rules from local, GitHub, or AI sources, merge and deduplicate before writing
- **Ignore** (`gig ignore`) — Add files/folders to `.gitignore`; auto-untrack already-tracked files
- **Health Check** (`gig doctor`) — Detect duplicate rules, dangerous patterns, and performance issues with one-click fix
- **Ignore Check** (`gig check`) — Check if a file is ignored, showing matching rules and sources
- **AI Refactor** (`gig refactor`) — Use AI to reclassify, deduplicate, and optimize `.gitignore` formatting
- **Force Track** (`gig track`) — Add negation rules to `.gitignore` to track otherwise-ignored files
- **View** (`gig view`) — View the current project's `.gitignore`, auto-searching up the directory tree
- **Template Management** (`gig template local/github`) — Manage local and GitHub templates

## Installation

Requires Go 1.23+.

```bash
go install github.com/cicbyte/gig@latest
```

Pre-built binaries are available on [GitHub Releases](https://github.com/cicbyte/gig/releases).

## Quick Start

```bash
# Auto-detect project type and add local templates
gig add

# Add templates from GitHub official repo (auto-clone on first use)
gig add Go Python -t github

# Generate and add rules with AI
gig add go,python -t ai

# Ignore files or folders directly
gig ignore build dist node_modules

# Check if a file is ignored
gig check .env

# Diagnose .gitignore health
gig doctor

# One-click fix for repairable issues
gig doctor --fix

# Refactor and optimize .gitignore with AI
gig refactor

# Force-track an ignored file
gig track important.log

# View current project's .gitignore
gig view
```

## Command Reference

### Project Operations

| Command | Description |
| :--- | :--- |
| `gig add [lang...]` | Add template rules to `.gitignore` |
| `gig ignore <path>...` | Add files/folders to `.gitignore` (auto-untrack if tracked) |
| `gig check <file>` | Check if a file is ignored by Git |
| `gig view` | View current project's `.gitignore` |
| `gig doctor` | Diagnose `.gitignore` health |
| `gig refactor` | Refactor and optimize `.gitignore` with AI |
| `gig track <file>` | Force-track an ignored file (add negation rule) |

### Template Management

| Command | Description |
| :--- | :--- |
| `gig template local list` | List all local templates |
| `gig template local view -n <name>` | View a specific template |
| `gig template local search <keyword>` | Search local templates |
| `gig template local add <file>` | Add a local template |
| `gig template local copy -n <name>` | Copy a GitHub template to local |
| `gig template local remove -n <name>` | Delete a local template |
| `gig template local edit [-n <name>]` | Open template directory or specific template in editor |
| `gig template github sync` | Clone/update GitHub official templates |
| `gig template github reset` | Reset GitHub templates |
| `gig template github list` | List GitHub templates |
| `gig template github view -n <name>` | View a GitHub template |
| `gig template github search <keyword>` | Search GitHub templates |

### Configuration & Utilities

| Command | Description |
| :--- | :--- |
| `gig config` | Show current configuration |
| `gig config set <key> <value>` | Set a configuration item |
| `gig config reset [key]` | Reset a configuration item |
| `gig config edit` | Open config file in editor |
| `gig version` | Show version information |
| `gig completion` | Generate shell completion scripts |

### Global Flags

| Flag | Description |
| :--- | :--- |
| `--config` | Specify config file path |
| `-t, --type` | Template source: `local`, `github`, `ai` (default `local`) |
| `--yes` | Skip confirmation prompts |

## Configuration

Config directory: `~/.cicbyte/gig/config/`

| File | Description |
| :--- | :--- |
| `config.yaml` | AI configuration (API Key, URL, Model) |
| `prompts/*.md` | AI prompt templates, one `.md` file per type, customizable |
| `detection.json` | Filesystem marker to project type mapping (e.g., `go.mod` → `go`) |

### AI Configuration

```bash
# Interactive configuration
gig config

# Set directly
gig config set ai.api_key sk-xxx
gig config set ai.url https://api.deepseek.com/chat/completions
gig config set ai.model deepseek-chat
```

Environment variable overrides:

| Environment Variable | Config Key |
| :--- | :--- |
| `GIG_AI_API_KEY` | `ai.api_key` |
| `GIG_AI_URL` | `ai.url` |
| `GIG_AI_MODEL` | `ai.model` |

Defaults to DeepSeek API (`deepseek-chat`), compatible with all OpenAI-format APIs. For better results, consider using a more capable model (e.g., `deepseek-reasoner`, `claude-sonnet-4-20250514`).

## Data Directory Structure

```
~/.cicbyte/gig/
├── config/
│   ├── config.yaml          # AI configuration
│   ├── detection.json       # Project type detection rules
│   └── prompts/             # AI prompt templates
│       ├── add.md
│       ├── refactor.md
│       └── ...
├── template/                # Local user templates
│   ├── Go.gitignore
│   ├── Python.gitignore
│   └── ...
└── template_github/         # GitHub official templates (git clone)
```

## Multi .gitignore Support

gig intelligently handles projects with multiple `.gitignore` files:

- **`gig ignore`** — Writes to the nearest `.gitignore` (searches up from current directory)
- **`gig check`** — Shows all matching rules and source files, sorted by priority
- **`gig track`** — Locates the `.gitignore` containing the matching rule and writes negation rules
- **`gig view`** — Searches up from the current directory for the nearest `.gitignore`

Priority: subdirectory `.gitignore` > parent directory `.gitignore` > root `.gitignore`

## Tech Stack

- Go 1.23
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Viper](https://github.com/spf13/viper) — Configuration management
- [go-diff](https://github.com/sergi/go-diff) — Diff computation
- [promptui](https://github.com/manifoldco/promptui) — Interactive prompts
- [Zap](https://github.com/uber-go/zap) — Structured logging

## License

[MIT](LICENSE)
