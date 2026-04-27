package synccmd

import (
	"context"
	"fmt"
	"time"

	"github.com/cicbyte/siyuan-cli/internal/log"
	"github.com/cicbyte/siyuan-cli/internal/output"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
)

func SyncStatus() error {
	logger := log.GetLogger()
	logger.Info("获取同步状态")

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.GetSyncInfo(ctx)
	if err != nil {
		fmt.Printf("❌ 获取同步信息失败: %v\n", err)
		return err
	}

	if output.IsJSON("") {
		output.PrintJSON(info)
		return nil
	}

	fmt.Printf("同步状态:\n")
	fmt.Printf("  已同步文件: %d\n", info.Synced)
	fmt.Printf("  冲突文件: %d\n", info.Conflict)
	fmt.Printf("  同步大小: %s\n", formatSize(info.SyncSize))
	fmt.Printf("  最后同步: %s\n", info.LastSync)
	return nil
}

func SyncNow() error {
	logger := log.GetLogger()
	logger.Info("执行同步")

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	fmt.Println("正在同步...")
	err = client.PerformSync(ctx)
	if err != nil {
		fmt.Printf("❌ 同步失败: %v\n", err)
		return err
	}

	fmt.Println("✅ 同步完成")
	return nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
