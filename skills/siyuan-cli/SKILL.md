---
name: siyuan-cli
description: 操作思源笔记的 CLI 工具。当用户要求管理笔记本、文档、块、标签、搜索笔记、导入导出、管理资源、同步数据时使用此 skill。可执行路径下的 siyuan-cli.exe（或 go run .）。不包含 auth/chat/mcp 命令。
---

# siyuan-cli 使用指南

可执行文件：项目根目录 `siyuan-cli.exe`（开发时用 `go run .`）
全局选项：`--format json` 输出 JSON，`--output/-o <file>` 导出到文件

## 前置条件

思源笔记实例需已运行且已通过 `siyuan-cli auth login` 配置连接。如连接失败，提示用户先执行 `siyuan-cli auth login`。

## 笔记本

```bash
siyuan-cli notebook list                              # 列出所有笔记本
siyuan-cli notebook list --closed                     # 包含已关闭
siyuan-cli notebook create "名称"                      # 创建
siyuan-cli notebook rename <nb> "新名"                  # 重命名
siyuan-cli notebook open <nb>                          # 打开
siyuan-cli notebook close <nb>                         # 关闭
siyuan-cli notebook delete <nb> -F                     # 删除（-F 强制）
siyuan-cli notebook getconf <nb>                       # 获取配置
siyuan-cli notebook setconf <nb> --sort 1              # 设置配置
```

`<nb>` 支持笔记本名称（模糊匹配）或笔记本 ID。

## 文档

```bash
siyuan-cli document list <nb>                          # 列出文档（depth=1）
siyuan-cli document list <nb> --depth 0                # 全部展开
siyuan-cli document list <nb> --path "/技术"           # 指定路径
siyuan-cli document get <nb> <doc-path>                # 获取 Markdown 内容
siyuan-cli document get <nb> <doc-path> -o output.md   # 导出到文件
siyuan-cli document outline <doc-id>                   # 大纲（支持 doc-id 或 nb/path）
siyuan-cli document createMd <nb> --title "标题" --content "# 标题\n内容"
siyuan-cli document createMd <nb> --file readme.md --path "/目录"
siyuan-cli document rename <nb> <path> "新名"
siyuan-cli document move <nb> <src> <dest>             # 支持跨笔记本：nb2:/path
siyuan-cli document copy <nb> <doc-path>
siyuan-cli document delete <nb> <doc-path>
siyuan-cli document daily --notebook <nb>              # 创建今日日记
siyuan-cli document history <nb> <path>                # 查看历史
siyuan-cli document history --query "关键词"            # 搜索历史
siyuan-cli document rollback --notebook <nb> --to <history-path>  # 回滚
```

## 搜索与查询

```bash
siyuan-cli search doc "关键词"                          # 按标题搜索
siyuan-cli search block "关键词"                        # 全文搜索块内容
siyuan-cli search block "关键词" --notebook <nb> -l 50  # 限定范围和数量
siyuan-cli query "SELECT id, content FROM blocks WHERE content LIKE '%关键词%'" -o result.json
```

SQL 仅限 SELECT。blocks 表常用字段：`id, type, content, hpath, box, created, updated`。

## 块

```bash
siyuan-cli block get <block-id>
siyuan-cli block source <block-id>                     # kramdown 源码
siyuan-cli block update <block-id> --content "新内容"
siyuan-cli block append <doc-id> --content "## 新段落"   # 也支持 nb/path
siyuan-cli block delete <block-id> --force
```

`block append` 的 `--type` 参数：`markdown`（默认）或 `dom`。

## 标签

```bash
siyuan-cli tag list
siyuan-cli tag search "关键词"
siyuan-cli tag add <doc-id> --tag "标签1" --tag "标签2"  # 支持 doc-id 或 nb/path
siyuan-cli tag remove <doc-id> --tag "标签"
```

## 导入导出

```bash
siyuan-cli export doc <doc-id|nb> [path] --format html -o output.html
siyuan-cli export doc <doc-id|nb> [path] --format md
siyuan-cli export notebook <nb> --format md -o ./backup/
siyuan-cli export notebook <nb> --format sy -o ./backup/
siyuan-cli import md <file-or-dir> --notebook <nb> --path "目标路径"
siyuan-cli import sy <file-or-dir> --notebook <nb>
```

## 资源与同步

```bash
siyuan-cli asset upload <file>
siyuan-cli asset list <doc-id|nb> [path]
siyuan-cli asset unused                                # 未使用资源
siyuan-cli asset clean -F                              # 清理未使用资源
siyuan-cli sync status
siyuan-cli sync now
```

## 其他

```bash
siyuan-cli fav "收藏内容"                              # 收藏到"我的收藏"笔记本
siyuan-cli config list                                 # 查看配置
siyuan-cli config get <key>                            # 获取配置项
siyuan-cli config set <key> <value>                    # 设置配置项
siyuan-cli version                                     # 版本信息
```

## 注意事项

- 笔记本/文档路径支持名称模糊匹配和 ID 精确匹配
- 文档路径用人类可读格式（如 `java面试/AI/index`），不用 `.sy` 后缀
- `--output/-o` 自动使用 JSON 格式
- `--format json` 为全局选项，放在命令前
- 删除操作建议加 `-F/--force` 避免交互确认阻塞
- `document history --query` 可搜索全局历史，不限定笔记本
