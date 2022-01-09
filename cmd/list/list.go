package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"
)

func Create(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered client IDs",
		Args:  cobra.ExactArgs(0),
		RunE: vault.RunOnVault(cfg, func(vlt vault.Vault, args ...string) error {
			secrets, err := vlt.GetSecrets()
			if err != nil {
				return err
			}

			for c := range secrets {
				fmt.Println(c)
			}
			return nil
		}),
	}
}
