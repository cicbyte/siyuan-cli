package block

import (
	"context"
	"fmt"
	"strings"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
)

func resolveDocID(ctx context.Context, client *siyuan.Client, notebookID, path string) (string, error) {
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return "", fmt.Errorf("获取笔记本列表失败: %w", err)
	}

	targetID, targetName, err := document.FindNotebook(notebooks, notebookID)
	if err != nil {
		return "", err
	}

	resolvedPath, err := document.ResolveDocPath(ctx, client, targetID, targetName, path)
	if err != nil {
		return "", err
	}
	docID := strings.TrimSuffix(strings.Trim(resolvedPath, "/"), "/")
	parts := strings.Split(docID, "/")
	return parts[len(parts)-1], nil
}

func findNotebook(notebooks []siyuan.Notebook, identifier string) (string, string, error) {
	return document.FindNotebook(notebooks, identifier)
}
