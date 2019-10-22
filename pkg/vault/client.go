package vault

import (
	"bitbucket.org/nordcloud/mfacli/config"

	"net/rpc"
	"os"
	"os/exec"
	"time"
)

func connect(cfg *config.Config, tryStart, create bool) (*rpc.Client, error) {
	connectFn := func() (*rpc.Client, error) {
		return rpc.Dial("unix", cfg.SocketPath)
	}

	c, err := connectFn()
	if err == nil {
		return c, nil
	}
	if !tryStart {
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
		c, err = connectFn()
		if time.Since(start) > 2*time.Second || err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return c, nil
}

func callServer(cfg *config.Config, tryStart, create bool, fn func(c *rpc.Client) error) error {
	c, err := connect(cfg, tryStart, create)
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
	c, _ := connect(cfg, false, false)
	if c != nil {
		c.Call("RemoteVault.Exit", struct{}{}, nil)
	}
}
