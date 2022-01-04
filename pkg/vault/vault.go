package vault

import (
	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/codec"
	"github.com/nordcloud/mfacli/pkg/secret"

	"fmt"
	"io/ioutil"
	"net/rpc"
	"os"
	"sort"
)

type Vault struct {
	Secrets map[string]string
	EncKey  []byte
	Path    string
}

const (
	ErrClientNotFound      = "Client ID not found"
	ErrClientAlreadyExists = "Client already exists"
)

func newVault(cfg *config.Config, create bool, key []byte) (*Vault, error) {
	vaultData, err := ioutil.ReadFile(cfg.VaultPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	if key == nil {
		key, err = buildKey(cfg.Password, create && vaultData == nil)
	}

	secrets := map[string]string{}
	if vaultData != nil {
		secrets, err = codec.Decrypt(vaultData, key)
		if err != nil {
			return nil, err
		}
	}

	vlt := &Vault{
		EncKey:  key,
		Secrets: secrets,
		Path:    cfg.VaultPath,
	}

	return vlt, nil
}

func buildKey(password secret.SecretValue, create bool) ([]byte, error) {
	var pswd string
	var err error

	if password.IsSet() {
		pswd = password.String()
	} else if create {
		pswd, err = secret.ReadSecret("Set up new password: ", "Repeat the password: ")
	} else {
		pswd, err = secret.ReadSecret("Password: ", "")
	}

	if err != nil {
		return nil, err
	}

	return codec.BuildEncKey(pswd), nil
}

func (v *Vault) getSecret(clientId string) (string, error) {
	secret, ok := v.Secrets[clientId]
	if !ok {
		return "", fmt.Errorf(ErrClientNotFound)
	}
	return secret, nil
}

func (v *Vault) addClient(clientId, secret string) error {
	v.Secrets[clientId] = secret
	return v.save()
}

func (v *Vault) listClients() (clients []string) {
	for c := range v.Secrets {
		clients = append(clients, c)
	}
	sort.Strings(clients)
	return
}

func (v *Vault) removeClient(clientId string) error {
	delete(v.Secrets, clientId)
	return v.save()
}

func (v *Vault) renameClient(oldClientId, newClientId string) error {
	secret, ok := v.Secrets[oldClientId]
	if !ok {
		return fmt.Errorf(ErrClientNotFound)
	}

	if _, ok := v.Secrets[newClientId]; ok {
		return fmt.Errorf(ErrClientAlreadyExists)
	}

	delete(v.Secrets, oldClientId)
	v.Secrets[newClientId] = secret

	return v.save()
}

func (v *Vault) save() error {
	vaultData, err := codec.Encrypt(v.Secrets, v.EncKey)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(v.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(vaultData)
	f.Close()
	return err
}

func GetSecret(clientId string, cfg *config.Config) (string, error) {
	if cfg.NoCache {
		v, err := newVault(cfg, false, nil)
		if err != nil {
			return "", err
		}
		return v.getSecret(clientId)
	}

	var secret string
	err := callServer(cfg, false, func(c *rpc.Client) error {
		return c.Call("RemoteVault.GetSecret", clientId, &secret)
	})
	if err != nil {
		return "", err
	}
	return secret, nil
}

func AddClient(clientId string, newSecret *secret.SecretValue, cfg *config.Config) error {
	getSecretFn := func() (string, error) {
		if !newSecret.IsSet() {
			return secret.ReadSecret("TOTP secret: ", "")
		}
		return newSecret.String(), nil
	}

	if cfg.NoCache {
		v, err := newVault(cfg, true, nil)
		if err != nil {
			return err
		}

		secret, err := getSecretFn()
		if err != nil {
			return err
		}

		return v.addClient(clientId, secret)
	}

	secret, err := getSecretFn()
	if err != nil {
		return err
	}

	return callServer(cfg, true, func(c *rpc.Client) error {
		return c.Call("RemoteVault.AddClient", AddClientInput{clientId, secret}, nil)
	})
}

func ListClients(cfg *config.Config) ([]string, error) {
	if cfg.NoCache {
		v, err := newVault(cfg, false, nil)
		if err != nil {
			return nil, err
		}
		return v.listClients(), nil
	}

	var clients []string
	err := callServer(cfg, false, func(c *rpc.Client) error {
		return c.Call("RemoteVault.ListClients", struct{}{}, &clients)
	})
	if err != nil {
		return nil, err
	}
	return clients, nil
}

func RemoveClient(clientId string, cfg *config.Config) error {
	if cfg.NoCache {
		v, err := newVault(cfg, false, nil)
		if err != nil {
			return err
		}
		return v.removeClient(clientId)
	}

	err := callServer(cfg, false, func(c *rpc.Client) error {
		return c.Call("RemoteVault.RemoveClient", clientId, nil)
	})
	if err != nil {
		return err
	}

	return nil
}

func RenameClient(oldClientId, newClientId string, cfg *config.Config) error {
	if cfg.NoCache {
		v, err := newVault(cfg, false, nil)
		if err != nil {
			return err
		}
		return v.renameClient(oldClientId, newClientId)
	}

	return callServer(cfg, false, func(c *rpc.Client) error {
		request := RenameClientRequest{
			Old: oldClientId,
			New: newClientId,
		}
		return c.Call("RemoteVault.RenameClient", request, nil)
	})
}

func GetSecrets(cfg *config.Config) (map[string]string, error) {
	if cfg.NoCache {
		v, err := newVault(cfg, false, nil)
		if err != nil {
			return nil, err
		}
		return v.Secrets, nil
	}

	var secrets map[string]string
	err := callServer(cfg, false, func(c *rpc.Client) error {
		return c.Call("RemoteVault.GetSecrets", struct{}{}, &secrets)
	})
	if err != nil {
		return nil, err
	}

	return secrets, nil
}
