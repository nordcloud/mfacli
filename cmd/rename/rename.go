package rename

import (
	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"
)

func Create(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "rename OLD_CLIENT_ID NEW_CLIENT_ID",
		Short: "Rename the client",
		Args:  cobra.ExactArgs(2),
		RunE: vault.RunOnVault(cfg, func(vlt vault.Vault, args ...string) error {
			return vlt.ModifySecrets(func(secrets map[string]string) error {
				old, new := args[0], args[1]

				secret := secrets[old]
				if secret == "" {
					return vault.ErrClientNotFound
				}

				secrets[new] = secret
				delete(secrets, old)
				return nil
			})
		}),
	}
}
