package siyuan

import "context"

// QuerySQL 执行 SQL 查询
// 路由：POST /api/query/sql
func (c *Client) QuerySQL(ctx context.Context, stmt string) ([]map[string]any, error) {
	var ret []map[string]any
	if err := c.post(ctx, "/api/query/sql", map[string]string{"stmt": stmt}, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
