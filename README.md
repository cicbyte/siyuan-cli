# siyuan-cli

> 思源笔记命令行工具 — 笔记本/文档/块管理、全文搜索、SQL 查询、AI 对话、MCP 集成，全部在终端完成。

简体中文 | [English](README.en.md)

![Release](https://img.shields.io/github/v/release/cicbyte/siyuan-cli?style=flat)
![Go Report Card](https://goreportcard.com/badge/github.com/cicbyte/siyuan-cli)
![License](https://img.shields.io/github/license/cicbyte/siyuan-cli)
![Last Commit](https://img.shields.io/github/last-commit/cicbyte/siyuan-cli)

## 功能特性

- **笔记本管理** — 创建、打开、关闭、重命名、删除，支持配置读写
- **文档操作** — 列表、读取、创建、重命名、删除、移动、复制、大纲、日记、历史、回滚
- **块操作** — 获取块信息、获取 kramdown 源码、更新块内容、追加块、删除块
- **全文搜索** — 按文档标题搜索、按块内容搜索
- **SQL 查询** — 直接查询思源 blocks 表，灵活检索
- **标签管理** — 列出、搜索、添加、移除标签
- **导入导出** — Markdown / .sy 格式导入，HTML / DOCX / Markdown / 思源格式 导出
- **资源管理** — 上传、查看文档资源、清理未使用资源
- **同步管理** — 查看同步状态、立即同步
- **AI 对话** — 基于 tool calling 的 Agent，LLM 自动检索笔记内容并生成回答
- **MCP Server** — stdio 模式，让 Claude Desktop、Cherry Studio 等 AI 客户端直接操作笔记
- **多格式输出** — table / JSON，`--format` 全局切换
- **多 AI 提供商** — OpenAI、Ollama、智谱等 OpenAI 兼容 API

## 安装

```bash
git clone https://github.com/cicbyte/siyuan-cli.git
cd siyuan-cli
go build -o siyuan-cli .
```

**环境要求：** Go >= 1.24、思源笔记实例（本地或远程）

## 快速开始

```bash
siyuan-cli auth login                          # 配置连接
siyuan-cli notebook list                        # 列出笔记本
siyuan-cli search block "关键词"                 # 全文搜索
siyuan-cli chat "搜索关于 Go 的笔记"            # AI 对话
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `auth login` / `logout` / `status` | 连接管理 |
| `config list` / `get` / `set` | 配置管理 |
| `notebook list` / `create` / `rename` / `delete` / `open` / `close` / `getconf` / `setconf` | 笔记本管理 |
| `document list` / `get` / `createMd` / `rename` / `delete` / `move` / `copy` / `outline` / `daily` / `history` / `rollback` | 文档管理 |
| `block get` / `source` / `update` / `append` / `delete` | 块操作 |
| `tag list` / `search` / `add` / `remove` | 标签管理 |
| `search doc` / `search block` | 搜索 |
| `query <sql>` | SQL 查询 |
| `export doc` / `export notebook` | 导出 |
| `import md` / `import sy` | 导入 |
| `asset upload` / `list` / `unused` / `clean` | 资源管理 |
| `sync status` / `now` | 同步管理 |
| `fav [content]` | 收藏 |
| `chat [question]` | AI 对话 |
| `mcp` | 启动 MCP Server |
| `version` | 版本信息 |

### 笔记本

```bash
siyuan-cli notebook list                        # 列出笔记本
siyuan-cli notebook create "我的笔记"            # 创建笔记本
siyuan-cli notebook rename "旧名" "新名"         # 重命名
siyuan-cli notebook delete "旧笔记" -F           # 删除（-F 强制）
siyuan-cli notebook open "笔记本"               # 打开
siyuan-cli notebook close "笔记本"              # 关闭
siyuan-cli notebook getconf "笔记本"            # 获取配置
siyuan-cli notebook setconf "笔记本" --sort 1    # 设置配置
```

### 文档

```bash
siyuan-cli document list "我的笔记"              # 列出文档
siyuan-cli document list "我的笔记" --path /技术   # 指定路径
siyuan-cli document get "我的笔记" /日记/2024-01-01  # 获取内容
siyuan-cli document createMd "我的笔记" --title "新文档" --content "# 标题\n内容"
siyuan-cli document outline "我的笔记" /技术笔记  # 查看大纲
siyuan-cli document daily --notebook 日记本      # 创建今日日记
siyuan-cli document move "笔记" /旧路径 /新路径    # 移动文档
siyuan-cli document copy "笔记" /文档路径         # 复制文档
siyuan-cli document history "笔记" /文档 --limit 5  # 查看历史
siyuan-cli document rollback "笔记" /文档 --to <history-path>  # 回滚文档
```

### 搜索与查询

```bash
siyuan-cli search doc "AI"                       # 按标题搜索文档
siyuan-cli search block "goroutine"               # 全文搜索块内容
siyuan-cli query "SELECT * FROM blocks WHERE type='d' LIMIT 10"
```

### 块 / 标签

```bash
siyuan-cli block get 20240101120000-xxx           # 获取块信息
siyuan-cli block source 20240101120000-xxx        # 获取块 kramdown 源码
siyuan-cli block update <id> --content "新内容"    # 更新块
siyuan-cli block append <doc-id> --content "内容"  # 追加内容到文档
siyuan-cli block delete <id>                      # 删除块
siyuan-cli tag list                               # 列出标签
siyuan-cli tag search "Go"                        # 搜索标签
siyuan-cli tag add "笔记" "文档路径" --tag "标签"  # 添加标签
siyuan-cli tag remove "笔记" "文档路径" --tag "标签"  # 移除标签
```

### 导入导出

```bash
siyuan-cli export doc "笔记" /文档 --format html -o output.html
siyuan-cli export notebook "笔记" --format md -o ./backup/
siyuan-cli import md ./notes/ --notebook "笔记"
siyuan-cli import sy ./backup/ --notebook "笔记"
```

### 资源与同步

```bash
siyuan-cli asset upload ./image.png
siyuan-cli asset list "笔记" /文档路径
siyuan-cli asset unused                          # 列出未使用资源
siyuan-cli asset clean -F                        # 清理未使用资源
siyuan-cli sync status                           # 同步状态
siyuan-cli sync now                              # 立即同步
```

### 全局选项

```bash
siyuan-cli notebook list --format json           # JSON 输出
```

## AI 对话

```bash
siyuan-cli chat "问题"                            # 单轮对话
siyuan-cli chat                                  # 多轮交互对话
siyuan-cli chat --non-stream "问题"              # 非流式输出
```

AI Agent 通过 10 个 function tools 检索笔记内容（列出笔记本、文档树、获取文档、大纲、全文搜索、文档搜索、标签、SQL 查询、创建文档、创建日记），自动选择检索策略后生成回答。

交互模式支持 `/quit` `/exit` `/q` 退出、`/clear` 清除上下文。

## 配置

配置文件：`~/.cicbyte/siyuan-cli/config/config.yaml`（首次运行自动创建）

```yaml
siyuan:
  base_url: "http://127.0.0.1:6806"
  api_token: ""
  timeout: 30
  retry_count: 3
  enabled: false

ai:
  provider: openai
  base_url: "https://open.bigmodel.cn/api/paas/v4/"
  api_key: ""
  model: "GLM-4-Flash-250414"
  max_tokens: 2048
  temperature: 0.8
  timeout: 30

output:
  format: table
```

```bash
siyuan-cli config set ai.api_key "your-api-key"
siyuan-cli config set ai.model "GLM-4-Flash-250414"
siyuan-cli config list
```

## MCP Server

`siyuan-cli mcp` 以 stdio 模式运行 MCP Server，注册 14 个工具：

| Tool | 描述 |
|------|------|
| `notebook_list` | 列出所有笔记本 |
| `document_list` | 列出文档树 |
| `document_get` | 获取文档内容 |
| `document_outline` | 获取文档大纲 |
| `block_get` | 获取块信息 |
| `block_get_kramdown` | 获取块 kramdown 源码 |
| `search_fulltext` | 全文搜索 |
| `search_docs` | 搜索文档 |
| `tag_list` | 列出标签 |
| `query_sql` | 执行 SQL 查询 |
| `document_create` | 创建文档 |
| `daily_note_create` | 创建日记 |
| `block_update` | 更新块内容 |
| `block_append` | 追加块到文档 |

**Claude Desktop：**

```json
{
  "mcpServers": {
    "siyuan": {
      "command": "siyuan-cli",
      "args": ["mcp"]
    }
  }
}
```

**Cherry Studio：** 设置 → MCP 服务器，命令 `siyuan-cli`，参数 `mcp`

## 技术栈

- Go 1.24
- [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Bubbles](https://github.com/charmbracelet/bubbles) + [Lipgloss](https://github.com/charmbracelet/lipgloss) + [Glamour](https://github.com/charmbracelet/glamour) — TUI
- [Cobra](https://github.com/spf13/cobra) — CLI 框架
- [mcp-go](https://github.com/mark3labs/mcp-go) — MCP Server
- [go-openai](https://github.com/sashabaranov/go-openai) — OpenAI 兼容 API
- [go-pretty](https://github.com/jedib0t/go-pretty) — 终端表格
- [Zap](https://github.com/uber-go/zap) — 日志
- [GORM](https://gorm.io/) + SQLite — 数据持久化

## 许可证

[MIT](LICENSE) © 2025 cicbyte
