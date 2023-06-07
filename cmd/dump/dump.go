package dump

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"
)

func Create(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:    "dump-secrets-unencrypted",
		Short:  "Dump secrets stored in the vault in un-encrypted form to stdout (e.g. for backup purposes)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newCfg := *cfg
			newCfg.NoCache = true // always ask for password for this action
			vlt, err := vault.Open(&newCfg)
			if err != nil {
				return err
			}

			secrets, err := vlt.GetSecrets()
			if err != nil {
				return err
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(secrets)
		},
	}
}
