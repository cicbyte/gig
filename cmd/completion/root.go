package completion

import (
	"github.com/spf13/cobra"
)

func GetCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "生成 Shell 自动补全脚本",
		Long: `生成指定 Shell 的自动补全脚本。

使用方式:
  eval "$(gig completion bash)"
  eval "$(gig completion zsh)"
  gig completion fish > ~/.config/fish/completions/gig.fish
  gig completion powershell > gig.ps1`,
		Args:             cobra.ExactArgs(1),
		ValidArgs:        []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
