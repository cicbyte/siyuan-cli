package export

import (
	"context"
	"strings"

	"github.com/cicbyte/siyuan-cli/internal/logic/document"
	"github.com/cicbyte/siyuan-cli/internal/siyuan"
)

func resolveDocID(ctx context.Context, client *siyuan.Client, notebookID, path string) (string, error) {
	notebooks, err := client.ListNotebooks(ctx)
	if err != nil {
		return "", err
	}

	targetID, targetName, err := document.FindNotebook(notebooks, notebookID)
	if err != nil {
		return "", err
	}

	resolvedPath, err := document.ResolveDocPath(ctx, client, targetID, targetName, path)
	if err != nil {
		return "", err
	}

	docID := resolvedPath
	if len(docID) > 0 && docID[len(docID)-1] == '/' {
		docID = docID[:len(docID)-1]
	}
	idx := strings.LastIndex(docID, "/")
	if idx >= 0 {
		return docID[idx+1:], nil
	}
	return docID, nil
}
