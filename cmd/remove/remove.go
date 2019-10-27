package remove

import (
	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"

	"github.com/spf13/cobra"
)

func Create(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "remove CLIENT_ID",
		Short: "Remove client ID from the vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return vault.RemoveClient(args[0], cfg)
		},
	}
}
