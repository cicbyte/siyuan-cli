package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/ai"
	"github.com/cicbyte/siyuan-cli/internal/common"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

func GetChatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat [question]",
		Short: "AI 对话，基于思源笔记数据回答问题",
		Long: `使用 AI 对话助手搜索、阅读和管理思源笔记内容。
AI 会自动调用工具搜索笔记、读取文档、创建内容等。

无参数进入交互式对话模式，提供问题参数则单次问答。

示例:
  siyuan-cli chat "列出所有笔记本"
  siyuan-cli chat "搜索关于 Go 的笔记"
  siyuan-cli chat`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !siyuan.IsSiYuanConfigValid() {
				return fmt.Errorf("思源笔记配置无效或未启用，请运行 siyuan-cli auth login 配置连接")
			}

			cfg := common.GetAppConfig()
			if cfg.AI.Provider != "ollama" && cfg.AI.ApiKey == "" {
				return fmt.Errorf("AI 未配置，请先设置 API Key：siyuan-cli config set ai.api_key <key>")
			}

			apiKey := cfg.AI.ApiKey
			if apiKey == "" {
				apiKey = "ollama"
			}

			svc := ai.NewAIService(
				cfg.AI.Provider,
				cfg.AI.BaseURL,
				apiKey,
				cfg.AI.Model,
			)

			nonStream, _ := cmd.Flags().GetBool("non-stream")

			if len(args) > 0 {
				return ask(svc, cmd.Context(), args[0], nonStream)
			}
			return runInteractive(svc, cmd.Context(), nonStream)
		},
	}

	cmd.Flags().Bool("non-stream", false, "使用非流式输出")

	return cmd
}

func ask(svc *ai.AIService, ctx context.Context, question string, nonStream bool) error {
	fmt.Printf("%s %s\n", bold("  user >"), question)
	_, err := askWithHistory(svc, ctx, question, nil, nonStream)
	return err
}

func askWithHistory(svc *ai.AIService, ctx context.Context, question string, history []ai.ChatMessage, nonStream bool) (string, error) {
	if nonStream {
		resp, err := svc.Ask(ctx, question, history)
		if err != nil {
			return "", fmt.Errorf("AI 请求失败: %w", err)
		}
		fmt.Print(renderMarkdown(resp.Answer))
		if resp.PromptTokens > 0 || resp.CompletionTokens > 0 {
			fmt.Printf("\n%s\nTokens: %d prompt + %d completion | Model: %s\n",
				dim("---"), resp.PromptTokens, resp.CompletionTokens, resp.Model)
		}
		return resp.Answer, nil
	}

	start := time.Now()
	var buf strings.Builder
	var promptTokens, completionTokens int

	err := svc.AskStream(ctx, question, history, func(event ai.StreamEvent) {
		switch event.Type {
		case "content":
			buf.WriteString(event.Content)
		case "content_clear":
			buf.Reset()
		case "tool_call":
			fmt.Printf("  %s %s\n", dim("▸"), event.Tool)
		case "tool_result":
			fmt.Printf("  %s %s\n", dim("✓"), event.Content)
		case "done":
			promptTokens = event.PromptTokens
			completionTokens = event.CompletionTokens
		case "error":
			fmt.Printf("  %s\n", dim("✗ "+event.Content))
		}
	})

	if err != nil {
		fmt.Println()
		return "", fmt.Errorf("AI 请求失败: %w", err)
	}

	raw := buf.String()
	if raw == "" {
		return "", nil
	}

	fmt.Print(renderMarkdown(raw))

	if promptTokens > 0 || completionTokens > 0 {
		elapsed := time.Since(start)
		fmt.Printf("\n%s\n", dim(fmt.Sprintf("Tokens: %d + %d · %.1fs",
			promptTokens, completionTokens, elapsed.Seconds())))
	}

	return raw, nil
}

func runInteractive(svc *ai.AIService, ctx context.Context, nonStream bool) error {
	fmt.Println(bold(" AI 对话模式"))
	fmt.Println(dim("  输入问题开始对话，/quit 退出，/clear 清除上下文"))
	fmt.Println()

	var history []ai.ChatMessage
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(bold("  user > "))
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "/quit" || line == "/exit" || line == "/q" {
			fmt.Println(dim("  再见!"))
			break
		}
		if line == "/clear" {
			history = nil
			fmt.Println(dim("  对话上下文已清除"))
			fmt.Println()
			continue
		}

		resp, err := askWithHistory(svc, ctx, line, history, nonStream)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  错误: %v\n", err)
			fmt.Println()
			continue
		}

		history = append(history, ai.ChatMessage{Role: "user", Content: line})
		history = append(history, ai.ChatMessage{Role: "assistant", Content: resp})
		fmt.Println()
	}

	return nil
}

func renderMarkdown(content string) string {
	if content == "" {
		return ""
	}
	w, _, _ := output.GetTermSize()
	if w <= 0 {
		w = 80
	}
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(w),
	)
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return out
}

func dim(s string) string {
	return "\x1b[2m" + s + "\x1b[0m"
}

func bold(s string) string {
	return "\x1b[1m" + s + "\x1b[0m"
}
