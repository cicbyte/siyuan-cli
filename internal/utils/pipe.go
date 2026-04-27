package utils

import (
	"os"
)

// IsInputFromPipe 检查是否有管道输入
func IsInputFromPipe() bool {
	stat, _ := os.Stdin.Stat()
	// ModeCharDevice 表示终端输入，如果不是则可能是管道或文件重定向
	return (stat.Mode() & os.ModeCharDevice) == 0
}