package server

import (
	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"
)

func CreateStopCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop-server",
		Short: "Stop the server if it's running",
		Run: func(cmd *cobra.Command, args []string) {
			vault.StopServer(cfg)
		},
	}

	return cmd
}
