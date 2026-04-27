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
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
)

// buildNameMapFromSQL 通过 SQL 批量查询文档 hpath，构建 ID→显示名 映射
func buildNameMapFromSQL(ctx context.Context, client *siyuan.Client, notebookID string) (map[string]string, error) {
	nameMap := make(map[string]string)
	var needContentIDs []string

	// 思源 SQL API 不带 LIMIT 时默认只返回 64 行，必须显式指定
	stmt := fmt.Sprintf(
		"SELECT id, hpath FROM blocks WHERE type = 'd' AND box = '%s' LIMIT 10000",
		strings.ReplaceAll(notebookID, "'", "''"),
	)

	rows, err := client.QuerySQL(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("SQL查询hpath失败: %w", err)
	}

	for _, row := range rows {
		id, _ := row["id"].(string)
		hpath, _ := row["hpath"].(string)
		if id == "" {
			continue
		}
		name := id
		if hpath != "" {
			parts := strings.Split(strings.Trim(hpath, "/"), "/")
			if len(parts) > 0 {
				last := parts[len(parts)-1]
				if last != "" {
					name = last
				}
			}
		}
		if looksLikeID(name) {
			needContentIDs = append(needContentIDs, id)
		}
		nameMap[id] = name
	}

	if len(needContentIDs) > 0 {
		fillNamesFromContent(ctx, client, needContentIDs, nameMap)
	}
	return nameMap, nil
}

func fillNamesFromContent(ctx context.Context, client *siyuan.Client, ids []string, nameMap map[string]string) {
	const batchSize = 64
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]

		quoted := make([]string, len(batch))
		for j, id := range batch {
			quoted[j] = fmt.Sprintf("'%s'", strings.ReplaceAll(id, "'", "''"))
		}
		stmt := fmt.Sprintf("SELECT id, content FROM blocks WHERE id IN (%s)", strings.Join(quoted, ","))

		rows, err := client.QuerySQL(ctx, stmt)
		if err != nil {
			continue
		}
		for _, row := range rows {
			id, _ := row["id"].(string)
			content, _ := row["content"].(string)
			if id == "" || content == "" {
				continue
			}
			if title := extractTitle(content); title != "" {
				nameMap[id] = title
			}
		}
	}
}

// looksLikeID 判断字符串是否看起来像 ID/日期而非人类可读名称
func looksLikeID(s string) bool {
	if len(s) <= 0 {
		return true
	}
	// 纯数字（如 20250223）或含连字符的 ID 格式
	for _, c := range s {
		if c >= '0' && c <= '9' || c == '-' {
			continue
		}
		return false
	}
	return true
}

// extractTitle 从文档内容中提取标题（第一行 # 开头的文本）
func extractTitle(content string) string {
	for line := range strings.SplitSeq(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			title := strings.TrimLeft(line, "# ")
			title = strings.TrimSpace(title)
			if title != "" {
				return title
			}
		}
	}
	return ""
}

type ListOptions struct {
	NotebookIdentifier string
	Path               string
	Sort               int
	Depth              int
	OutputFile         string
}

func ListDocuments(opts ListOptions) error {
	logger := log.GetLogger()
	logger.Info("开始列出文档树",
		zap.String("notebook", opts.NotebookIdentifier),
		zap.String("path", opts.Path),
		zap.Int("sort", opts.Sort),
		zap.Int("depth", opts.Depth),
		zap.String("output_file", opts.OutputFile))

	if !siyuan.IsSiYuanConfigValid() {
		fmt.Println("❌ 思源笔记配置无效或未启用")
		fmt.Println("请运行 'siyuan-cli auth login' 配置连接")
		return fmt.Errorf("思源笔记配置无效")
	}

	if strings.TrimSpace(opts.NotebookIdentifier) == "" {
		fmt.Println("❌ 错误: 笔记本标识符不能为空")
		fmt.Println("💡 使用方法: siyuan-cli document list <笔记本名称或ID>")
		fmt.Println("💡 使用 'siyuan-cli notebook list' 查看可用的笔记本")
		return fmt.Errorf("笔记本标识符不能为空")
	}

	if opts.OutputFile != "" && !output.IsJSON("") {
		fmt.Println("⚠️  警告: 导出到文件时强制使用JSON格式")
	}

	if opts.Sort < 0 || opts.Sort > 3 {
		err := fmt.Errorf("无效的排序方式: %d", opts.Sort)
		fmt.Printf("❌ 错误: %v\n", err)
		return err
	}

	client, err := siyuan.GetDefaultSiYuanClient()
	if err != nil {
		fmt.Printf("❌ 创建思源笔记客户端失败: %v\n", err)
		return fmt.Errorf("创建客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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

	docTree, err := client.ListDocTree(ctx, siyuan.ListDocTreeOptions{
		NotebookID: targetID,
		Path:       resolvedPath,
		Sort:       opts.Sort,
	})
	if err != nil {
		if syErr, ok := siyuan.IsAPIError(err); ok && syErr.Code == -1 && opts.Path != "" {
			fmt.Printf("'%s' 是一个文档，不是目录，无法列出子内容\n", opts.Path)
			fmt.Println("💡 提示: 使用上级目录路径查看子内容")
			return nil
		}
		fmt.Printf("❌ 获取文档树失败: %v\n", err)
		if syErr, ok := siyuan.IsAPIError(err); ok {
			fmt.Printf("❌ 思源笔记API错误 (code=%d): %s\n", syErr.Code, syErr.Msg)
		}
		return fmt.Errorf("获取文档树失败: %w", err)
	}

	nameMap, err := buildNameMapFromSQL(ctx, client, targetID)
	if err != nil {
		logger.Warn("SQL查询hpath失败，回退到tree name", zap.Error(err))
		nameMap = buildNameMapFromTree(docTree.Tree)
	}

	if output.IsJSON("") || opts.OutputFile != "" {
		return outputJSON(docTree, nameMap, targetName, opts)
	}
	return outputTree(docTree.Tree, nameMap, targetName, opts.Depth)
}

func ResolveDocPath(ctx context.Context, client *siyuan.Client, notebookID, notebookName, path string) (string, error) {
	if path == "" || path == "/" {
		return "/", nil
	}

	if IsDocID(path) {
		p := "/" + path
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		return p, nil
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")

	// 尝试跳过笔记本名前缀（UI 复制的完整 HPath 包含笔记本名）
	startIdx := 0
	if len(parts) > 1 && notebookName != "" && strings.EqualFold(parts[0], notebookName) {
		startIdx = 1
	}

	hPath := "/" + strings.Join(parts[startIdx:], "/")

	// 优先：直接尝试完整路径解析（高效且避免中间层级不存在的问题）
	if ids, err := client.GetIDsByHPath(ctx, siyuan.GetIDsByHPathOptions{
		Notebook: notebookID,
		Path:     hPath,
	}); err == nil && len(ids) > 0 {
		p := "/" + ids[len(ids)-1]
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		return p, nil
	}

	// 回退：逐段解析，构建完整的 ID 路径
	idPath := make([]string, 0, len(parts)-startIdx)
	for i := startIdx; i < len(parts); i++ {
		segPath := "/" + strings.Join(parts[startIdx:i+1], "/")
		ids, err := client.GetIDsByHPath(ctx, siyuan.GetIDsByHPathOptions{
			Notebook: notebookID,
			Path:     segPath,
		})
		if err != nil {
			return "", fmt.Errorf("未找到文档路径 '%s': %w", segPath, err)
		}
		if len(ids) == 0 {
			return "", fmt.Errorf("未找到文档路径 '%s'", segPath)
		}
		idPath = append(idPath, ids[len(ids)-1])
	}

	if len(idPath) == 0 {
		return "", fmt.Errorf("未找到文档路径 '%s'", path)
	}

	resolved := "/" + strings.Join(idPath, "/")
	if !strings.HasSuffix(resolved, "/") {
		resolved += "/"
	}
	return resolved, nil
}

func IsDocID(s string) bool {
	s = strings.Trim(s, "/")
	parts := strings.SplitN(s, "/", 2)
	id := parts[0]
	return len(id) >= 14 && strings.Contains(id, "-") && len(parts[0]) <= 24
}

func buildNameMapFromTree(tree []siyuan.DocTreeNode) map[string]string {
	nameMap := make(map[string]string)
	collectNamesFromTree(tree, "", nameMap)
	return nameMap
}

func collectNamesFromTree(nodes []siyuan.DocTreeNode, parentPath string, nameMap map[string]string) {
	for _, node := range nodes {
		var hpath string
		if node.Name != "" {
			hpath = node.Name
		} else if parentPath != "" {
			hpath = parentPath + "/" + node.ID[:8]
		} else {
			hpath = node.ID[:8]
		}
		nameMap[node.ID] = hpath
		if len(node.Children) > 0 {
			collectNamesFromTree(node.Children, hpath, nameMap)
		}
	}
}

func outputTree(nodes []siyuan.DocTreeNode, nameMap map[string]string, notebookName string, depth int) error {
	if len(nodes) == 0 {
		fmt.Println("暂无文档")
		return nil
	}

	totalDocs, totalDirs := 0, 0
	printTreeNodes(nodes, nameMap, "", true, depth, &totalDocs, &totalDirs)
	fmt.Printf("\n笔记本: %s  |  %d 个文档, %d 个目录\n", notebookName, totalDocs, totalDirs)
	return nil
}

func printTreeNodes(nodes []siyuan.DocTreeNode, nameMap map[string]string, prefix string, _ bool, depth int, totalDocs, totalDirs *int) {
	for i, node := range nodes {
		isLast := i == len(nodes)-1
		name := nameMap[node.ID]
		hasChildren := len(node.Children) > 0

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		childPrefix := "│   "
		if isLast {
			childPrefix = "    "
		}

		icon := "📄 "
		suffix := ""
		if hasChildren {
			icon = "📂 "
			suffix = fmt.Sprintf(" (%d)", len(node.Children))
			*totalDirs++
		} else {
			*totalDocs++
		}

			idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			fmt.Printf("%s%s%s%s%s  %s\n", prefix, connector, icon, name, suffix, idStyle.Render(node.ID))

		if hasChildren && depth != 1 {
			nextDepth := depth
			if depth > 1 {
				nextDepth = depth - 1
			}
			printTreeNodes(node.Children, nameMap, prefix+childPrefix, false, nextDepth, totalDocs, totalDirs)
		}
	}
}

type TreeNodeDisplay struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Children []TreeNodeDisplay `json:"children,omitempty"`
}

func outputJSON(docTree *siyuan.DocTreeData, nameMap map[string]string, notebookName string, opts ListOptions) error {
	displayTree := convertToDisplayTree(docTree.Tree, nameMap)
	data := map[string]any{
		"notebook": notebookName,
		"tree":     displayTree,
		"statistics": map[string]int{
			"totalNodes": countNodes(docTree.Tree),
		},
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

func convertToDisplayTree(nodes []siyuan.DocTreeNode, nameMap map[string]string) []TreeNodeDisplay {
	result := make([]TreeNodeDisplay, 0, len(nodes))
	for _, node := range nodes {
		display := TreeNodeDisplay{ID: node.ID, Name: nameMap[node.ID]}
		if len(node.Children) > 0 {
			display.Children = convertToDisplayTree(node.Children, nameMap)
		}
		result = append(result, display)
	}
	return result
}

func countNodes(nodes []siyuan.DocTreeNode) int {
	count := 0
	for _, node := range nodes {
		count++
		count += countNodes(node.Children)
	}
	return count
}

func FindNotebook(notebooks []siyuan.Notebook, identifier string) (string, string, error) {
	return findNotebookInternal(notebooks, identifier)
}

func findNotebookInternal(notebooks []siyuan.Notebook, identifier string) (string, string, error) {
	if len(identifier) >= 14 && strings.Contains(identifier, "-") {
		for _, nb := range notebooks {
			if nb.ID == identifier {
				if nb.Closed {
					return "", "", fmt.Errorf("笔记本 '%s' 已关闭，请先打开笔记本后再操作", nb.Name)
				}
				return nb.ID, nb.Name, nil
			}
		}
	}

	var matches []siyuan.Notebook
	identifierLower := strings.ToLower(identifier)

	for _, nb := range notebooks {
		if strings.EqualFold(nb.Name, identifier) {
			matches = append(matches, nb)
		}
	}
	if len(matches) == 1 {
		if matches[0].Closed {
			return "", "", fmt.Errorf("笔记本 '%s' 已关闭，请先打开笔记本后再操作", matches[0].Name)
		}
		return matches[0].ID, matches[0].Name, nil
	}
	if len(matches) > 1 {
		var names []string
		for _, nb := range matches {
			names = append(names, fmt.Sprintf("  %s (%s)", nb.Name, nb.ID))
		}
		return "", "", fmt.Errorf("找到多个匹配的笔记本：\n%s\n\n请使用更具体的名称或ID", strings.Join(names, "\n"))
	}

	for _, nb := range notebooks {
		if strings.Contains(strings.ToLower(nb.Name), identifierLower) {
			matches = append(matches, nb)
		}
	}
	if len(matches) >= 1 {
		if matches[0].Closed {
			return "", "", fmt.Errorf("笔记本 '%s' 已关闭，请先打开笔记本后再操作", matches[0].Name)
		}
		return matches[0].ID, matches[0].Name, nil
	}

	return "", "", fmt.Errorf("未找到匹配的笔记本: %s", identifier)
}
