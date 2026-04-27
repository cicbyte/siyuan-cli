package siyuan

import (
	"context"
	"fmt"
	"time"
)

// Notebook 代表一个笔记本元信息。
type Notebook struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Sort   int    `json:"sort"`
	Closed bool   `json:"closed"`
}

// NotebookConf 代表笔记本的详细配置信息。
type NotebookConf struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Sort     int       `json:"sort"`
	Icon     string    `json:"icon"`
	IconType int       `json:"iconType"` // 0=emoji, 1=custom icon
	Closed   bool      `json:"closed"`
	SortMode int       `json:"sortMode"` // 排序模式
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
	// 其他可能的配置字段
	Avatar   string `json:"avatar,omitempty"`   // 头像
	Algo     string `json:"algo,omitempty"`     // 排序算法
	RefCreateAnchor string `json:"refCreateAnchor,omitempty"` // 引用创建锚点
	DocCreateSaveFolder string `json:"docCreateSaveFolder,omitempty"` // 文档创建保存文件夹
}

// ListNotebooks 列出全部笔记本。
// 路由：POST /api/notebook/lsNotebooks
func (c *Client) ListNotebooks(ctx context.Context) ([]Notebook, error) {
	var ret struct {
		Notebooks []Notebook `json:"notebooks"`
	}
	err := c.post(ctx, "/api/notebook/lsNotebooks", nil, &ret)
	return ret.Notebooks, err
}

// OpenNotebook 打开指定笔记本。
// 路由：POST /api/notebook/openNotebook
func (c *Client) OpenNotebook(ctx context.Context, notebook string) error {
	return c.post(ctx, "/api/notebook/openNotebook", map[string]interface{}{"notebook": notebook}, nil)
}

// CloseNotebook 关闭指定笔记本。
// 路由：POST /api/notebook/closeNotebook
func (c *Client) CloseNotebook(ctx context.Context, notebook string) error {
	return c.post(ctx, "/api/notebook/closeNotebook", map[string]interface{}{"notebook": notebook}, nil)
}

// CreateNotebook 创建笔记本。
// 返回值中 Notebook.ID 为新笔记本 ID。
// 路由：POST /api/notebook/createNotebook
func (c *Client) CreateNotebook(ctx context.Context, name string) (*Notebook, error) {
	var ret struct {
		Notebook Notebook `json:"notebook"`
	}
	err := c.post(ctx, "/api/notebook/createNotebook", map[string]interface{}{"name": name}, &ret)
	return &ret.Notebook, err
}

// CreateNotebookWithIcon 创建带图标的笔记本。
// 返回值中 Notebook.ID 为新笔记本 ID。
// 路由：POST /api/notebook/createNotebookWithIcon
func (c *Client) CreateNotebookWithIcon(ctx context.Context, name, icon string) (*Notebook, error) {
	var ret struct {
		Notebook Notebook `json:"notebook"`
	}
	err := c.post(ctx, "/api/notebook/createNotebookWithIcon", map[string]interface{}{"notebook": name, "icon": icon}, &ret)
	return &ret.Notebook, err
}

// RenameNotebook 重命名笔记本。
// 路由：POST /api/notebook/renameNotebook
func (c *Client) RenameNotebook(ctx context.Context, notebook, newName string) error {
	return c.post(ctx, "/api/notebook/renameNotebook", map[string]interface{}{"notebook": notebook, "name": newName}, nil)
}

// RemoveNotebook 删除笔记本。
// 路由：POST /api/notebook/removeNotebook
func (c *Client) RemoveNotebook(ctx context.Context, notebook string) error {
	return c.post(ctx, "/api/notebook/removeNotebook", map[string]interface{}{"notebook": notebook}, nil)
}

// GetNotebookConf 获取笔记本配置信息。
// 路由：POST /api/notebook/getConf
func (c *Client) GetNotebookConf(ctx context.Context, notebook string) (*NotebookConf, error) {
	var ret struct {
		Conf NotebookConf `json:"data"`
	}
	err := c.post(ctx, "/api/notebook/getConf", map[string]interface{}{"notebook": notebook}, &ret)
	if err != nil {
		return nil, err
	}
	// 设置ID，因为API返回中可能没有包含ID字段
	ret.Conf.ID = notebook
	return &ret.Conf, nil
}

// SetNotebookConf 设置笔记本配置信息。
// 路由：POST /api/notebook/setNotebookConf
func (c *Client) SetNotebookConf(ctx context.Context, notebook string, conf *NotebookConf) error {
	// 只发送可配置的字段
	configData := map[string]interface{}{
		"name": conf.Name,
		"closed": conf.Closed,
		"sort": conf.Sort,
		"sortMode": conf.SortMode,
	}

	// 只包含非空的可选字段
	if conf.Icon != "" {
		configData["icon"] = conf.Icon
	}
	if conf.RefCreateAnchor != "" {
		configData["refCreateAnchor"] = conf.RefCreateAnchor
	}
	if conf.DocCreateSaveFolder != "" {
		configData["docCreateSaveFolder"] = conf.DocCreateSaveFolder
	}

	requestData := map[string]interface{}{
		"notebook": notebook,
		"conf":     configData,
	}

	return c.post(ctx, "/api/notebook/setNotebookConf", requestData, nil)
}

// SetNotebookConfByName 通过名称设置笔记本配置（会先查找ID）
func (c *Client) SetNotebookConfByName(ctx context.Context, notebookName string, conf *NotebookConf) error {
	// 先获取所有笔记本来查找ID
	notebooks, err := c.ListNotebooks(ctx)
	if err != nil {
		return fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	// 查找匹配的笔记本ID
	var targetID string
	for _, nb := range notebooks {
		if nb.Name == notebookName {
			targetID = nb.ID
			break
		}
	}

	if targetID == "" {
		return fmt.Errorf("未找到名称为 '%s' 的笔记本", notebookName)
	}

	return c.SetNotebookConf(ctx, targetID, conf)
}