# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

思源笔记（SiYuan Note）命令行管理工具，Go 语言编写。通过思源内核 API 实现笔记本管理、文档操作、块编辑等功能，并提供交互式 TUI 界面。

## 构建与运行

```bash
# 开发构建
go build -o siyuan-cli .

# 直接运行
go run cmd/*.go

# 生产构建（含 Web 前端构建 + UPX 压缩）
python build.py
```

`build.py` 流程：`npm build(web/dist)` → 清理 `resources/static` → 复制前端产物 → `go build` → UPX 压缩。构建信息（版本号、commit hash、分支）通过 `-ldflags` 注入 `main` 包变量。

当前无单元测试框架。

## 架构

分层结构，职责清晰：

- **`cmd/`** — Cobra 命令定义，每个子命令一个文件（root.go、notebook.go、document.go、fav.go）
- **`internal/logic/`** — 业务逻辑层，按领域分包（notebook/、document/、fav/）
- **`internal/siyuan/`** — 思源内核 API 的 Go SDK 封装。`client.go` 定义 `Client` 结构体和 `post`/`postMultipart` 基础请求方法；`factory.go` 从全局配置创建客户端实例
- **`internal/ui/`** — Bubble Tea TUI 组件（笔记本列表、文档树、路径选择器等）
- **`internal/models/`** — 数据结构定义，核心是 `AppConfig`（YAML 配置模型）
- **`internal/utils/config.go`** — 配置管理单例，处理目录创建、配置加载/保存、默认值
- **`internal/common/`** — 全局变量（`AppConfigModel`）
- **`internal/log/`** — Zap 日志初始化

## 关键设计

**配置系统**：YAML 配置文件位于 `~/.cicbyte/siyuan-cli/config/config.yaml`，首次运行自动创建。`common.AppConfigModel` 是全局配置入口，在 `cmd/root.go` 的 `init()` 中加载。

**SiYuan SDK**：`Client` 线程安全、可复用。所有 API 均为 POST 请求，返回 `{code, msg, data}` 结构，`code != 0` 时返回 `*SiYuanError`。通过 `factory.go` 的 `GetDefaultSiYuanClient()` 获取客户端。

**启动流程**（`cmd/root.go` init）：初始化应用目录 → 加载配置 → 初始化日志 → 连接 SQLite 数据库。

## 依赖

- **Cobra** — CLI 框架
- **Bubble Tea + Lipgloss + Bubbles** — TUI
- **Zap + Lumberjack** — 日志（按天轮转）
- **GORM + SQLite** — 数据持久化
- **go.yaml.in/yaml/v3** — 配置解析
