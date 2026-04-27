package siyuan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

// ExportDocOptions 导出文档选项
type ExportDocOptions struct {
	ID         string `json:"id"`
	Index      int    `json:"index,omitempty"`
	IsBacklink bool   `json:"isBacklink,omitempty"`
	PDF        bool   `json:"pdf"`
}

// ExportHTML 导出文档为 HTML
// 路由：POST /api/export/exportMdHTML
// 直接返回 HTML 内容（不依赖服务端临时文件）
func (c *Client) ExportHTML(ctx context.Context, opts ExportDocOptions) (string, error) {
	var ret struct {
		Content string `json:"content"`
	}
	if err := c.post(ctx, "/api/export/exportMdHTML", opts, &ret); err != nil {
		return "", err
	}
	return ret.Content, nil
}

// ExportDocx 导出文档为 Word
// 路由：POST /api/export/exportDocx
func (c *Client) ExportDocx(ctx context.Context, opts ExportDocOptions) (string, error) {
	var ret struct {
		Path string `json:"path"`
	}
	if err := c.post(ctx, "/api/export/exportDocx", opts, &ret); err != nil {
		return "", err
	}
	return ret.Path, nil
}

// ExportNotebookMdOptions 导出笔记本为 Markdown 选项
type ExportNotebookMdOptions struct {
	Notebook string `json:"notebook"`
}

// ExportNotebookSYOptions 导出笔记本为思源格式选项
type ExportNotebookSYOptions struct {
	ID string `json:"id"`
}

// ExportNotebookMdResponse 导出笔记本响应
type ExportNotebookMdResponse struct {
	Name string `json:"name"` // zip 文件名
	Zip  string `json:"zip"`  // 服务端路径如 /export/xxx.zip
}

// ExportNotebookMd 导出笔记本为 Markdown
// 路由：POST /api/export/exportNotebookMd
// 使用 http.DefaultClient 避免默认 30s 超时限制（大笔记本导出耗时较长）
func (c *Client) ExportNotebookMd(ctx context.Context, opts ExportNotebookMdOptions) (*ExportNotebookMdResponse, error) {
	var ret ExportNotebookMdResponse
	if err := c.postLongRunning(ctx, "/api/export/exportNotebookMd", opts, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// ExportNotebookSY 导出笔记本为思源格式
// 路由：POST /api/export/exportNotebookSY
// 注意：该 API 参数为 id（笔记本 ID），且返回值只有 zip 路径，没有 name
func (c *Client) ExportNotebookSY(ctx context.Context, opts ExportNotebookSYOptions) (*ExportNotebookMdResponse, error) {
	var ret ExportNotebookMdResponse
	if err := c.postLongRunning(ctx, "/api/export/exportNotebookSY", opts, &ret); err != nil {
		return nil, err
	}
	// 从 zip 路径提取文件名作为 name
	if ret.Name == "" && ret.Zip != "" {
		ret.Name = filepath.Base(ret.Zip)
	}
	return &ret, nil
}

// postLongRunning 与 post 相同但不使用 c.httpc（避免 30s 超时），适用于耗时较长的 API
func (c *Client) postLongRunning(ctx context.Context, route string, v interface{}, dst interface{}) error {
	url := c.base + route
	bs, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bs))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Token "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status: %d", resp.StatusCode)
	}
	var raw struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if raw.Code != 0 {
		return &SiYuanError{Code: raw.Code, Msg: raw.Msg}
	}
	if dst != nil && raw.Data != nil {
		if err := json.Unmarshal(raw.Data, dst); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	}
	return nil
}

// DownloadExport 从思源服务器下载导出的 zip 文件，支持进度回调
func (c *Client) DownloadExport(ctx context.Context, filename string, onProgress func(downloaded, total int64)) ([]byte, error) {
	url := c.base + "/export/" + filename
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建下载请求: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Token "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载文件: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	total := resp.ContentLength
	var buf bytes.Buffer
	_, err = io.CopyBuffer(&buf, io.TeeReader(resp.Body, &progressWriter{total: total, onProgress: onProgress}), nil)
	if err != nil {
		return nil, fmt.Errorf("读取下载内容: %w", err)
	}
	return buf.Bytes(), nil
}

type progressWriter struct {
	total     int64
	downloaded int64
	onProgress func(downloaded, total int64)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.downloaded += int64(n)
	if w.onProgress != nil {
		w.onProgress(w.downloaded, w.total)
	}
	return n, nil
}
