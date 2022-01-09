package add

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/secret"
	"github.com/nordcloud/mfacli/pkg/vault"
)

const (
	clientFlag    = "client"
	secretFlag    = "secret"
	overwriteFlag = "overwrite"
)

func Create(cfg *config.Config) *cobra.Command {
	var (
		newSecret secret.SecretValue
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:   "add CLIENT_ID",
		Short: "Add a client with secret to the vault",
		Args:  cobra.ExactArgs(1),
		RunE: vault.RunOnVault(cfg, func(vlt vault.Vault, args ...string) error {
			clientId := args[0]

			return vlt.ModifySecrets(func(secrets map[string]string) error {
				if secret := secrets[clientId]; secret != "" && !overwrite {
					return fmt.Errorf("The client ID %s already exists in the vault. Pass --%s option to overwrite with new value.", clientId, overwriteFlag)
				}

				newSecretValue, err := newSecret.ReadSecret("TOTP secret: ", "Confirm TOTP secret")
				if err != nil {
					return err
				}

				secrets[clientId] = newSecretValue

				return nil
			})
		}),
	}

	cmd.Flags().VarP(&newSecret, secretFlag, "s", "Client secret")
	cmd.Flags().BoolVar(&overwrite, overwriteFlag, false, "Overwrite existing client ID")

	return cmd
}
