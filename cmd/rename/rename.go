package rename

import (
	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"

	"github.com/spf13/cobra"
)

func Create(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "rename OLD_CLIENT_ID NEW_CLIENT_ID",
		Short: "Rename the client",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return vault.RenameClient(args[0], args[1], cfg)
		},
	}
}
