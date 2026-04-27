package query

import (
	"fmt"

	"github.com/cicbyte/siyuan-cli/internal/logic/query"
	"github.com/spf13/cobra"
)

var queryOutput string
var queryRaw bool

func GetQueryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query <sql>",
		Short: "执行 SQL 查询",
		Long: `直接执行 SQL 查询思源笔记数据库。

示例：
  siyuan-cli query "SELECT * FROM blocks WHERE type='d' LIMIT 10"
  siyuan-cli query "SELECT id, content FROM blocks WHERE content LIKE '%TODO%'" -o todos.json
  siyuan-cli query "SELECT * FROM blocks" --raw`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts := query.QueryOptions{
				SQL:        args[0],
				OutputFile: queryOutput,
				Raw:        queryRaw,
			}
			if err := query.RunQuery(opts); err != nil {
				fmt.Printf("❌ 命令执行失败: %v\n", err)
			}
		},
	}

	cmd.Flags().StringVarP(&queryOutput, "output", "o", "", "导出到文件")
	cmd.Flags().BoolVar(&queryRaw, "raw", false, "输出原始 JSON（不格式化表格）")
	return cmd
}
