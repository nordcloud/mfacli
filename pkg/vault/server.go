package vault

import (
	"github.com/nordcloud/mfacli/config"

	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type RemoteVault struct {
	vault *Vault
	lis   net.Listener
}

type AddClientInput struct {
	ClientId string
	Secret   string
}

func (r *RemoteVault) GetSecret(clientId string, secretP *string) error {
	secret, err := r.vault.getSecret(clientId)
	if err != nil {
		return err
	}
	*secretP = secret
	return nil
}

func (r *RemoteVault) AddClient(input AddClientInput, output *bool) error {
	return r.vault.addClient(input.ClientId, input.Secret)
}

func (r *RemoteVault) ListClients(input struct{}, output *[]string) error {
	clients := r.vault.listClients()
	*output = clients
	return nil
}

func (r *RemoteVault) RemoveClient(clientId string, output *bool) error {
	return r.vault.removeClient(clientId)
}

func (r *RemoteVault) Exit(input struct{}, output *bool) error {
	r.lis.Close()
	return nil
}

func RunServer(cfg *config.Config) error {
	key, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	vault, err := newVault(cfg, false, key)
	if err != nil {
		return err
	}

	lis, err := net.Listen("unix", cfg.SocketPath)
	if err != nil {
		return err
	}

	defer func(lis net.Listener) {
		lis.Close()
	}(lis)

	go handleSignals(lis)

	err = rpc.Register(&RemoteVault{
		lis:   lis,
		vault: vault,
	})
	if err != nil {
		log.WithError(err).Error("Failed to register remote object")
		return err
	}
	rpc.Accept(lis)
	return nil
}

func handleSignals(lis net.Listener) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Info("Caught signal, closing server")
	lis.Close()
	os.Exit(0)
}
