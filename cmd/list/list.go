package list

import (
	"bitbucket.org/nordcloud/mfacli/config"
	"bitbucket.org/nordcloud/mfacli/pkg/vault"

	"fmt"

	"github.com/spf13/cobra"
)

func Create(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered client IDs",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clients, err := vault.ListClients(cfg)
			if err != nil {
				return err
			}

			for _, c := range clients {
				fmt.Println(c)
			}
			return nil
		},
	}
}
