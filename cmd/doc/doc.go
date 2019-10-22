package doc

import (
	"os"

	"github.com/spf13/cobra"
	cobradoc "github.com/spf13/cobra/doc"
)

func Bind(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "markdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cobradoc.GenMarkdown(rootCmd, os.Stdout)
		},
	}
	rootCmd.AddCommand(cmd)

	return cmd
}
