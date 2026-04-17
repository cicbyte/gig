package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 以下变量在编译时通过 -ldflags 注入
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func GetVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示 gig 版本号",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gig %s\n", Version)
			if GitCommit != "unknown" && len(GitCommit) >= 8 {
				fmt.Printf("commit: %s\n", GitCommit[:8])
			} else if GitCommit != "unknown" {
				fmt.Printf("commit: %s\n", GitCommit)
			}
			if BuildTime != "unknown" {
				fmt.Printf("built:  %s\n", BuildTime)
			}
		},
	}
}
