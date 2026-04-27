package notebook

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cicbyte/siyuan-cli/internal/siyuan"
)

// isNotebookID 检查字符串是否符合思源笔记ID格式
// 思源笔记ID格式：时间戳-随机字符串，如 20231201120000-abc123def
func isNotebookID(s string) bool {
	// ID应该以数字开头（时间戳），包含连字符
	if len(s) < 10 {
		return false
	}

	// 检查是否以数字开头
	match, _ := regexp.MatchString(`^\d{8,}-\w+$`, s)
	return match
}

// FindNotebook 在笔记本列表中查找匹配的笔记本
func FindNotebook(notebooks []siyuan.Notebook, identifier string) (string, string, error) {
	identifier = strings.TrimSpace(identifier)

	// 首先检查是否为精确ID匹配
	if isNotebookID(identifier) {
		for _, nb := range notebooks {
			if nb.ID == identifier {
				if nb.Closed {
					return "", "", fmt.Errorf("笔记本 '%s' 已关闭，请先打开笔记本后再操作", nb.Name)
				}
				return nb.ID, nb.Name, nil
			}
		}
		return "", "", fmt.Errorf("不存在ID为 '%s' 的笔记本", identifier)
	}

	// 如果不是ID格式，进行名称匹配
	// 1. 精确名称匹配
	for _, nb := range notebooks {
		if nb.Name == identifier {
			if nb.Closed {
				return "", "", fmt.Errorf("笔记本 '%s' 已关闭，请先打开笔记本后再操作", nb.Name)
			}
			return nb.ID, nb.Name, nil
		}
	}

	// 2. 不区分大小写的名称匹配
	lowerIdentifier := strings.ToLower(identifier)
	for _, nb := range notebooks {
		if strings.ToLower(nb.Name) == lowerIdentifier {
			if nb.Closed {
				return "", "", fmt.Errorf("笔记本 '%s' 已关闭，请先打开笔记本后再操作", nb.Name)
			}
			return nb.ID, nb.Name, nil
		}
	}

	// 3. 包含匹配（不区分大小写）
	var matches []siyuan.Notebook
	for _, nb := range notebooks {
		if strings.Contains(strings.ToLower(nb.Name), lowerIdentifier) {
			matches = append(matches, nb)
		}
	}

	if len(matches) == 1 {
		if matches[0].Closed {
			return "", "", fmt.Errorf("笔记本 '%s' 已关闭，请先打开笔记本后再操作", matches[0].Name)
		}
		return matches[0].ID, matches[0].Name, nil
	} else if len(matches) > 1 {
		// 多个匹配项，提供选择
		fmt.Printf("🔍 找到多个匹配的笔记本：\n")
		for i, nb := range matches {
			status := "关闭"
			if !nb.Closed {
				status = "打开"
			}
			fmt.Printf("  %d. %s (%s) - %s\n", i+1, nb.Name, nb.ID, status)
		}
		return "", "", fmt.Errorf("找到多个匹配的笔记本，请使用更精确的名称或ID")
	}

	// 没有找到匹配项
	return "", "", fmt.Errorf("不存在名称为 '%s' 的笔记本", identifier)
}