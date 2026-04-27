package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/sashabaranov/go-openai"
)

const maxIterations = 5

type Agent struct {
	llmClient *openai.Client
	model     string
}

func NewAgent(llmClient *openai.Client, model string) *Agent {
	return &Agent{llmClient: llmClient, model: model}
}

func (a *Agent) buildSystemPrompt() string {
	now := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf(`你是思源笔记 AI 助手，可以帮助用户搜索、阅读和管理思源笔记中的内容。

当前时间：%s

## 可用工具

- notebook_list：列出所有笔记本
- document_list：列出指定笔记本下的文档树
- document_get：获取文档的 Markdown 内容
- document_outline：获取文档的大纲结构（标题层级）
- search_fulltext：全文搜索笔记块内容
- search_docs：按标题搜索文档
- tag_list：列出所有标签
- query_sql：执行 SQL 查询（blocks 表），用于高级数据检索
- document_create：创建新的 Markdown 文档
- daily_note_create：创建今日日记

## 回答策略

1. **搜索优先**：当用户问及笔记内容时，先用 search_fulltext 或 search_docs 搜索，再基于结果回答
2. **引用来源**：回答时标注信息来自哪个文档（文档标题/路径）
3. **结构化输出**：使用 Markdown 格式化回答
4. **追问确认**：如果用户意图不明确，先搜索再给出建议

## 写操作

当用户明确要求创建文档或日记时才调用 create 工具，不要主动创建。

## SQL 查询

query_sql 只用于 SELECT 查询，不要执行修改数据的语句。
blocks 表常用字段：id, type, content, hpath, box（笔记本ID）, created, updated。

## 错误处理

如果工具调用返回 error，立即停止调用该工具，直接向用户报告错误信息，不要重试。`, now)
}

func (a *Agent) getTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "notebook_list",
				Description: "列出所有思源笔记笔记本，返回 ID、名称、图标和开关状态",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "document_list",
				Description: "列出指定笔记本下的文档树结构",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"notebook": map[string]any{"type": "string", "description": "笔记本名称或 ID"},
						"path":     map[string]any{"type": "string", "description": "文档路径，默认为根路径 /"},
					},
					"required": []string{"notebook"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "document_get",
				Description: "获取文档的 Markdown 内容",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{"type": "string", "description": "文档块 ID"},
					},
					"required": []string{"id"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "document_outline",
				Description: "获取文档的大纲结构（标题层级）",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{"type": "string", "description": "文档块 ID"},
					},
					"required": []string{"id"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "search_fulltext",
				Description: "全文搜索思源笔记中的块内容",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{"type": "string", "description": "搜索关键词"},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "search_docs",
				Description: "按标题搜索文档",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"keyword": map[string]any{"type": "string", "description": "搜索关键词"},
					},
					"required": []string{"keyword"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "tag_list",
				Description: "列出所有标签",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "query_sql",
				Description: "执行 SQL 查询，用于高级数据检索（仅限 SELECT）",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"stmt": map[string]any{"type": "string", "description": "SQL 语句"},
					},
					"required": []string{"stmt"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "document_create",
				Description: "创建新的 Markdown 文档。仅当用户明确要求创建文档时调用。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"notebook": map[string]any{"type": "string", "description": "笔记本名称或 ID"},
						"markdown": map[string]any{"type": "string", "description": "Markdown 内容"},
						"path":     map[string]any{"type": "string", "description": "文档路径，不包含文件名，默认为 /"},
						"title":    map[string]any{"type": "string", "description": "文档标题，默认取 Markdown 首个标题"},
					},
					"required": []string{"notebook", "markdown"},
				},
			},
		},
		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "daily_note_create",
				Description: "在指定笔记本中创建今日日记。仅当用户明确要求创建日记时调用。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"notebook": map[string]any{"type": "string", "description": "笔记本名称或 ID"},
					},
					"required": []string{"notebook"},
				},
			},
		},
	}
}

func (a *Agent) getClient() (*siyuan.Client, error) {
	return siyuan.GetDefaultSiYuanClient()
}

func (a *Agent) resolveNotebook(ctx context.Context, client *siyuan.Client, identifier string) (string, error) {
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return "", fmt.Errorf("获取笔记本列表失败: %w", err)
	}
	id, _, err := document.FindNotebook(notebooks, identifier)
	if err == nil {
		return id, nil
	}
	identifierLower := strings.ToLower(identifier)
	for _, nb := range notebooks {
		if strings.EqualFold(nb.Name, identifier) && !nb.Closed {
			return nb.ID, nil
		}
	}
	for _, nb := range notebooks {
		if strings.Contains(strings.ToLower(nb.Name), identifierLower) && !nb.Closed {
			return nb.ID, nil
		}
	}
	return "", err
}

func debugLog(name string, result string) {
	if len(result) > 500 {
		fmt.Fprintf(os.Stderr, "[agent] result=%.100s...\n", result)
	} else {
		fmt.Fprintf(os.Stderr, "[agent] result=%s\n", result)
	}
}

func (a *Agent) executeTool(ctx context.Context, name string, args string) string {
	fmt.Fprintf(os.Stderr, "[agent] tool=%s args=%s\n", name, args)
	client, err := a.getClient()
	if err != nil {
		return fmt.Sprintf("error: 获取思源客户端失败: %v", err)
	}

	switch name {
	case "notebook_list":
		notebooks, err := client.ListNotebooks(ctx)
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		type nb struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Icon   string `json:"icon"`
			Closed bool   `json:"closed"`
		}
		list := make([]nb, len(notebooks))
		for i, n := range notebooks {
			list[i] = nb{ID: n.ID, Name: n.Name, Icon: n.Icon, Closed: n.Closed}
		}
		data, _ := json.Marshal(list)
		return string(data)

	case "document_list":
		var params struct {
			Notebook string `json:"notebook"`
			Path     string `json:"path"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		nbID, err := a.resolveNotebook(ctx, client, params.Notebook)
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		path := params.Path
		if path == "" {
			path = "/"
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		tree, err := client.ListDocTree(ctx, siyuan.ListDocTreeOptions{
			NotebookID: nbID,
			Path:       path,
		})
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		result := formatDocTree(tree.Tree, 0)
		return fmt.Sprintf("笔记本 [%s] 下的文档：\n%s", params.Notebook, result)

	case "document_get":
		var params struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		content, hpath, err := client.ExportMdContent(ctx, params.ID)
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("文档: %s\n\n%s", hpath, content)

	case "document_outline":
		var params struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		outline, err := client.GetDocOutline(ctx, params.ID)
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		if len(outline) == 0 {
			return "该文档没有大纲结构"
		}
		return formatOutline(outline, 0)

	case "search_fulltext":
		var params struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		blocks, err := client.FullTextSearchBlock(ctx, siyuan.FullTextSearchBlockOptions{
			Query: params.Query,
		})
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		if len(blocks) == 0 {
			return fmt.Sprintf("未找到包含「%s」的内容", params.Query)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "找到 %d 条结果：\n\n", len(blocks))
		for i, block := range blocks {
			content := strings.TrimSpace(block.Content)
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, block.HPath, content)
		}
		return b.String()

	case "search_docs":
		var params struct {
			Keyword string `json:"keyword"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		docs, err := client.SearchDocs(ctx, siyuan.SearchDocsOptions{K: params.Keyword})
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		if len(docs) == 0 {
			return fmt.Sprintf("未找到包含「%s」的文档", params.Keyword)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "找到 %d 个文档：\n\n", len(docs))
		for i, doc := range docs {
			fmt.Fprintf(&b, "%d. %s (ID: %s)\n", i+1, doc.Title, doc.ID)
		}
		return b.String()

	case "tag_list":
		tags, err := client.GetTag(ctx)
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		if len(tags) == 0 {
			return "没有标签"
		}
		var b strings.Builder
		for _, tag := range tags {
			fmt.Fprintf(&b, "- %s\n", tag.Label)
		}
		return b.String()

	case "query_sql":
		var params struct {
			Stmt string `json:"stmt"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		upper := strings.ToUpper(strings.TrimSpace(params.Stmt))
		if !strings.HasPrefix(upper, "SELECT") {
			return "error: 仅支持 SELECT 查询，不允许修改数据"
		}
		rows, err := client.QuerySQL(ctx, params.Stmt)
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		if len(rows) == 0 {
			return "查询结果为空"
		}
		data, _ := json.Marshal(rows)
		return string(data)

	case "document_create":
		var params struct {
			Notebook string `json:"notebook"`
			Markdown string `json:"markdown"`
			Path     string `json:"path"`
			Title    string `json:"title"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		nbID, err := a.resolveNotebook(ctx, client, params.Notebook)
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		title := params.Title
		if title == "" {
			for _, line := range strings.Split(params.Markdown, "\n") {
				line = strings.TrimSpace(line)
				if len(line) > 2 && line[0] == '#' && line[1] == ' ' {
					title = line[2:]
					break
				}
			}
		}
		if title == "" {
			title = "doc" + time.Now().Format("150405")
		}
		p := params.Path
		if p == "" || p == "/" {
			p = "/" + title
		} else {
			if p[0] != '/' {
				p = "/" + p
			}
			p = p + "/" + title
		}
		result, err := client.CreateDocWithMd(ctx, siyuan.CreateDocWithMdOptions{
			Notebook: nbID,
			Path:     p,
			Markdown: params.Markdown,
			Title:    params.Title,
		})
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("已成功创建文档 (ID: %s)", result.ID)

	case "daily_note_create":
		var params struct {
			Notebook string `json:"notebook"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		nbID, err := a.resolveNotebook(ctx, client, params.Notebook)
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		result, err := client.CreateDailyNote(ctx, siyuan.CreateDailyNoteOptions{
			Notebook: nbID,
			App:      "siyuan-cli",
		})
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("已创建日记: %s (ID: %s)", result.HPath, result.ID)

	default:
		return fmt.Sprintf("unknown tool: %s", name)
	}
}

func formatOutline(items []siyuan.OutlineItem, depth int) string {
	var b strings.Builder
	for _, item := range items {
		indent := strings.Repeat("  ", item.Depth)
		fmt.Fprintf(&b, "%s%s\n", indent, item.Content)
	}
	return b.String()
}

func formatDocTree(nodes []siyuan.DocTreeNode, depth int) string {
	var b strings.Builder
	indent := strings.Repeat("  ", depth)
	for _, n := range nodes {
		icon := ""
		if n.Icon != "" {
			icon = n.Icon + " "
		}
		fmt.Fprintf(&b, "%s%s%s (ID: %s)\n", indent, icon, n.Name, n.ID)
		if len(n.Children) > 0 {
			b.WriteString(formatDocTree(n.Children, depth+1))
		}
	}
	return b.String()
}

func (a *Agent) buildMessages(question string, history []ChatMessage) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: a.buildSystemPrompt()},
	}
	for _, msg := range history {
		messages = append(messages, openai.ChatCompletionMessage{Role: msg.Role, Content: msg.Content})
	}
	messages = append(messages, openai.ChatCompletionMessage{Role: "user", Content: question})
	return messages
}

func (a *Agent) Ask(ctx context.Context, question string, history []ChatMessage) (*AskResponse, error) {
	messages := a.buildMessages(question, history)
	tools := a.getTools()
	var totalUsage openai.Usage

	for range maxIterations {
		resp, err := a.llmClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return nil, err
		}

		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.CompletionTokens += resp.Usage.CompletionTokens

		if len(resp.Choices) == 0 || len(resp.Choices[0].Message.ToolCalls) == 0 {
			answer := ""
			if len(resp.Choices) > 0 {
				answer = resp.Choices[0].Message.Content
			}
			return &AskResponse{
				Answer:           answer,
				Model:            a.model,
				PromptTokens:     totalUsage.PromptTokens,
				CompletionTokens: totalUsage.CompletionTokens,
			}, nil
		}

		choice := resp.Choices[0]
		assistantMsg := choice.Message
		if assistantMsg.Content == "" && len(assistantMsg.ToolCalls) > 0 {
			assistantMsg.Content = " "
		}
		messages = append(messages, assistantMsg)

		for _, tc := range choice.Message.ToolCalls {
			result := a.executeTool(ctx, tc.Function.Name, tc.Function.Arguments)
			debugLog(tc.Function.Name, result)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	return nil, fmt.Errorf("agent exceeded max iterations (%d)", maxIterations)
}

func (a *Agent) AskStream(ctx context.Context, question string, history []ChatMessage, cb StreamCallback) error {
	messages := a.buildMessages(question, history)
	return a.streamLoop(ctx, messages, cb)
}

func (a *Agent) streamLoop(ctx context.Context, messages []openai.ChatCompletionMessage, cb StreamCallback) error {
	tools := a.getTools()
	var totalUsage openai.Usage

	for range maxIterations {
		stream, err := a.llmClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			cb(StreamEvent{Type: "error", Content: fmt.Sprintf("请求失败: %v", err)})
			return err
		}

		var assistantContent string
		var reasoningContent string
		toolCallMap := make(map[int]*openai.ToolCall)

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				cb(StreamEvent{Type: "error", Content: fmt.Sprintf("流式读取失败: %v", err)})
				return err
			}

			if resp.Usage != nil {
				totalUsage.PromptTokens += resp.Usage.PromptTokens
				totalUsage.CompletionTokens += resp.Usage.CompletionTokens
			}

			if len(resp.Choices) == 0 {
				continue
			}

			delta := resp.Choices[0].Delta

			if delta.ReasoningContent != "" {
				reasoningContent += delta.ReasoningContent
			}
			if delta.Content != "" {
				assistantContent += delta.Content
				if len(delta.ToolCalls) == 0 {
					cb(StreamEvent{Type: "content", Content: delta.Content})
				}
			}

			for _, tc := range delta.ToolCalls {
				idx := 0
				if tc.Index != nil {
					idx = *tc.Index
				}
				if _, ok := toolCallMap[idx]; !ok {
					toolCallMap[idx] = &openai.ToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						Function: openai.FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				} else {
					toolCallMap[idx].Function.Arguments += tc.Function.Arguments
					if tc.ID != "" {
						toolCallMap[idx].ID = tc.ID
					}
				}
			}
		}

		assistantMsg := openai.ChatCompletionMessage{Role: "assistant"}
		if reasoningContent != "" {
			assistantMsg.ReasoningContent = reasoningContent
		}
		if len(toolCallMap) > 0 {
			tcs := make([]openai.ToolCall, 0, len(toolCallMap))
			for j := range len(toolCallMap) {
				tcs = append(tcs, *toolCallMap[j])
			}
			assistantMsg.ToolCalls = tcs
		}
		if assistantContent == "" && len(toolCallMap) > 0 {
			assistantContent = " "
		}
		assistantMsg.Content = assistantContent
		messages = append(messages, assistantMsg)

		if len(toolCallMap) == 0 {
			cb(StreamEvent{
				Type:             "done",
				PromptTokens:     totalUsage.PromptTokens,
				CompletionTokens: totalUsage.CompletionTokens,
			})
			return nil
		}

		if assistantContent != "" && assistantContent != " " {
			cb(StreamEvent{Type: "content_clear"})
		}

		for j := range len(toolCallMap) {
			tc := toolCallMap[j]
			cb(StreamEvent{Type: "tool_call", Tool: tc.Function.Name})

			result := a.executeTool(ctx, tc.Function.Name, tc.Function.Arguments)
			debugLog(tc.Function.Name, result)
			summary := toolSummary(tc.Function.Name)
			cb(StreamEvent{Type: "tool_result", Tool: tc.Function.Name, Content: summary})

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	return fmt.Errorf("agent exceeded max iterations (%d)", maxIterations)
}

func toolSummary(name string) string {
	switch name {
	case "notebook_list":
		return "列出笔记本"
	case "document_list":
		return "列出文档树"
	case "document_get":
		return "获取文档内容"
	case "document_outline":
		return "获取文档大纲"
	case "search_fulltext":
		return "全文搜索"
	case "search_docs":
		return "搜索文档"
	case "tag_list":
		return "列出标签"
	case "query_sql":
		return "SQL 查询"
	case "document_create":
		return "已创建文档"
	case "daily_note_create":
		return "已创建日记"
	default:
		return "完成"
	}
}
