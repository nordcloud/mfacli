package add

import (
	"bitbucket.org/nordcloud/mfacli/config"
	"bitbucket.org/nordcloud/mfacli/pkg/secret"
	"bitbucket.org/nordcloud/mfacli/pkg/vault"

	"fmt"

	"github.com/spf13/cobra"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			clientId := args[0]

			if !overwrite {
				_, err := vault.GetSecret(clientId, cfg)
				if err == nil {
					return fmt.Errorf("The client ID %s already exists in the vault. Pass --%s option to overwrite with new value.", clientId, overwriteFlag)
				}
				if err.Error() != vault.ErrClientNotFound {
					return err
				}
			}

			return vault.AddClient(clientId, &newSecret, cfg)
		},
	}

	cmd.Flags().VarP(&newSecret, secretFlag, "s", "Client secret")
	cmd.Flags().BoolVar(&overwrite, overwriteFlag, false, "Overwrite existing client ID")

	return cmd
}
