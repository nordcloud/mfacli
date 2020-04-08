package dump

import (
	"encoding/json"
	"os"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"

	"github.com/spf13/cobra"
)

func Create(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "dump-secrets-unencrypted",
		Short: "Dump secrets stored in the vault in un-encrypted form to stdout (e.g. for backup purposes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := vault.GetSecrets(cfg)
			if err != nil {
				return err
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(secrets)
		},
	}
}
