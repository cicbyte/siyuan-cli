package siyuan

import (
	"context"
)

// SyncInfo 同步信息
type SyncInfo struct {
	Synced    int    `json:"synced"`
	Conflict  int    `json:"conflict"`
	SyncSize  int64  `json:"syncSize"`
	LastSync  string `json:"lastSync"`
}

// GetSyncInfo 获取同步信息
// 路由：POST /api/sync/getSyncInfo
func (c *Client) GetSyncInfo(ctx context.Context) (*SyncInfo, error) {
	var ret SyncInfo
	if err := c.post(ctx, "/api/sync/getSyncInfo", map[string]interface{}{}, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// PerformSync 执行同步
// 路由：POST /api/sync/performSync
func (c *Client) PerformSync(ctx context.Context) error {
	return c.post(ctx, "/api/sync/performSync", map[string]interface{}{}, nil)
}
