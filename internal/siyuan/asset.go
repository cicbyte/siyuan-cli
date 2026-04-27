package siyuan

import (
	"context"
	"io"
	"os"
)

// UploadAsset 上传资源文件
// 路由：POST /api/asset/upload
func (c *Client) UploadAsset(ctx context.Context, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fields := map[string]io.Reader{
		"file[]": f,
	}
	var ret struct {
		SuccMap map[string]string `json:"succMap"`
	}
	if err := c.postMultipart(ctx, "/api/asset/upload", fields, &ret); err != nil {
		return "", err
	}
	for _, v := range ret.SuccMap {
		return v, nil
	}
	return "", nil
}

// DocAsset 文档关联资源
type DocAsset struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Size    int64  `json:"size"`
	Updated string `json:"updated"`
}

// GetDocAssets 获取文档资源
// 路由：POST /api/asset/getDocAssets
func (c *Client) GetDocAssets(ctx context.Context, id string) ([]DocAsset, error) {
	var ret []DocAsset
	if err := c.post(ctx, "/api/asset/getDocAssets", map[string]string{"id": id}, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// UnusedAsset 未使用资源
type UnusedAsset struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Updated string `json:"updated"`
}

// GetUnusedAssets 获取未使用资源
// 路由：POST /api/asset/getUnusedAssets
func (c *Client) GetUnusedAssets(ctx context.Context) ([]UnusedAsset, error) {
	var ret []UnusedAsset
	if err := c.post(ctx, "/api/asset/getUnusedAssets", nil, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// RemoveUnusedAssets 清理未使用资源
// 路由：POST /api/asset/removeUnusedAssets
func (c *Client) RemoveUnusedAssets(ctx context.Context) error {
	return c.post(ctx, "/api/asset/removeUnusedAssets", nil, nil)
}

// StatAssetResult 资源统计结果
type StatAssetResult struct {
	AssetCount int   `json:"assetCount"`
	AssetSize  int64 `json:"assetSize"`
}

// StatAssetInfo 统计资源
// 路由：POST /api/asset/statAsset
func (c *Client) StatAssetInfo(ctx context.Context) (*StatAssetResult, error) {
	var ret StatAssetResult
	if err := c.post(ctx, "/api/asset/statAsset", nil, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
