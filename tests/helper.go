package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// findBinary 查找 siyuan-cli 二进制文件路径
func findBinary(t *testing.T) string {
	t.Helper()

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	// 项目根目录（tests/ 的上级）
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal("获取工作目录失败")
	}
	projectRoot := filepath.Join(dir, "..")

	// 1. SIYUAN_CLI_BIN 环境变量（支持相对路径，相对于项目根目录解析）
	if env := os.Getenv("SIYUAN_CLI_BIN"); env != "" {
		candidate := env
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(projectRoot, candidate)
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		// 尝试原值（可能已经是绝对路径）
		if _, err := os.Stat(env); err == nil {
			return env
		}
	}

	// 2. 默认：项目根目录下的 siyuan-cli[.exe]
	bin := filepath.Join(projectRoot, "siyuan-cli"+ext)
	if _, err := os.Stat(bin); err == nil {
		return bin
	}

	t.Fatalf("未找到 siyuan-cli 二进制。请先运行 go build -o siyuan-cli%s . 或设置 SIYUAN_CLI_BIN", ext)
	return ""
}

// cli 运行 siyuan-cli 二进制，返回 stdout、stderr、exitCode
func cli(t *testing.T, args ...string) (string, string, int) {
	t.Helper()

	bin := findBinary(t)

	cmd := exec.Command(bin, args...)

	// 使用独立配置目录，避免干扰用户数据
	cmd.Env = os.Environ()
	if home := os.Getenv("SIYUAN_CLI_HOME"); home != "" {
		cmd.Env = append(cmd.Env, "SIYUAN_CLI_HOME="+home)
	}
	// 防止 Git Bash MSYS 将 /path 参数转换为 Windows 路径
	cmd.Env = append(cmd.Env, "MSYS_NO_PATHCONV=1")

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("执行命令失败: %v", err)
		}
	}

	return stdout.String(), stderr.String(), exitCode
}

// assertExitCode 验证退出码
func assertExitCode(t *testing.T, actual, expected int) {
	t.Helper()
	if actual != expected {
		t.Errorf("退出码: 期望 %d, 实际 %d", expected, actual)
	}
}

// assertContains 验证输出包含指定文本（任一匹配即可）
func assertContains(t *testing.T, output string, substrs []string) {
	t.Helper()
	for _, s := range substrs {
		if strings.Contains(output, s) {
			return
		}
	}
	t.Errorf("输出中未找到以下任一文本: %v\n输出: %s", substrs, truncate(output, 200))
}

// assertNotContains 验证输出不包含指定文本
func assertNotContains(t *testing.T, output string, substrs []string) {
	t.Helper()
	for _, s := range substrs {
		if strings.Contains(output, s) {
			t.Errorf("输出中不应包含: %q", s)
		}
	}
}

// assertJSONValid 验证输出是合法 JSON
func assertJSONValid(t *testing.T, output string) {
	t.Helper()
	trimmed := strings.TrimSpace(output)
	if !json.Valid([]byte(trimmed)) {
		t.Errorf("输出不是合法 JSON: %s", truncate(output, 200))
	}
}

// assertJSONContains 验证 JSON 输出中包含指定字段
func assertJSONContains(t *testing.T, output string, fields []string) {
	t.Helper()
	assertJSONValid(t, output)

	var data map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &data); err != nil {
		t.Errorf("JSON 解析失败: %v", err)
		return
	}

	for _, field := range fields {
		if _, ok := data[field]; !ok {
			t.Errorf("JSON 中缺少字段 %q, 已有字段: %v", field, keys(data))
		}
	}
}

// assertMatchCount 验证文本匹配次数
func assertMatchCount(t *testing.T, output string, substr string, expected int) {
	t.Helper()
	count := strings.Count(output, substr)
	if count != expected {
		t.Errorf("文本 %q 出现次数: 期望 %d, 实际 %d", substr, expected, count)
	}
}

// assertFileExists 验证文件存在
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("文件不存在: %s", path)
	}
}

// keys 获取 map 的所有 key
func keys(m map[string]any) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// truncate 截断字符串用于错误信息
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen-3]) + "..."
}
