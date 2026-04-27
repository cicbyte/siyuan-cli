package siyuan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// DocFile 代表文档树中的文件或文件夹
type DocFile struct {
	Path         string `json:"path"`         // 文件或文件夹路径
	Name         string `json:"name"`         // 文件或文件夹名称
	Icon         string `json:"icon"`         // 图标
	ID           string `json:"id"`           // 文档块ID（仅文档有）
	Count        int    `json:"count"`        // 文档中的块数量
	HSize        string `json:"hSize"`        // 人类可读大小
	Size         int64  `json:"size"`         // 大小（字节）
	Updated      string `json:"updated"`      // 更新时间
	Created      string `json:"created"`      // 创建时间
	SubFileCount int    `json:"subFileCount"` // 子文件数量（仅文件夹有）
}

// DocTreeNode 代表树形结构的文档节点
type DocTreeNode struct {
	ID          string        `json:"id"`          // 节点ID
	Name        string        `json:"name"`        // 文档名称
	HPath       string        `json:"hPath"`       // 人类可读路径
	Type        string        `json:"type"`        // 类型 (document)
	Icon        string        `json:"icon"`        // 图标
	SubDocCount int           `json:"subDocCount"` // 子文档数量
	Created     string        `json:"created"`     // 创建时间
	Updated     string        `json:"updated"`     // 更新时间
	Children    []DocTreeNode `json:"children"`    // 子节点
}

// DocTreeData 表示文档树响应数据
type DocTreeData struct {
	Path  string         `json:"path"`   // 文档路径
	Files []DocFile      `json:"files,omitempty"` // 文件列表（旧格式）
	Tree  []DocTreeNode  `json:"tree,omitempty"`  // 树形结构（新格式）
}

// ListDocTreeOptions 列出文档树的选项
type ListDocTreeOptions struct {
	NotebookID string // 笔记本ID（必需）
	Path       string // 文档路径，从根路径开始（如 "/a/b"）。如果为空，则返回整个笔记本的文档树
	Sort       int    // 排序方式（0：按名称，1：按更新时间，2：按创建时间，3：自定义）
}

// ListDocTree 列出指定笔记本或文档的文档树结构
// 路由：POST /api/filetree/listDocTree
func (c *Client) ListDocTree(ctx context.Context, opts ListDocTreeOptions) (*DocTreeData, error) {
	var ret DocTreeData

	// 构建请求参数
	requestData := map[string]interface{}{
		"notebook": opts.NotebookID,
	}

	// 处理路径参数：如果为空，默认为"/"
	path := opts.Path
	if path == "" {
		path = "/"
	} else {
		// 确保路径以 / 开头
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
	}
	requestData["path"] = path

	if opts.Sort >= 0 {
		requestData["sort"] = opts.Sort
	}

	// 发送请求
	if err := c.post(ctx, "/api/filetree/listDocTree", requestData, &ret); err != nil {
		return &DocTreeData{}, err
	}

	if ret.Tree == nil {
		ret.Tree = []DocTreeNode{}
	}

	return &ret, nil
}

// IsDocFile 判断是否为文档文件（.sy结尾）
func (f *DocFile) IsDocFile() bool {
	return strings.HasSuffix(f.Path, ".sy")
}

// IsFolder 判断是否为文件夹
func (f *DocFile) IsFolder() bool {
	return !f.IsDocFile()
}

// RenameDocOptions 重命名文档的选项
type RenameDocOptions struct {
	NotebookID string `json:"notebook"` // 笔记本ID
	Path       string `json:"path"`       // 文档路径，如 "/foo/bar" 或 ""
	Title      string `json:"title"`      // 新文档标题
}

// RenameDocResponse 重命名文档的响应
type RenameDocResponse struct {
	ID     string `json:"id"`     // 文档ID
	Name   string `json:"name"`   // 新文档名称
	Path   string `json:"path"`   // 文档路径
	HPath  string `json:"hPath"`  // 人类可读的文档路径
	Icon   string `json:"icon"`   // 文档图标
}

// RenameDoc 重命名文档
// 路由：POST /api/filetree/renameDoc
func (c *Client) RenameDoc(ctx context.Context, opts RenameDocOptions) (*RenameDocResponse, error) {
	var ret struct {
		Code int                `json:"code"`
		Msg  string             `json:"msg"`
		Data RenameDocResponse `json:"data"`
	}

	// 构建请求参数
	requestData := map[string]interface{}{
		"notebook": opts.NotebookID,
		"path":     opts.Path,
		"title":    opts.Title,
	}

	// 发送请求
	if err := c.post(ctx, "/api/filetree/renameDoc", requestData, &ret); err != nil {
		return &RenameDocResponse{}, err
	}

	if ret.Code != 0 {
		return &RenameDocResponse{}, fmt.Errorf("API error (code=%d): %s", ret.Code, ret.Msg)
	}

	return &ret.Data, nil
}

// GetDocID 获取文档ID（去掉路径和扩展名）
func (f *DocFile) GetDocID() string {
	if !f.IsDocFile() {
		return ""
	}
	// 从路径中提取文件名（去掉扩展名）
	filename := filepath.Base(f.Path)
	return strings.TrimSuffix(filename, ".sy")
}

// GetFileExtension 获取文件扩展名
func (f *DocFile) GetFileExtension() string {
	return filepath.Ext(f.Path)
}

// GetHPathByID 获取文档的人类可读路径
// 路由：GET /api/filetree/getHPathByID
func (c *Client) GetHPathByID(ctx context.Context, id string) (string, error) {
	requestData := map[string]interface{}{
		"id": id,
	}

	// 先尝试作为标准API响应（有code,msg,data结构）
	var standardRet struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}

	err := c.post(ctx, "/api/filetree/getHPathByID", requestData, &standardRet)
	if err != nil {
		// 如果标准结构解析失败，尝试直接返回字符串
		var directRet string
		err = c.post(ctx, "/api/filetree/getHPathByID", requestData, &directRet)
		if err != nil {
			return "", err
		}
		return directRet, nil
	}

	if standardRet.Code != 0 {
		return "", fmt.Errorf("API error (code=%d): %s", standardRet.Code, standardRet.Msg)
	}

	return standardRet.Data, nil
}

// CreateDocWithMdOptions 创建Markdown文档的选项
type CreateDocWithMdOptions struct {
	Notebook  string `json:"notebook"`  // 笔记本ID
	Path      string `json:"path"`      // 文档路径，不包含文件名，如"/foo/bar"
	Markdown  string `json:"markdown"`  // Markdown文本内容
	Title     string `json:"title"`     // 文档标题，默认使用Markdown内容的第一个标题
}

// CreateDocWithMdResponse 创建Markdown文档的响应
type CreateDocWithMdResponse struct {
	ID     string `json:"id"`     // 新创建的文档ID
	RootID string `json:"rootID"` // 文档根ID
	Box    string `json:"box"`    // 笔记本ID
	Path   string `json:"path"`   // 文档路径
	HPath  string `json:"hPath"`  // 人类可读的文档路径
	Name   string `json:"name"`   // 文档名称
}

// CreateDocWithMd 使用Markdown内容创建新文档
// 路由：POST /api/filetree/createDocWithMd
func (c *Client) CreateDocWithMd(ctx context.Context, opts CreateDocWithMdOptions) (*CreateDocWithMdResponse, error) {
	var docID string
	requestData := map[string]interface{}{
		"notebook": opts.Notebook,
		"path":     opts.Path,
		"markdown": opts.Markdown,
	}
	if opts.Title != "" {
		requestData["title"] = opts.Title
	}
	if err := c.post(ctx, "/api/filetree/createDocWithMd", requestData, &docID); err != nil {
		return nil, err
	}
	return &CreateDocWithMdResponse{ID: docID}, nil
}

// GetIDsByHPathOptions 根据人类可读路径获取文档ID的选项
type GetIDsByHPathOptions struct {
	Notebook string `json:"notebook"` // 笔记本ID
	Path     string `json:"path"`     // 人类可读路径（HPath），如 "/笔记/子文件夹/文档标题"
}

// GetIDsByHPath 根据人类可读路径获取文档ID
// 路由：POST /api/filetree/getIDsByHPath
func (c *Client) GetIDsByHPath(ctx context.Context, opts GetIDsByHPathOptions) ([]string, error) {
	// 构建请求参数
	requestData := map[string]interface{}{
		"notebook": opts.Notebook,
		"path":     opts.Path,
	}

	// 创建请求
	url := c.base + "/api/filetree/getIDsByHPath"
	var body io.Reader
	if requestData != nil {
		bs, err := json.Marshal(requestData)
		if err != nil {
			return nil, fmt.Errorf("json marshal: %w", err)
		}
		body = bytes.NewReader(bs)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Token "+c.token)
	}

	// 发送请求
	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http status: %d", resp.StatusCode)
	}

	// 首先尝试解码为标准结构
	var standardRet struct {
		Code int      `json:"code"`
		Msg  string   `json:"msg"`
		Data []string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&standardRet); err == nil {
		if standardRet.Code != 0 {
			return nil, fmt.Errorf("API error (code=%d): %s", standardRet.Code, standardRet.Msg)
		}
		if standardRet.Data == nil {
			return []string{}, nil
		}
		return standardRet.Data, nil
	}

	// 如果标准结构解析失败，重新读取响应体
	resp.Body.Close()

	// 需要重新发送请求以获取新的响应体
	resp2, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request again: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http status: %d", resp2.StatusCode)
	}

	// 尝试直接解码为数组
	var directRet []string
	if err := json.NewDecoder(resp2.Body).Decode(&directRet); err == nil {
		if directRet == nil {
			return []string{}, nil
		}
		return directRet, nil
	}

	// 如果都失败了，返回错误
	return nil, fmt.Errorf("failed to decode response as either standard API format or direct array format")
}

// RemoveDocOptions 删除文档的选项
type RemoveDocOptions struct {
	NotebookID string `json:"notebook"` // 笔记本ID
	Path       string `json:"path"`       // 文档路径，相对于笔记本的物理路径，如 "/20231201120000-abc123.sy"
}

// RemoveDoc 删除文档
// 路由：POST /api/filetree/removeDoc
func (c *Client) RemoveDoc(ctx context.Context, opts RemoveDocOptions) error {
	// 构建请求参数
	requestData := map[string]interface{}{
		"notebook": opts.NotebookID,
		"path":     opts.Path,
	}

	// 发送请求，删除接口通常返回空data
	return c.post(ctx, "/api/filetree/removeDoc", requestData, nil)
}

// ExportMdContent 导出文档的 Markdown 内容
// 路由：POST /api/export/exportMdContent
func (c *Client) ExportMdContent(ctx context.Context, id string) (string, string, error) {
	var ret struct {
		Content string `json:"content"`
		HPath   string `json:"hpath"`
	}
	if err := c.post(ctx, "/api/export/exportMdContent", map[string]any{"id": id}, &ret); err != nil {
		return "", "", err
	}
	return ret.Content, ret.HPath, nil
}

// DuplicateDoc 复制文档
// 路由：POST /api/filetree/duplicateDoc
func (c *Client) DuplicateDoc(ctx context.Context, id string) error {
	return c.post(ctx, "/api/filetree/duplicateDoc", map[string]string{"id": id}, nil)
}

// MoveDocsOptions 移动文档选项
type MoveDocsOptions struct {
	FromPaths  []string `json:"fromPaths"`
	ToNotebook string   `json:"toNotebook"`
	ToPath     string   `json:"toPath"`
}

// MoveDocs 移动文档
// 路由：POST /api/filetree/moveDocs
func (c *Client) MoveDocs(ctx context.Context, opts MoveDocsOptions) error {
	return c.post(ctx, "/api/filetree/moveDocs", opts, nil)
}

// GetDocHistory 获取文档更新历史（兼容旧接口，已废弃）
// 路由：POST /api/history/searchHistory
func (c *Client) GetDocHistory(ctx context.Context, opts SearchHistoryOptions) (*SearchHistoryResponse, error) {
	return c.SearchHistory(ctx, opts)
}

// RollbackDocHistory 回滚文档到指定历史版本
// 路由：POST /api/history/rollbackDocHistory
func (c *Client) RollbackDocHistory(ctx context.Context, notebook, historyPath string) error {
	return c.post(ctx, "/api/history/rollbackDocHistory", RollbackDocHistoryOptions{
		Notebook:    notebook,
		HistoryPath: historyPath,
	}, nil)
}