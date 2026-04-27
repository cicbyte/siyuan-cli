package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

var globalFormat string

func SetFormat(f string) { globalFormat = f }

func GetFormat(override string) string {
	if override != "" {
		return override
	}
	return globalFormat
}

func IsJSON(override string) bool {
	return GetFormat(override) == "json"
}

func GetOutputFormat(override string) string {
	return GetFormat(override)
}

func PrintTable(headers []string, rows [][]string) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false

	colConfigs := make([]table.ColumnConfig, len(headers))
	for i := range headers {
		colConfigs[i] = table.ColumnConfig{Number: i + 1, WidthMax: 60}
	}
	t.SetColumnConfigs(colConfigs)
	t.AppendHeader(table.Row(fromSlice(headers)))
	for _, row := range rows {
		t.AppendRow(table.Row(fromSlice(row)))
	}
	fmt.Println(t.Render())
}

func PrintTableRight(headers []string, rows [][]string, rightCols ...int) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false

	rightSet := make(map[int]struct{}, len(rightCols))
	for _, c := range rightCols {
		rightSet[c] = struct{}{}
	}

	colConfigs := make([]table.ColumnConfig, len(headers))
	for i := range headers {
		cfg := table.ColumnConfig{Number: i + 1, WidthMax: 60}
		if _, ok := rightSet[i+1]; ok {
			cfg.Align = text.AlignRight
		}
		colConfigs[i] = cfg
	}
	t.SetColumnConfigs(colConfigs)
	t.AppendHeader(table.Row(fromSlice(headers)))
	for _, row := range rows {
		t.AppendRow(table.Row(fromSlice(row)))
	}
	fmt.Println(t.Render())
}

func fromSlice(s []string) table.Row {
	r := make(table.Row, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r
}

func Truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	runes := []rune(s)
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen-1]) + "…"
}

func PrintJSON(data any) {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON 编码失败: %v\n", err)
		return
	}
	fmt.Println(string(out))
}

func ReadPipeOrFile(filePath string) (string, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return "", nil
}

func GetTermSize() (int, int, error) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24, nil
	}
	return w, h, nil
}
