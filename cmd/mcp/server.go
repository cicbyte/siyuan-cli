package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cicbyte/siyuan-cli/cmd/version"
	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func runMCPServer() error {
	if !siyuan.IsSiYuanConfigValid() {
		return fmt.Errorf("思源笔记配置无效或未启用，请运行 siyuan-cli auth login 配置连接")
	}

	s := server.NewMCPServer(
		"siyuan-cli",
		version.Version,
		server.WithToolCapabilities(true),
	)

	// ── 只读工具 ──

	s.AddTool(mcp.NewTool("notebook_list",
		mcp.WithDescription("列出所有思源笔记笔记本，返回 ID、名称、图标和开关状态"),
	), handleNotebookList)

	s.AddTool(mcp.NewTool("document_list",
		mcp.WithDescription("列出指定笔记本下的文档树结构"),
		mcp.WithString("notebook", mcp.Required(), mcp.Description("笔记本名称或 ID")),
		mcp.WithString("path", mcp.Description("文档路径，默认为根路径 /")),
		mcp.WithNumber("depth", mcp.Description("最大深度，0 表示不限制")),
	), handleDocumentList)

	s.AddTool(mcp.NewTool("document_get",
		mcp.WithDescription("获取文档的 Markdown 内容"),
		mcp.WithString("id", mcp.Required(), mcp.Description("文档块 ID")),
	), handleDocumentGet)

	s.AddTool(mcp.NewTool("document_outline",
		mcp.WithDescription("获取文档的大纲结构（标题层级）"),
		mcp.WithString("id", mcp.Required(), mcp.Description("文档块 ID")),
	), handleDocumentOutline)

	s.AddTool(mcp.NewTool("block_get",
		mcp.WithDescription("获取块信息，包含内容、类型、路径等"),
		mcp.WithString("id", mcp.Required(), mcp.Description("块 ID")),
	), handleBlockGet)

	s.AddTool(mcp.NewTool("block_get_kramdown",
		mcp.WithDescription("获取块的 kramdown 源码"),
		mcp.WithString("id", mcp.Required(), mcp.Description("块 ID")),
	), handleBlockGetKramdown)

	s.AddTool(mcp.NewTool("search_fulltext",
		mcp.WithDescription("全文搜索思源笔记中的块内容"),
		mcp.WithString("query", mcp.Required(), mcp.Description("搜索关键词")),
	), handleSearchFulltext)

	s.AddTool(mcp.NewTool("search_docs",
		mcp.WithDescription("按标题搜索文档"),
		mcp.WithString("keyword", mcp.Required(), mcp.Description("搜索关键词")),
	), handleSearchDocs)

	s.AddTool(mcp.NewTool("tag_list",
		mcp.WithDescription("列出所有标签"),
	), handleTagList)

	s.AddTool(mcp.NewTool("query_sql",
		mcp.WithDescription("执行 SQL 查询，用于高级数据检索"),
		mcp.WithString("stmt", mcp.Required(), mcp.Description("SQL 语句")),
	), handleQuerySQL)

	// ── 写入工具 ──

	s.AddTool(mcp.NewTool("document_create",
		mcp.WithDescription("创建新的 Markdown 文档"),
		mcp.WithString("notebook", mcp.Required(), mcp.Description("笔记本名称或 ID")),
		mcp.WithString("markdown", mcp.Required(), mcp.Description("Markdown 内容")),
		mcp.WithString("path", mcp.Description("文档路径，不包含文件名，默认为 /")),
		mcp.WithString("title", mcp.Description("文档标题，默认取 Markdown 首个标题")),
	), handleDocumentCreate)

	s.AddTool(mcp.NewTool("daily_note_create",
		mcp.WithDescription("在指定笔记本中创建今日日记"),
		mcp.WithString("notebook", mcp.Required(), mcp.Description("笔记本名称或 ID")),
	), handleDailyNoteCreate)

	s.AddTool(mcp.NewTool("block_update",
		mcp.WithDescription("更新块内容（支持 Markdown 和 DOM 格式）"),
		mcp.WithString("id", mcp.Required(), mcp.Description("块 ID")),
		mcp.WithString("data", mcp.Required(), mcp.Description("新内容")),
		mcp.WithString("data_type", mcp.Description("内容类型：markdown（默认）或 dom")),
	), handleBlockUpdate)

	s.AddTool(mcp.NewTool("block_append",
		mcp.WithDescription("向文档末尾追加新块"),
		mcp.WithString("parent_id", mcp.Required(), mcp.Description("父块 ID（通常是文档 ID）")),
		mcp.WithString("data", mcp.Required(), mcp.Description("追加的内容")),
		mcp.WithString("data_type", mcp.Description("内容类型：markdown（默认）或 dom")),
	), handleBlockAppend)

	return server.ServeStdio(s)
}

// getClient 获取思源 API 客户端
func getClient() (*siyuan.Client, error) {
	return siyuan.GetDefaultSiYuanClient()
}

// resolveNotebook 解析笔记本标识符为 ID
func resolveNotebook(ctx context.Context, client *siyuan.Client, identifier string) (string, error) {
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return "", fmt.Errorf("获取笔记本列表失败: %w", err)
	}
	id, _, err := document.FindNotebook(notebooks, identifier)
	return id, err
}

// jsonResult 快速构建 JSON 文本结果
func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("序列化失败: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

// apiError 格式化 API 错误
func apiError(err error) (*mcp.CallToolResult, error) {
	if syErr, ok := siyuan.IsAPIError(err); ok {
		return mcp.NewToolResultError(fmt.Sprintf("思源 API 错误 (code=%d): %s", syErr.Code, syErr.Msg)), nil
	}
	return mcp.NewToolResultError(err.Error()), nil
}

// ── 只读 handler ──

func handleNotebookList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return apiError(err)
	}
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return apiError(err)
	}
	type nb struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Icon   string `json:"icon"`
		Closed bool   `json:"closed"`
	}
	list := make([]nb, len(notebooks))
	for i, n := range notebooks {
		list[i] = nb{ID: n.ID, Name: n.Name, Icon: n.Icon, Closed: n.Closed}
	}
	return jsonResult(list)
}

func handleDocumentList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nbName, err := req.RequireString("notebook")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	nbID, err := resolveNotebook(ctx, client, nbName)
	if err != nil {
		return apiError(err)
	}

	path := req.GetString("path", "/")
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	tree, err := client.ListDocTree(ctx, siyuan.ListDocTreeOptions{
		NotebookID: nbID,
		Path:       path,
	})
	if err != nil {
		return apiError(err)
	}

	type node struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		HPath string  `json:"hpath"`
		Icon  string  `json:"icon,omitempty"`
		Child []node  `json:"children,omitempty"`
	}

	var buildTree func(nodes []siyuan.DocTreeNode) []node
	buildTree = func(nodes []siyuan.DocTreeNode) []node {
		result := make([]node, len(nodes))
		for i, n := range nodes {
			item := node{ID: n.ID, Name: n.Name, HPath: n.HPath, Icon: n.Icon}
			if len(n.Children) > 0 {
				item.Child = buildTree(n.Children)
			}
			result[i] = item
		}
		return result
	}

	return jsonResult(map[string]any{
		"notebook": nbName,
		"tree":     buildTree(tree.Tree),
	})
}

func handleDocumentGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	content, hpath, err := client.ExportMdContent(ctx, id)
	if err != nil {
		return apiError(err)
	}

	info, _ := client.QuerySQL(ctx,
		fmt.Sprintf("SELECT created, updated FROM blocks WHERE id = '%s' LIMIT 1", id))
	result := map[string]any{
		"id":      id,
		"hpath":   hpath,
		"content": content,
	}
	if len(info) > 0 {
		result["created"] = info[0]["created"]
		result["updated"] = info[0]["updated"]
	}
	return jsonResult(result)
}

func handleDocumentOutline(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	outline, err := client.GetDocOutline(ctx, id)
	if err != nil {
		return apiError(err)
	}
	if outline == nil {
		outline = []siyuan.OutlineItem{}
	}
	return jsonResult(outline)
}

func handleBlockGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	rows, err := client.QuerySQL(ctx,
		fmt.Sprintf("SELECT id, type, content, hpath, root_id, parent_id, created, updated FROM blocks WHERE id = '%s' LIMIT 1", id))
	if err != nil {
		return apiError(err)
	}
	if len(rows) == 0 {
		return mcp.NewToolResultError("block not found"), nil
	}
	return jsonResult(rows[0])
}

func handleBlockGetKramdown(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	kramdown, err := client.GetBlockKramdown(ctx, id)
	if err != nil {
		return apiError(err)
	}
	return jsonResult(map[string]any{
		"id":       id,
		"kramdown": kramdown,
	})
}

func handleSearchFulltext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	blocks, err := client.FullTextSearchBlock(ctx, siyuan.FullTextSearchBlockOptions{
		Query: query,
	})
	if err != nil {
		return apiError(err)
	}
	if blocks == nil {
		blocks = []siyuan.FullTextSearchBlockResult{}
	}
	return jsonResult(blocks)
}

func handleSearchDocs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, err := req.RequireString("keyword")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	docs, err := client.SearchDocs(ctx, siyuan.SearchDocsOptions{K: keyword})
	if err != nil {
		return apiError(err)
	}
	if docs == nil {
		docs = []siyuan.SearchDocResult{}
	}
	return jsonResult(docs)
}

func handleTagList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	tags, err := client.GetTag(ctx)
	if err != nil {
		return apiError(err)
	}
	if tags == nil {
		tags = []*siyuan.TagNode{}
	}
	return jsonResult(tags)
}

func handleQuerySQL(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stmt, err := req.RequireString("stmt")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	rows, err := client.QuerySQL(ctx, stmt)
	if err != nil {
		return apiError(err)
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	return jsonResult(rows)
}

// ── 写入 handler ──

func handleDocumentCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nbName, err := req.RequireString("notebook")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	markdown, err := req.RequireString("markdown")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	nbID, err := resolveNotebook(ctx, client, nbName)
	if err != nil {
		return apiError(err)
	}

	path := req.GetString("path", "/")
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	opts := siyuan.CreateDocWithMdOptions{
		Notebook: nbID,
		Path:     path,
		Markdown: markdown,
		Title:    req.GetString("title", ""),
	}

	data, err := client.CreateDocWithMd(ctx, opts)
	if err != nil {
		return apiError(err)
	}

	// API 可能返回纯字符串 ID 或结构体
	type createResult struct {
		ID     string `json:"id"`
		HPath  string `json:"hPath"`
		RootID string `json:"rootID"`
	}
	var result createResult
	b, _ := json.Marshal(data)
	if err := json.Unmarshal(b, &result); err != nil {
		result.ID = string(b)
	}

	return jsonResult(map[string]any{
		"id":       result.ID,
		"hpath":    result.HPath,
		"notebook": nbName,
		"path":     path,
	})
}

func handleDailyNoteCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nbName, err := req.RequireString("notebook")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	nbID, err := resolveNotebook(ctx, client, nbName)
	if err != nil {
		return apiError(err)
	}

	result, err := client.CreateDailyNote(ctx, siyuan.CreateDailyNoteOptions{
		Notebook: nbID,
		App:      "siyuan-cli",
	})
	if err != nil {
		return apiError(err)
	}

	return jsonResult(result)
}

func handleBlockUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, err := req.RequireString("data")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	dataType := req.GetString("data_type", "markdown")
	if err := client.UpdateBlock(ctx, siyuan.UpdateBlockOptions{
		ID:       id,
		Data:     data,
		DataType: dataType,
	}); err != nil {
		return apiError(err)
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"id":"%s","updated":true}`, id)), nil
}

func handleBlockAppend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	parentID, err := req.RequireString("parent_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, err := req.RequireString("data")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient()
	if err != nil {
		return apiError(err)
	}

	dataType := req.GetString("data_type", "markdown")
	newID, err := client.AppendBlock(ctx, siyuan.AppendBlockOptions{
		ParentID: parentID,
		Data:     data,
		DataType: dataType,
	})
	if err != nil {
		return apiError(err)
	}

	return jsonResult(map[string]any{
		"id":        newID,
		"parent_id": parentID,
	})
}
