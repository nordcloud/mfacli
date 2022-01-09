package vault

import (
	"io/ioutil"
	"os"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/codec"
)

type localVault struct {
	secrets map[string]string
	encKey  []byte
	path    string
}

func (v *localVault) GetSecrets() (map[string]string, error) {
	secrets := make(map[string]string, len(v.secrets))
	for k, v := range v.secrets {
		secrets[k] = v
	}
	return secrets, nil
}

func (v *localVault) ModifySecrets(modify func(map[string]string) error) error {
	if err := modify(v.secrets); err != nil {
		return err
	}

	return v.save()
}

func (v *localVault) save() error {
	vaultData, err := codec.Encrypt(v.secrets, v.encKey)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(v.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(vaultData)
	return err
}

func openLocal(cfg *config.Config) (*localVault, error) {
	vaultFile, err := os.Open(cfg.VaultPath)
	if err == nil {
		defer vaultFile.Close()

		password, err := cfg.Password.ReadSecret("Password: ", "")
		if err != nil {
			return nil, err
		}

		return loadVaultFile(vaultFile, codec.BuildEncKey(password))
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	password, err := cfg.Password.ReadSecret("Set up new password: ", "Repeat the password: ")
	if err != nil {
		return nil, err
	}

	vault := &localVault{
		secrets: make(map[string]string),
		encKey:  codec.BuildEncKey(password),
		path:    cfg.VaultPath,
	}

	if err := vault.save(); err != nil {
		return nil, err
	}

	return vault, nil
}

func loadVaultFile(file *os.File, key []byte) (*localVault, error) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	secrets, err := codec.Decrypt(data, key)
	if err != nil {
		return nil, err
	}

	return &localVault{
		secrets: secrets,
		encKey:  key,
		path:    file.Name(),
	}, nil
}
