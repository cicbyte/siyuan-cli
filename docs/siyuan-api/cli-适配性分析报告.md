# 思源笔记 API CLI 适配性分析报告

> 基于 `.reference/wiki/siyuan/reference.md` 中整理的 100+ API 端点，按 CLI 适配优先级分类。

## 当前已实现

| 命令组 | 已对接 API |
|:---|:---|
| `notebook` | list, create, rename, delete, open, close, getconf, setconf |
| `document` | list (listDocTree), get (exportMdContent), createMd, rename, delete |
| `auth` | 配置管理（非 API） |

---

## 高优先级 — 强烈推荐对接

| API | CLI 命令建议 | 理由 |
|:---|:---|:---|
| `/api/search/fullTextSearchBlock` | `search block <keyword>` | **全文搜索是 CLI 最核心的缺失功能**，用户输入关键词即可定位任意块 |
| `/api/filetree/searchDocs` | `search doc <keyword>` | 文档级搜索，与 block 搜索互补 |
| `/api/attr/setBlockAttrs` | `document set-tag <doc> --tag "标签"` | 通过属性系统管理标签/别名等元数据，CLI 批量操作优势明显 |
| `/api/attr/getBlockAttrs` | `document get --attrs` | 查看/管理文档自定义属性 |
| `/api/filetree/moveDocs` | `document move <src> <dest>` | 文档移动/重组，CLI 批量操作效率远超 GUI |
| `/api/filetree/duplicateDoc` | `document copy <doc>` | 文档复制，快速模板化 |
| `/api/outline/getDocOutline` | `document outline <doc>` | 大纲/T 目录查看，定位长文档结构 |
| `/api/query/sql` | `query "SELECT * FROM blocks WHERE ..."` | **杀手级功能**：直接 SQL 查询思源数据库，自动化场景无可替代 |
| `/api/filetree/createDailyNote` | `document daily` | 快速创建日记，CLI 用户高频操作 |
| `/api/asset/upload` | `asset upload <file>` | 上传图片/文件到思源 |
| `/api/history/getDocHistoryContent` | `document history <doc>` | 查看文档历史版本，CLI diff 更方便 |
| `/api/history/rollbackDocHistory` | `document rollback <doc> --to <rev>` | 回滚文档，CLI 独有优势 |

## 中优先级 — 值得对接

| API | CLI 命令建议 | 理由 |
|:---|:---|:---|
| `/api/block/getBlockInfo` | `block get <id>` | 查看块详细信息，调试/开发场景 |
| `/api/block/getBlockKramdown` | `block source <id>` | 获取块源码，开发者友好 |
| `/api/block/updateBlock` | `block update <id> --content "..."` | 编辑块内容，配合管道实现 |
| `/api/block/insertBlock` | `block append <doc> --content "..."` | 追加内容到文档 |
| `/api/block/deleteBlock` | `block delete <id>` | 删除指定块 |
| `/api/export/exportHTML` | `document export <doc> --format html` | 导出为 HTML，用于分享/发布 |
| `/api/export/exportDocx` | `document export <doc> --format docx` | 导出 Word，办公场景 |
| `/api/export/exportNotebookMd` | `notebook export <nb> --format md` | 整个笔记本导出为 Markdown，备份/迁移 |
| `/api/import/importStdMd` | `document import <file> --notebook <nb>` | 导入 Markdown 文件 |
| `/api/import/importSY` | `document import --format sy` | 导入思源格式，跨实例迁移 |
| `/api/asset/getUnusedAssets` | `asset unused` | 清理无用资源，磁盘管理 |
| `/api/asset/removeUnusedAssets` | `asset clean` | 一键清理 |
| `/api/search/searchTag` | `tag search <keyword>` | 标签搜索 |
| `/api/tag/getTag` | `tag list` | 查看所有标签 |
| `/api/sync/performSync` | `sync` | 触发同步，自动化场景 |
| `/api/sync/getSyncInfo` | `sync status` | 查看同步状态 |
| `/api/filetree/removeDocs` | `document delete --batch` | 批量删除，大清理场景 |
| `/api/attr/getBookmarkLabels` | `bookmark list` | 查看书签 |

## 低优先级 — 可选

| API | 说明 |
|:---|:---|
| `/api/ref/getBacklink` | 反向链接查询（需配合 block 系统使用） |
| `/api/graph/getLocalGraph` | 关系图（CLI 难以可视化） |
| `/api/repo/*` | 仓库快照管理，适合高级用户 |
| `/api/av/*` | 属性视图/数据库，CLI 表现力有限 |
| `/api/broadcast/*` | 广播/WebSocket，不适合 CLI |
| `/api/ui/*` | 界面操作，CLI 无意义 |
| `/api/notification/*` | 通知推送，CLI 无意义 |
| `/api/network/*` | 网络代理，低频需求 |
| `/api/template/*` | 模板渲染，可选扩展 |

---

## 建议实现路线图

### Phase 1 — 搜索与查询（最高价值）

- `search` 命令组（block 全文搜索 + 文档搜索）
- `query` 命令（SQL 直查，开发者杀手级功能）

### Phase 2 — 文档增强

- `document move` / `document copy`
- `document outline`
- `document daily`（日记）
- `document history` / `document rollback`

### Phase 3 — 元数据与属性

- `document set-tag` / `document get --attrs`
- `tag list` / `tag search`

### Phase 4 — 导入导出

- `document export --format html/docx`
- `notebook export`
- `document import`

### Phase 5 — 资源与同步

- `asset upload` / `asset list` / `asset clean`
- `sync` / `sync status`
