package vault

import (
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/nordcloud/mfacli/config"
)

type remoteVault struct {
	client *rpc.Client
}

func (v *remoteVault) GetSecrets() (map[string]string, error) {
	var secrets map[string]string
	if err := v.client.Call(serverName+".GetSecrets", struct{}{}, &secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

func (v *remoteVault) ModifySecrets(modify func(map[string]string) error) error {
	secrets, err := v.GetSecrets()
	if err != nil {
		return err
	}

	if err := modify(secrets); err != nil {
		return err
	}

	if err := v.client.Call(serverName+".StoreSecrets", secrets, nil); err != nil {
		return err
	}

	return nil
}

func StartServer(cfg *config.Config) error {
	vault, err := openRemote(cfg)
	if err != nil {
		return err
	}

	if err := vault.client.Close(); err != nil {
		log.WithError(err).Info("Failed to close temporary client")
	}
	return nil
}

func StopServer(cfg *config.Config) {
	client, _ := connect(cfg)
	if client != nil {
		client.Call(serverName+".Stop", struct{}{}, nil)
	}
}

func openRemote(cfg *config.Config) (*remoteVault, error) {
	client, err := connect(cfg)
	if err == nil {
		return &remoteVault{client: client}, nil
	}

	if err := handleConnectError(err, cfg.SocketPath); err != nil {
		return nil, err
	}

	vault, err := openLocal(cfg)
	if err != nil {
		return nil, err
	}
	if err := startServer(cfg, vault.encKey); err != nil {
		return nil, err
	}

	start := time.Now()
	for {
		client, err = connect(cfg)
		if time.Since(start) > 2*time.Second || err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return &remoteVault{client: client}, nil
}

func connect(cfg *config.Config) (*rpc.Client, error) {
	return rpc.Dial("unix", cfg.SocketPath)
}

func handleConnectError(connErr error, sockPath string) error {
	opErr, ok := connErr.(*net.OpError)
	if !ok {
		return connErr
	}
	netErr, ok := opErr.Unwrap().(*os.SyscallError)
	if !ok {
		return connErr
	}

	switch netErr.Unwrap() {
	case syscall.ECONNREFUSED:
		// The socket file exists but either not bound, or is not a socket
		s, err := os.Stat(sockPath)
		if err != nil {
			log.WithError(err).WithField("socket_path", sockPath).Debug("Failed to check socket file")
			return connErr
		}
		if s.Mode().Type()&os.ModeSocket == 0 {
			return errors.Wrapf(connErr, "not a socket: %s", sockPath)
		}

		log.WithField("socket_path", sockPath).Debug("Socket exists, but not bound, removing")
		if err := os.Remove(sockPath); err != nil {
			log.WithField("socket_path", sockPath).WithError(err).Debug("Failed to remove orphaned socket file")
			return connErr
		}

		return nil
	case syscall.ENOENT:
		log.Debug("No socket file found, starting the server")
		return nil
	}

	return connErr
}

func startServer(cfg *config.Config, key []byte) error {
	progname := getExecutableName()
	args := []string{
		config.InternalRunServerCmd,
		"--socket", cfg.SocketPath,
		"--vault", cfg.VaultPath,
	}
	if cfg.ServerLogFile != "" {
		args = append(args, "--"+config.FlagServerLogFile, cfg.ServerLogFile)
	}
	cmd := exec.Command(progname, args...)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	_, err = pipe.Write(key)
	if err != nil {
		return err
	}
	err = pipe.Close()
	if err != nil {
		return err
	}

	return nil
}

func getExecutableName() string {
	path, err := os.Readlink("/proc/self/exe")
	if err == nil {
		return path
	}
	return os.Args[0]
}
