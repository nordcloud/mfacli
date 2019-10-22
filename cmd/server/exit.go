package server

import (
	"bitbucket.org/nordcloud/mfacli/config"
	"bitbucket.org/nordcloud/mfacli/pkg/vault"

	"github.com/spf13/cobra"
)

func CreateStopCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop-server",
		Short: "Stop the server if it's running",
		Run: func(cmd *cobra.Command, args []string) {
			vault.CloseServer(cfg)
		},
	}

	return cmd
}
