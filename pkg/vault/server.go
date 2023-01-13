package vault

import (
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/codec"
)

const (
	serverName = "VaultServer"
)

type VaultServer struct {
	vault *localVault
	lis   net.Listener
}

func (s *VaultServer) GetSecrets(input struct{}, secrets *map[string]string) error {
	*secrets = s.vault.secrets
	return nil
}

func (s *VaultServer) StoreSecrets(secrets map[string]string, output *struct{}) error {
	s.vault.secrets = secrets
	return s.vault.save()
}

func (s *VaultServer) Stop(input struct{}, output *struct{}) error {
	s.lis.Close()
	return nil
}

func RunServer(cfg *config.Config) error {
	key, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	vault, err := openLocalWithKey(cfg.VaultPath, key)
	if err != nil {
		return err
	}

	lis, err := net.Listen("unix", cfg.SocketPath)
	if err != nil {
		return err
	}

	defer lis.Close()
	go handleSignals(lis)

	err = rpc.Register(&VaultServer{
		lis:   lis,
		vault: vault,
	})
	if err != nil {
		return err
	}
	rpc.Accept(lis)
	return nil
}

func openLocalWithKey(path string, key []byte) (*localVault, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
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
		path:    path,
	}, nil
}

func handleSignals(lis net.Listener) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	s := <-c
	log.Infof("Caught the %s signal, closing server", s.String())
	lis.Close()
	os.Exit(0)
}
