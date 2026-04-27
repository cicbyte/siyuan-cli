package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// TestCase YAML 测试用例结构
type TestCase struct {
	Name           string   `yaml:"name"`
	Skip           bool     `yaml:"skip"`
	RequiresServer bool     `yaml:"requires_server"`
	Command        []string `yaml:"command"`
	Cleanup        []string `yaml:"cleanup"` // 测试后清理的文件
	Expect         Expect   `yaml:"expect"`
}

// Expect 断言期望
type Expect struct {
	ExitCode         int      `yaml:"exit_code"`
	Contains         []string `yaml:"contains"`
	NotContains      []string `yaml:"not_contains"`
	ContainsStderr   []string `yaml:"contains_stderr"`
	JSONValid        bool     `yaml:"json_valid"`
	ContainsJSON     []string `yaml:"contains_json"`
	MatchCount       int      `yaml:"match_count"`
	MatchCountText   string   `yaml:"match_count_text"`
	OutputFileExists string   `yaml:"output_file_exists"`
}

// TestFile YAML 文件顶层结构
type TestFile struct {
	Name  string     `yaml:"name"`
	Tests []TestCase `yaml:"tests"`
}

// loadTestCases 从 YAML 文件加载测试用例
func loadTestCases(path string) (*TestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tf TestFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return nil, err
	}

	// 用文件名作为默认 name
	if tf.Name == "" {
		base := filepath.Base(path)
		tf.Name = strings.TrimSuffix(base, ".yaml")
	}

	return &tf, nil
}

// defaultVars 从环境变量加载模板变量
func defaultVars() map[string]string {
	vars := map[string]string{
		"notebook":  os.Getenv("TEST_NOTEBOOK"),
		"doc_path":  fixMSYSPath(os.Getenv("TEST_DOC_PATH")),
		"block_id":  os.Getenv("TEST_BLOCK_ID"),
	}
	return vars
}

// fixMSYSPath 反转 Git Bash MSYS 路径转换
// Git Bash 会将环境变量中的 /path 转换为 C:/Program Files/Git/path
// 此函数将转换后的路径还原为原始的 /path
func fixMSYSPath(s string) string {
	if s == "" {
		return s
	}
	// 检测 MSYS 转换模式: C:/Program Files/Git/xxx -> /xxx
	gitPrefix := "C:/Program Files/Git/"
	if strings.HasPrefix(s, gitPrefix) {
		return "/" + s[len(gitPrefix):]
	}
	return s
}

// applyVars 替换模板变量 {{var}}
func applyVars(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

// applyVarsSlice 替换字符串切片中的模板变量
func applyVarsSlice(args []string, vars map[string]string) []string {
	result := make([]string, len(args))
	for i, a := range args {
		result[i] = applyVars(a, vars)
	}
	return result
}

// TestAll 主测试入口：遍历 testdata/*.yaml 并执行
func TestAll(t *testing.T) {
	pattern := filepath.Join("testdata", "*.yaml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("读取测试数据目录失败: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("未找到任何 YAML 测试文件")
	}

	for _, file := range files {
		tf, err := loadTestCases(file)
		if err != nil {
			t.Errorf("加载 %s 失败: %v", file, err)
			continue
		}

		t.Run(tf.Name, func(t *testing.T) {
			for _, tc := range tf.Tests {
				t.Run(tc.Name, func(t *testing.T) {
					runTestCase(t, &tc)
				})
			}
		})
	}
}

func runTestCase(t *testing.T, tc *TestCase) {
	t.Helper()

	// 跳过标记
	if tc.Skip {
		t.Skip("测试用例标记为跳过")
	}

	// 需要 SiYuan 实例但未配置时跳过
	if tc.RequiresServer && os.Getenv("SIYUAN_CLI_HOME") == "" {
		t.Skip("需要 SiYuan 实例，未设置 SIYUAN_CLI_HOME")
	}

	// 检查模板变量是否都有值（在替换前检查）
	vars := defaultVars()
	for _, arg := range tc.Command {
		if strings.Contains(arg, "{{") {
			varName := extractVarName(arg)
			if val, ok := vars[varName]; !ok || val == "" {
				t.Skipf("缺少环境变量 TEST_%s，无法替换模板", strings.ToUpper(varName))
				return
			}
		}
	}

	// 应用模板变量
	command := applyVarsSlice(tc.Command, vars)

	// 解析命令中的输出文件路径（用于清理）
	outputFile := applyVars(extractOutputFile(command), vars)

	// 执行命令
	stdout, stderr, exitCode := cli(t, command...)

	// 退出码
	if tc.Expect.ExitCode != 0 {
		assertExitCode(t, exitCode, tc.Expect.ExitCode)
	}

	// stdout 包含检查
	if len(tc.Expect.Contains) > 0 {
		assertContains(t, stdout, tc.Expect.Contains)
	}

	// stdout 不包含检查
	if len(tc.Expect.NotContains) > 0 {
		assertNotContains(t, stdout, tc.Expect.NotContains)
	}

	// stderr 包含检查
	if len(tc.Expect.ContainsStderr) > 0 {
		assertContains(t, stderr, tc.Expect.ContainsStderr)
	}

	// JSON 有效性
	if tc.Expect.JSONValid {
		assertJSONValid(t, stdout)
	}

	// JSON 字段包含检查
	if len(tc.Expect.ContainsJSON) > 0 {
		assertJSONContains(t, stdout, tc.Expect.ContainsJSON)
	}

	// 匹配次数
	if tc.Expect.MatchCountText != "" {
		assertMatchCount(t, stdout, tc.Expect.MatchCountText, tc.Expect.MatchCount)
	}

	// 输出文件存在性
	if tc.Expect.OutputFileExists != "" {
		assertFileExists(t, tc.Expect.OutputFileExists)
	}

	// 清理
	t.Cleanup(func() {
		cleanupFiles := tc.Cleanup
		if outputFile != "" {
			cleanupFiles = append(cleanupFiles, outputFile)
		}
		for _, f := range cleanupFiles {
			os.Remove(f)
		}
	})
}

// extractOutputFile 从命令参数中提取 --output/-o 的值
func extractOutputFile(args []string) string {
	for i, a := range args {
		if (a == "--output" || a == "-o") && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// extractVarName 从模板字符串中提取变量名（如 "{{notebook}}" -> "notebook"）
func extractVarName(s string) string {
	start := strings.Index(s, "{{")
	if start == -1 {
		return ""
	}
	s = s[start+2:]
	end := strings.Index(s, "}}")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(s[:end])
}
