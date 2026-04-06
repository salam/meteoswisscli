package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for whiterisk.

Bash:
  $ source <(whiterisk completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ whiterisk completion bash > /etc/bash_completion.d/whiterisk
  # macOS:
  $ whiterisk completion bash > $(brew --prefix)/etc/bash_completion.d/whiterisk

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ whiterisk completion zsh > "${fpath[1]}/_whiterisk"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ whiterisk completion fish | source

  # To load completions for each session, execute once:
  $ whiterisk completion fish > ~/.config/fish/completions/whiterisk.fish
`,
	Example: `  whiterisk completion zsh
  source <(whiterisk completion bash)`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}
