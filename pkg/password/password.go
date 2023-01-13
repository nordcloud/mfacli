package password

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/nordcloud/mfacli/config"
	"github.com/pkg/errors"
)

var (
	ErrCancelled = fmt.Errorf("Cancelled")
)

func ReadPassword(cfg *config.Config, prompt string) (string, error) {
	if cfg.PasswordCommand == "" {
		return readTerm(prompt)
	}

	return readCommand(cfg.PasswordCommand, prompt)
}

func CreatePassword(cfg *config.Config) (string, error) {
	pwd, err := ReadPassword(cfg, "Set up a new password")
	if err != nil {
		return "", err
	}

	var confirmation string
	prompt := "Repeat the password"
	for confirmation != pwd {
		confirmation, err = ReadPassword(cfg, prompt)
		if err != nil {
			return "", err
		}
		prompt = "Passwords don't match, try again"
	}

	return pwd, nil
}

func readTerm(prompt string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", prompt)
	data, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprint(os.Stderr, "\n")
	if err != nil {
		return "", errors.Wrap(err, "reading password from terminal")
	}

	if len(data) == 0 {
		return "", ErrCancelled
	}

	return string(data), nil
}

func readCommand(cmd, prompt string) (string, error) {
	execCmd := exec.Command(cmd, prompt)
	out, err := execCmd.Output()
	if err != nil {
		if exerr, ok := err.(*exec.ExitError); ok {
			if exerr.ExitCode() == 1 {
				return "", ErrCancelled
			}
		}
		return "", errors.Wrapf(err, "reading password with %s", cmd)
	}

	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}

	if len(out) == 0 {
		return "", ErrCancelled
	}

	return string(out), nil
}
