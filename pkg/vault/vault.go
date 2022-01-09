package vault

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
)

type Vault interface {
	GetSecrets() (map[string]string, error)
	ModifySecrets(func(map[string]string) error) error
}

type CobraFn func(*cobra.Command, []string) error
type VaultFn func(Vault, ...string) error

var (
	ErrClientNotFound = fmt.Errorf("Client ID not found")
)

func Open(cfg *config.Config) (Vault, error) {
	if cfg.NoCache {
		return openLocal(cfg)
	}

	return openRemote(cfg)
}

func RunOnVault(cfg *config.Config, fn VaultFn) CobraFn {
	return func(cmd *cobra.Command, args []string) error {
		vault, err := Open(cfg)
		if err != nil {
			return err
		}

		return fn(vault, args...)
	}
}
