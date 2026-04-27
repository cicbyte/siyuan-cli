package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("siyuan-cli %s\n", Version)
			fmt.Printf("  commit: %s\n", GitCommit)
			fmt.Printf("  branch: %s\n", GitBranch)
			fmt.Printf("  built:  %s\n", BuildTime)
		},
	}
}
