package vault

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/codec"
	"github.com/nordcloud/mfacli/pkg/password"
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
	vault, err := readVaultFile(cfg)
	if err == nil {
		return vault, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	pwd, err := password.CreatePassword(cfg)
	if err != nil {
		return nil, err
	}

	vault = &localVault{
		secrets: make(map[string]string),
		encKey:  codec.BuildEncKey(pwd),
		path:    cfg.VaultPath,
	}
	if err := vault.save(); err != nil {
		return nil, err
	}

	return vault, nil
}

func readVaultFile(cfg *config.Config) (*localVault, error) {
	file, err := os.Open(cfg.VaultPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	pwd, err := password.ReadPassword(cfg, "Password")
	if err != nil {
		return nil, err
	}
	secrets, err := codec.Decrypt(data, codec.BuildEncKey(pwd))
	for errors.Is(err, codec.ErrInvalidPassword) {
		pwd, err = password.ReadPassword(cfg, "Invalid password. Try again")
		if err != nil {
			return nil, err
		}
		secrets, err = codec.Decrypt(data, codec.BuildEncKey(pwd))
	}
	if err != nil {
		return nil, err
	}

	return &localVault{
		secrets: secrets,
		encKey:  codec.BuildEncKey(pwd),
		path:    file.Name(),
	}, nil
}
