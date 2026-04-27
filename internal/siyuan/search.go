package siyuan

import "context"

// SearchDocsOptions 搜索文档选项
type SearchDocsOptions struct {
	K        string `json:"k"`        // 搜索关键词
	Flashcard bool   `json:"flashcard"` // 是否仅搜索闪卡
}

// SearchDocResult 搜索文档结果
type SearchDocResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	HPath string `json:"hpath"`
	Box   string `json:"box"`
	Path  string `json:"path"`
}

// SearchDocs 搜索文档
// 路由：POST /api/filetree/searchDocs
func (c *Client) SearchDocs(ctx context.Context, opts SearchDocsOptions) ([]SearchDocResult, error) {
	var ret []SearchDocResult
	if err := c.post(ctx, "/api/filetree/searchDocs", opts, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// FullTextSearchBlockOptions 全文搜索块选项
type FullTextSearchBlockOptions struct {
	Query string `json:"query"`       // 搜索关键词
	Types map[string]bool `json:"types,omitempty"` // 限制块类型
}

// FullTextSearchBlockResult 全文搜索块结果
type FullTextSearchBlockResult struct {
	ID        string `json:"id"`
 Content   string `json:"content"`
 HPath     string `json:"hpath"`
 Box       string `json:"box"`
 Path      string `json:"path"`
 BlockType string `json:"type"`
}

// FullTextSearchBlock 全文搜索块
// 路由：POST /api/search/fullTextSearchBlock
func (c *Client) FullTextSearchBlock(ctx context.Context, opts FullTextSearchBlockOptions) ([]FullTextSearchBlockResult, error) {
	var ret struct {
		Blocks []FullTextSearchBlockResult `json:"blocks"`
	}
	if err := c.post(ctx, "/api/search/fullTextSearchBlock", opts, &ret); err != nil {
		return nil, err
	}
	return ret.Blocks, nil
}

// TagNode 标签树节点（getTag API 返回）
type TagNode struct {
	Name     string    `json:"name"`
	Label    string    `json:"label"`
	Children []*TagNode `json:"children"`
	Type     string    `json:"type"`
	Depth    int       `json:"depth"`
	Count    int       `json:"count"`
}

// SearchTag 搜索标签
// 路由：POST /api/search/searchTag
// 返回匹配的标签名列表（字符串数组）
func (c *Client) SearchTag(ctx context.Context, keyword string) ([]string, error) {
	var ret struct {
		Tags []string `json:"tags"`
		K    string   `json:"k"`
	}
	if err := c.post(ctx, "/api/search/searchTag", map[string]string{"k": keyword}, &ret); err != nil {
		return nil, err
	}
	return ret.Tags, nil
}

// GetTag 获取所有标签（树形结构）
// 路由：POST /api/tag/getTag
func (c *Client) GetTag(ctx context.Context) ([]*TagNode, error) {
	var ret []*TagNode
	if err := c.post(ctx, "/api/tag/getTag", map[string]interface{}{}, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// FlattenTags 将标签树展平为标签名列表
func FlattenTags(nodes []*TagNode) []string {
	var result []string
	for _, node := range nodes {
		result = append(result, node.Label)
		if len(node.Children) > 0 {
			result = append(result, FlattenTags(node.Children)...)
		}
	}
	return result
}
