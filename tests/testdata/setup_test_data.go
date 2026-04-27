// +build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "http://127.0.0.1:6806"
const token = "aqkxinx9l7dn5cvn"

type apiResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func post(path string, data interface{}) (json.RawMessage, error) {
	bs, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewReader(bs))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+token)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var ret apiResponse
	if err := json.Unmarshal(body, &ret); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if ret.Code != 0 {
		return nil, fmt.Errorf("API error (code=%d): %s", ret.Code, ret.Msg)
	}
	return ret.Data, nil
}

func main() {
	data, err := post("/api/notebook/lsNotebooks", map[string]interface{}{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "list notebooks: %v\n", err)
		os.Exit(1)
	}

	var nbData struct {
		Notebooks []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"notebooks"`
	}
	json.Unmarshal(data, &nbData)

	var notebookID string
	for _, nb := range nbData.Notebooks {
		if nb.Name == "siyuan-test" {
			notebookID = nb.ID
			break
		}
	}
	if notebookID == "" {
		fmt.Fprintf(os.Stderr, "siyuan-test notebook not found\n")
		os.Exit(1)
	}
	fmt.Printf("Found notebook: siyuan-test (ID: %s)\n", notebookID)

	// Create test document
	md1 := `# 测试文档

这是一个用于集成测试的文档。

## 第一节

这是第一节的内容，包含一些**加粗**文本和*斜体*文本。

### 要点列表

- 要点一：这是一个列表项
- 要点二：这是另一个列表项
- 要点三：包含代码的列表项

## 第二节

第二节内容，包含代码块：

` + "```go\nfunc hello() string {\n    return \"world\"\n}\n```" + `

> 这是一段引用文本

## 标签测试

本段包含一些测试标签用的关键词：测试 集成 CLI

---

*最后更新于 2026-04-26*`

	docData, err := post("/api/filetree/createDocWithMd", map[string]interface{}{
		"notebook": notebookID,
		"path":     "/",
		"markdown": md1,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create doc: %v\n", err)
		os.Exit(1)
	}
	var doc1ID string
	json.Unmarshal(docData, &doc1ID)
	fmt.Printf("Created doc (ID: %s)\n", doc1ID)

	// Get actual HPath from block info
	blockData, err := post("/api/block/getBlockInfo", map[string]interface{}{"id": doc1ID})
	if err != nil {
		fmt.Fprintf(os.Stderr, "get block info: %v\n", err)
		os.Exit(1)
	}
	var info struct {
		HPath     string `json:"hPath"`
		RootTitle string `json:"rootTitle"`
	}
	json.Unmarshal(blockData, &info)

	docPath := info.RootTitle
	if docPath == "" || docPath == "未命名文档" {
		// Extract title from markdown
		docPath = "未命名文档"
	}
	hPath := info.HPath
	if hPath == "" {
		hPath = "/" + docPath
	}

	fmt.Printf("Doc HPath: %s, Title: %s\n", hPath, info.RootTitle)

	// Output env vars for test usage
	fmt.Println("\n=== Test Environment Variables ===")
	fmt.Printf("TEST_NOTEBOOK=siyuan-test\n")
	fmt.Printf("TEST_DOC_PATH=%s\n", hPath)
	fmt.Printf("TEST_BLOCK_ID=%s\n", doc1ID)
}
