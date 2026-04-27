package block

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
	"go.uber.org/zap"
)

type GetBlockOptions struct {
	ID         string
	OutputFile string
}

func GetBlock(opts GetBlockOptions) error {
	logger := log.GetLogger()
	logger.Info("获取块信息", zap.String("id", opts.ID))

	if strings.TrimSpace(opts.ID) == "" {
		fmt.Println("❌ 错误: 请提供块 ID")
		fmt.Println("💡 使用方法: siyuan-cli block get <block-id>")
		return fmt.Errorf("块 ID 不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用 SQL 查询 blocks 表获取块信息（思源无 getBlockInfo API）
	stmt := fmt.Sprintf(
		"SELECT b.id, b.type, b.content, b.created, b.updated, b.root_id, d.hpath "+
			"FROM blocks b LEFT JOIN blocks d ON b.root_id = d.id WHERE b.id = '%s' LIMIT 1",
		opts.ID,
	)
	rows, err := client.QuerySQL(ctx, stmt)
	if err != nil {
		fmt.Printf("❌ 获取块信息失败: %v\n", err)
		return err
	}
	if len(rows) == 0 {
		fmt.Printf("❌ 未找到块: %s\n", opts.ID)
		return fmt.Errorf("块不存在")
	}

	row := rows[0]
	getStr := func(key string) string {
		v, _ := row[key].(string)
		return v
	}

	info := map[string]string{
		"ID":       getStr("id"),
		"类型":     blockTypeLabel(getStr("type")),
		"内容":     getStr("content"),
		"文档":     getStr("hpath"),
		"根文档ID": getStr("root_id"),
		"创建":     formatBlockTime(getStr("created")),
		"更新":     formatBlockTime(getStr("updated")),
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		return outputJSON(row, opts.OutputFile)
	}

	content := strings.ReplaceAll(info["内容"], "\n", " ")
	fmt.Printf("块 ID: %s\n", info["ID"])
	fmt.Printf("类型: %s\n", info["类型"])
	fmt.Printf("内容: %s\n", output.Truncate(content, 80))
	fmt.Printf("文档: %s\n", info["文档"])
	fmt.Printf("创建: %s\n", info["创建"])
	fmt.Printf("更新: %s\n", info["更新"])
	return nil
}

type SourceOptions struct {
	ID         string
	OutputFile string
}

func GetBlockSource(opts SourceOptions) error {
	logger := log.GetLogger()
	logger.Info("获取块源码", zap.String("id", opts.ID))

	if strings.TrimSpace(opts.ID) == "" {
		fmt.Println("❌ 错误: 请提供块 ID")
		fmt.Println("💡 使用方法: siyuan-cli block source <block-id>")
		return fmt.Errorf("块 ID 不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kramdown, err := client.GetBlockKramdown(ctx, opts.ID)
	if err != nil {
		fmt.Printf("❌ 获取块源码失败: %v\n", err)
		return err
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		data := map[string]any{"id": opts.ID, "kramdown": kramdown}
		return outputJSON(data, opts.OutputFile)
	}

	fmt.Println(kramdown)
	return nil
}

func outputJSON(data any, outputFile string) error {
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

func blockTypeLabel(t string) string {
	labels := map[string]string{
		"d":    "文档",
		"h":    "标题",
		"p":    "段落",
		"NodeDocument":   "文档",
		"NodeHeading":    "标题",
		"NodeParagraph":  "段落",
		"NodeList":       "列表",
		"NodeListItem":   "列表项",
		"NodeBlockquote": "引用",
		"NodeCodeBlock":  "代码块",
		"NodeTable":      "表格",
		"NodeHR":         "分隔线",
		"NodeSuperBlock": "超级块",
		"NodeIFrame":     "嵌入块",
		"NodeWidget":     "挂件块",
		"NodeAudio":      "音频",
		"NodeVideo":      "视频",
	}
	if label, ok := labels[t]; ok {
		return label
	}
	return t
}

func formatBlockTime(s string) string {
	if len(s) != 14 {
		return s
	}
	return fmt.Sprintf("%s-%s-%s %s:%s:%s", s[0:4], s[4:6], s[6:8], s[8:10], s[10:12], s[12:14])
}
