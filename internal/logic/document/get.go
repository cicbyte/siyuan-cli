package document

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
	"github.com/charmbracelet/glamour"
	"go.uber.org/zap"
)

type GetOptions struct {
	NotebookIdentifier string
	Path               string
	OutputFile         string
}

type blockMeta struct {
	Created string
	Updated string
}

func getBlockMeta(ctx context.Context, client *siyuan.Client, id string) (*blockMeta, error) {
	rows, err := client.QuerySQL(ctx,
		fmt.Sprintf("SELECT created, updated FROM blocks WHERE id = '%s' LIMIT 1", id))
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	getStr := func(key string) string {
		v, _ := rows[0][key].(string)
		return v
	}
	return &blockMeta{Created: getStr("created"), Updated: getStr("updated")}, nil
}

func formatDocTime(s string) string {
	if len(s) != 14 {
		return s
	}
	return fmt.Sprintf("%s-%s-%s %s:%s:%s", s[0:4], s[4:6], s[6:8], s[8:10], s[10:12], s[12:14])
}

func GetDocument(opts GetOptions) error {
	logger := log.GetLogger()
	logger.Info("开始获取文档内容",
		zap.String("notebook", opts.NotebookIdentifier),
		zap.String("path", opts.Path))

	if !siyuan.IsSiYuanConfigValid() {
		fmt.Println("❌ 思源笔记配置无效或未启用")
		fmt.Println("请运行 'siyuan-cli auth login' 配置连接")
		return fmt.Errorf("思源笔记配置无效")
	}

	if strings.TrimSpace(opts.NotebookIdentifier) == "" {
		fmt.Println("❌ 错误: 请提供笔记本名称或ID")
		fmt.Println("💡 使用方法: siyuan-cli document get <笔记本> <文档路径>")
		return fmt.Errorf("笔记本标识符不能为空")
	}

	if strings.TrimSpace(opts.Path) == "" {
		fmt.Println("❌ 错误: 请提供文档路径")
		fmt.Println("💡 使用方法: siyuan-cli document get <笔记本> <文档路径>")
		return fmt.Errorf("文档路径不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建思源笔记客户端失败: %v\n", err)
		return fmt.Errorf("创建客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		fmt.Printf("❌ 获取笔记本列表失败: %v\n", err)
		return fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	targetID, targetName, err := FindNotebook(notebooks, opts.NotebookIdentifier)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看所有可用的笔记本")
		return err
	}

	resolvedPath, err := ResolveDocPath(ctx, client, targetID, targetName, opts.Path)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return err
	}

	docID := strings.TrimSuffix(strings.Trim(resolvedPath, "/"), "/")
	parts := strings.Split(docID, "/")
	docID = parts[len(parts)-1]

	logger.Info("获取文档内容", zap.String("doc_id", docID))

	content, hpath, err := client.ExportMdContent(ctx, docID)
	if err != nil {
		fmt.Printf("❌ 获取文档内容失败: %v\n", err)
		return fmt.Errorf("获取文档内容失败: %w", err)
	}

	blockInfo, err := getBlockMeta(ctx, client, docID)
	if err != nil {
		blockInfo = nil
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		return outputGetJSON(hpath, docID, blockInfo, content, opts.OutputFile)
	}

	printDocumentDetail(hpath, docID, blockInfo, content)
	return nil
}

func printDocumentDetail(hpath, docID string, info *blockMeta, content string) {
	fmt.Printf("路径:   %s\n", hpath)
	fmt.Printf("ID:     %s\n", docID)
	if info != nil {
		if info.Created != "" {
			fmt.Printf("创建:   %s\n", formatDocTime(info.Created))
		}
		if info.Updated != "" {
			fmt.Printf("更新:   %s\n", formatDocTime(info.Updated))
		}
	}
	fmt.Println("────────────────────────────────────────")

	w, _, _ := output.GetTermSize()
	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(w),
		glamour.WithAutoStyle(),
	)
	if err != nil {
		fmt.Print(content)
		return
	}
	out, err := renderer.Render(content)
	if err != nil {
		fmt.Print(content)
		return
	}
	fmt.Print(out)
}

func outputGetJSON(hpath, docID string, info *blockMeta, content, outputFile string) error {
	data := map[string]any{
		"path":    hpath,
		"id":      docID,
		"content": content,
	}
	if info != nil {
		data["created"] = info.Created
		data["updated"] = info.Updated
	}

	if outputFile != "" {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化JSON失败: %w", err)
		}
		dir := filepath.Dir(outputFile)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建输出目录失败: %w", err)
			}
		}
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
		fmt.Printf("已导出到: %s\n", outputFile)
		return nil
	}

	output.PrintJSON(data)
	return nil
}
