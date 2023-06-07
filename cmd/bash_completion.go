package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func createBachCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bash_completion",
		Short: "Generate Bash-completion script",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Root().GenBashCompletion(os.Stdout)
		},
	}
}
