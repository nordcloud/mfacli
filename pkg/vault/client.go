package vault

import (
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/ncerrors/errors"
)

func connectBasic(cfg *config.Config) (*rpc.Client, error) {
	return rpc.Dial("unix", cfg.SocketPath)
}

func connect(cfg *config.Config, create bool) (*rpc.Client, error) {
	c, err := connectBasic(cfg)
	if err == nil {
		return c, nil
	}
	if err := handleConnectError(err, cfg.SocketPath); err != nil {
		return nil, err
	}

	vlt, err := newVault(cfg, create, nil)
	if err != nil {
		return nil, err
	}

	err = startServer(cfg, vlt)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	for {
		c, err = connectBasic(cfg)
		if time.Since(start) > 2*time.Second || err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return c, nil
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
			return errors.WithContext(connErr, "not a socket", errors.Fields{
				"socket_path": sockPath,
			})
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

func callServer(cfg *config.Config, create bool, fn func(c *rpc.Client) error) error {
	c, err := connect(cfg, create)
	if err != nil {
		return err
	}
	defer c.Close()

	return fn(c)
}

func getExecutableName() string {
	path, err := os.Readlink("/proc/self/exe")
	if err == nil {
		return path
	}
	return os.Args[0]
}

func startServer(cfg *config.Config, vlt *Vault) error {
	progname := getExecutableName()
	args := []string{config.InternalRunServerCmd,
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

	_, err = pipe.Write(vlt.EncKey)
	if err != nil {
		return err
	}
	err = pipe.Close()
	if err != nil {
		return err
	}

	return nil
}

func StartServer(cfg *config.Config) error {
	vlt, err := newVault(cfg, true, nil)
	if err != nil {
		return err
	}
	return startServer(cfg, vlt)
}

func CloseServer(cfg *config.Config) {
	c, _ := connectBasic(cfg)
	if c != nil {
		c.Call("RemoteVault.Exit", struct{}{}, nil)
	}
}
