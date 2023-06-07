package list

import (
	"fmt"
	"sort"

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

			names := make([]string, 0, len(secrets))
			for name := range secrets {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				fmt.Println(name)
			}

			return nil
		}),
	}
}
