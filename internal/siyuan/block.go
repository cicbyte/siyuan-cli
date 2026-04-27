package siyuan

import "context"

// GetBlockKramdown 获取块 kramdown 源码
// 路由：POST /api/block/getBlockKramdown
func (c *Client) GetBlockKramdown(ctx context.Context, id string) (string, error) {
	var ret struct {
		ID      string `json:"id"`
		Kramdown string `json:"kramdown"`
	}
	if err := c.post(ctx, "/api/block/getBlockKramdown", map[string]string{"id": id}, &ret); err != nil {
		return "", err
	}
	return ret.Kramdown, nil
}

// UpdateBlockOptions 更新块选项
type UpdateBlockOptions struct {
	ID       string `json:"id"`
	DataType string `json:"dataType"`
	Data     string `json:"data"`
}

// UpdateBlock 更新块
// 路由：POST /api/block/updateBlock
func (c *Client) UpdateBlock(ctx context.Context, opts UpdateBlockOptions) error {
	return c.post(ctx, "/api/block/updateBlock", opts, nil)
}

// AppendBlockOptions 追加块选项
type AppendBlockOptions struct {
	ParentID string `json:"parentID"`
	DataType string `json:"dataType"`
	Data     string `json:"data"`
}

// AppendBlock 追加子块到文档末尾
// 路由：POST /api/block/appendBlock
func (c *Client) AppendBlock(ctx context.Context, opts AppendBlockOptions) (string, error) {
	var ret []struct {
		ID string `json:"id"`
	}
	if err := c.post(ctx, "/api/block/appendBlock", opts, &ret); err != nil {
		return "", err
	}
	if len(ret) > 0 {
		return ret[0].ID, nil
	}
	return "", nil
}

// DeleteBlock 删除块
// 路由：POST /api/block/deleteBlock
func (c *Client) DeleteBlock(ctx context.Context, id string) error {
	return c.post(ctx, "/api/block/deleteBlock", map[string]string{"id": id}, nil)
}

// SetBlockAttrsOptions 设置块属性选项
type SetBlockAttrsOptions struct {
	ID    string            `json:"id"`
	Attrs map[string]string `json:"attrs"`
}

// SetBlockAttrs 设置块属性
// 路由：POST /api/attr/setBlockAttrs
func (c *Client) SetBlockAttrs(ctx context.Context, opts SetBlockAttrsOptions) error {
	return c.post(ctx, "/api/attr/setBlockAttrs", opts, nil)
}

// GetBlockAttrsOptions 获取块属性选项
type GetBlockAttrsResult struct {
	ID    string            `json:"id"`
	Attrs map[string]string `json:"attrs"`
}

// GetBlockAttrs 获取块属性
// 路由：POST /api/attr/getBlockAttrs
func (c *Client) GetBlockAttrs(ctx context.Context, id string) (*GetBlockAttrsResult, error) {
	var ret GetBlockAttrsResult
	if err := c.post(ctx, "/api/attr/getBlockAttrs", map[string]string{"id": id}, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
