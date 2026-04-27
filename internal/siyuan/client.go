// Package siyuan 为思源笔记（SiYuan Note）内核 API 提供类型安全、注释详尽的 Go SDK。
// 所有方法均封装了官方文档中的 POST 接口，自动完成鉴权、JSON 序列化/反序列化。
//
// 快速开始：
//
//	client := siyuan.New("http://127.0.0.1:6806", "") // 第二个参数为 token，若未设置可留空
//	books, _ := client.ListNotebooks()
//	for _, b := range books {
//		fmt.Println(b.Name)
//	}
//
// 错误处理：
//
//	所有方法返回的 error 均实现了 *SiYuanError 类型，可获取业务 code 与 msg。
package siyuan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ============================================================================
// 客户端定义
// ============================================================================

// Client 是思源 API 的客户端，线程安全，可复用。
type Client struct {
	base  string       // 形如 http://127.0.0.1:6806，结尾无 /
	token string       // 若内核设置了访问授权，则填写；否则留空
	httpc *http.Client // 可替换为带超时、代理、Cookie 等的自定义 Client
}

// Config 客户端配置
type Config struct {
	BaseURL    string        // 思源笔记基础 URL，如 "http://127.0.0.1:6806"
	Token      string        // 访问令牌（可选）
	Timeout    time.Duration // 请求超时时间
	HTTPClient *http.Client  // 自定义 HTTP 客户端（可选）
}

// New 创建新的思源笔记客户端
func New(baseURL, token string) *Client {
	return &Client{
		base:  trimSlash(baseURL),
		token: token,
		httpc: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewWithConfig 使用配置创建客户端
func NewWithConfig(cfg Config) *Client {
	if len(cfg.BaseURL) > 0 && cfg.BaseURL[len(cfg.BaseURL)-1] == '/' {
		cfg.BaseURL = cfg.BaseURL[:len(cfg.BaseURL)-1]
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		base:  trimSlash(cfg.BaseURL),
		token: cfg.Token,
		httpc: httpClient,
	}
}

// SetHTTPClient 可注入自定义 *http.Client（如超时、代理、日志中间件等）
func (c *Client) SetHTTPClient(h *http.Client) { c.httpc = h }

// SetToken 设置访问令牌
func (c *Client) SetToken(token string) { c.token = token }

// trimSlash 移除 URL 末尾的斜杠
func trimSlash(url string) string {
	if len(url) > 0 && url[len(url)-1] == '/' {
		return url[:len(url)-1]
	}
	return url
}

// ============================================================================
// 基础请求封装
// ============================================================================

// post 统一封装 POST /api/xxx 请求。
// 入参 v 会被 json.Marshal 为 body；出参 dst 为反序列化目标，nil 时忽略 body。
func (c *Client) post(ctx context.Context, route string, v interface{}, dst interface{}) error {
	url := c.base + route
	var body io.Reader
	if v != nil {
		bs, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("json marshal: %w", err)
		}
		body = bytes.NewReader(bs)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Token "+c.token)
	}
	resp, err := c.httpc.Do(req)
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
		return fmt.Errorf("json decode wrapper: %w", err)
	}
	if raw.Code != 0 {
		return &SiYuanError{Code: raw.Code, Msg: raw.Msg}
	}
	if dst != nil && len(raw.Data) > 0 {
		if err := json.Unmarshal(raw.Data, dst); err != nil {
			return fmt.Errorf("json decode data: %w", err)
		}
	}
	return nil
}

// postMultipart 发送 multipart/form-data 请求
func (c *Client) postMultipart(ctx context.Context, route string, fields map[string]io.Reader, dst interface{}) error {
	url := c.base + route

	// 创建 multipart buffer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加表单字段
	for key, r := range fields {
		if file, ok := r.(*os.File); ok {
			// 处理文件字段
			part, err := writer.CreateFormFile(key, filepath.Base(file.Name()))
			if err != nil {
				return fmt.Errorf("create form file: %w", err)
			}
			if _, err := io.Copy(part, file); err != nil {
				return fmt.Errorf("copy file: %w", err)
			}
		} else {
			// 处理普通字段
			part, err := writer.CreateFormField(key)
			if err != nil {
				return fmt.Errorf("create form field: %w", err)
			}
			if _, err := io.Copy(part, r); err != nil {
				return fmt.Errorf("copy field: %w", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Token "+c.token)
	}

	resp, err := c.httpc.Do(req)
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
		return fmt.Errorf("json decode wrapper: %w", err)
	}
	if raw.Code != 0 {
		return &SiYuanError{Code: raw.Code, Msg: raw.Msg}
	}
	if dst != nil && len(raw.Data) > 0 {
		if err := json.Unmarshal(raw.Data, dst); err != nil {
			return fmt.Errorf("json decode data: %w", err)
		}
	}
	return nil
}

// SiYuanError 代表思源业务错误。
type SiYuanError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *SiYuanError) Error() string {
	return fmt.Sprintf("siyuan error code=%d msg=%s", e.Code, e.Msg)
}

// IsAPIError 判断是否为 API 错误
func IsAPIError(err error) (*SiYuanError, bool) {
	if syErr, ok := err.(*SiYuanError); ok {
		return syErr, true
	}
	return nil, false
}

// ============================================================================
// 工具函数
// ============================================================================

// M 快速构造 map[string]interface{}，写示例代码更短。
func M(kv ...interface{}) map[string]interface{} {
	if len(kv)%2 != 0 {
		panic("M: odd number of args")
	}
	m := make(map[string]interface{}, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			panic("M: key must be string")
		}
		m[key] = kv[i+1]
	}
	return m
}

// StringPtr 返回字符串指针，用于可选参数
func StringPtr(s string) *string { return &s }

// IntPtr 返回整数指针，用于可选参数
func IntPtr(i int) *int { return &i }
