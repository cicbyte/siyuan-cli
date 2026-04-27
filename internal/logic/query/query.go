package query

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

type QueryOptions struct {
	SQL        string
	OutputFile string
	Raw        bool
}

func RunQuery(opts QueryOptions) error {
	logger := log.GetLogger()
	logger.Info("执行SQL查询", zap.String("sql", opts.SQL))

	stmt := strings.TrimSpace(opts.SQL)
	if stmt == "" {
		fmt.Println("❌ 错误: 请提供 SQL 语句")
		fmt.Println("💡 使用方法: siyuan-cli query \"SELECT * FROM blocks WHERE type='d' LIMIT 10\"")
		return fmt.Errorf("SQL 语句不能为空")
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	results, err := client.QuerySQL(ctx, stmt)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return err
	}

	if output.IsJSON("") || opts.OutputFile != "" || opts.Raw {
		data := map[string]any{
			"sql":    stmt,
			"count":  len(results),
			"result": results,
		}
		if opts.OutputFile != "" {
			jsonData, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return fmt.Errorf("序列化JSON失败: %w", err)
			}
			dir := filepath.Dir(opts.OutputFile)
			if dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("创建输出目录失败: %w", err)
				}
			}
			if err := os.WriteFile(opts.OutputFile, jsonData, 0644); err != nil {
				return fmt.Errorf("写入文件失败: %w", err)
			}
			fmt.Printf("已导出到: %s\n", opts.OutputFile)
			return nil
		}
		output.PrintJSON(data)
		return nil
	}

	if len(results) == 0 {
		fmt.Println("查询结果为空")
		return nil
	}

	fmt.Printf("返回 %d 行:\n\n", len(results))

	// 从第一行推断列名
	keys := make([]string, 0, len(results[0]))
	for k := range results[0] {
		keys = append(keys, k)
	}

	rows := make([][]string, 0, len(results))
	for _, row := range results {
		r := make([]string, 0, len(keys))
		for _, k := range keys {
			val := fmt.Sprintf("%v", row[k])
			val = strings.ReplaceAll(val, "\n", " ")
			r = append(r, val)
		}
		rows = append(rows, r)
	}
	output.PrintTable(keys, rows)
	return nil
}
