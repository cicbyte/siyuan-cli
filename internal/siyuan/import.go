package siyuan

import "context"

// ImportStdMdOptions 导入标准 Markdown 选项
type ImportStdMdOptions struct {
	Notebook string `json:"notebook"`
	Path     string `json:"path,omitempty"`
}

// ImportStdMd 导入标准 Markdown
// 路由：POST /api/import/importStdMd
func (c *Client) ImportStdMd(ctx context.Context, opts ImportStdMdOptions) error {
	return c.post(ctx, "/api/import/importStdMd", opts, nil)
}

// ImportZipMdOptions 导入 ZIP Markdown 选项
type ImportZipMdOptions struct {
	Notebook string `json:"notebook"`
	Path     string `json:"path,omitempty"`
}

// ImportZipMd 导入 ZIP Markdown
// 路由：POST /api/import/importZipMd
func (c *Client) ImportZipMd(ctx context.Context, opts ImportZipMdOptions) error {
	return c.post(ctx, "/api/import/importZipMd", opts, nil)
}

// ImportSYOptions 导入思源格式选项
type ImportSYOptions struct {
	Notebook string `json:"notebook"`
	Path     string `json:"path,omitempty"`
}

// ImportSY 导入思源格式数据
// 路由：POST /api/import/importSY
func (c *Client) ImportSY(ctx context.Context, opts ImportSYOptions) error {
	return c.post(ctx, "/api/import/importSY", opts, nil)
}
