package remove

import (
	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"
)

func Create(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "remove CLIENT_ID",
		Short: "Remove client ID from the vault",
		Args:  cobra.ExactArgs(1),
		RunE: vault.RunOnVault(cfg, func(vlt vault.Vault, args ...string) error {
			return vlt.ModifySecrets(func(secrets map[string]string) error {
				delete(secrets, args[0])
				return nil
			})
		}),
	}
}
