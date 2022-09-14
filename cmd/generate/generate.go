package generate

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"
)

const (
	newLineFlag = "newline"

	xdotoolCmd = "xdotool"

	xselCmd            = "xsel"
	xselTargetsFlag    = "xsel-targets"
	defaultXselTargets = "primary,clipboard"

	pbcopyCmd = "pbcopy"
)

func CreateTypeCmd(cfg *config.Config) *cobra.Command {
	return createGenerateCmd(cfg, "type", "Simulate typing of the TOTP code", func(code string, newLine bool) error {
		args := []string{"type", "--clearmodifiers", code}
		log.WithField("args", args).Debug("Running " + xdotoolCmd)
		cmd := exec.Command(xdotoolCmd, args...)
		if err := cmd.Run(); err != nil {
			return err
		}

		if newLine {
			args := []string{"key", "--clearmodifiers", "Return"}
			log.WithField("args", args).Debug("Running " + xdotoolCmd)
			cmd = exec.Command(xdotoolCmd, args...)
			return cmd.Run()
		}

		return nil
	})
}

func CreateClipboardCmd(cfg *config.Config) *cobra.Command {
	var xselTargets string

	cmd := createGenerateCmd(cfg, "clipboard", "Copy the TOTP code to the clipboard", func(code string, newLine bool) error {
		for _, target := range strings.Split(xselTargets, ",") {
			var cmd *exec.Cmd
			if runtime.GOOS == config.DarwinGOOS {
				cmd = exec.Command(pbcopyCmd, "<<<", target)
			} else {
				cmd = exec.Command(xselCmd, "--input", "--"+target)
			}
			pipe, err := cmd.StdinPipe()
			if err != nil {
				return err
			}
			if err := cmd.Start(); err != nil {
				return err
			}

			if newLine {
				code += "\n"
			}
			if _, err := pipe.Write([]byte(code)); err != nil {
				return err
			}

			if err := pipe.Close(); err != nil {
				return err
			}
			if err := cmd.Wait(); err != nil {
				return err
			}
		}

		return nil
	})

	cmd.Flags().StringVar(&xselTargets, xselTargetsFlag, defaultXselTargets, "comma-separated xsel targets (clipboard, primary or secondary)")

	return cmd
}

func CreatePrintCmd(cfg *config.Config) *cobra.Command {
	return createGenerateCmd(cfg, "print", "Print the TOTP code to the stdout", func(code string, newLine bool) error {
		if newLine {
			fmt.Println(code)
		} else {
			fmt.Print(code)
		}
		return nil
	})
}

func createGenerateCmd(cfg *config.Config, name, description string, handlerFn func(code string, newLine bool) error) *cobra.Command {
	var newLine bool

	cmd := &cobra.Command{
		Use:   name + " CLIENT_ID",
		Short: description,
		Args:  cobra.ExactArgs(1),
		RunE: vault.RunOnVault(cfg, func(vlt vault.Vault, args ...string) error {
			secrets, err := vlt.GetSecrets()
			if err != nil {
				return err
			}

			if secret := secrets[args[0]]; secret != "" {
				code, err := totp.GenerateCode(secret, time.Now())
				if err != nil {
					return err
				}

				return handlerFn(code, newLine)
			}

			return vault.ErrClientNotFound
		}),
	}

	cmd.Flags().BoolVarP(&newLine, newLineFlag, "n", false, "Append a newline character to the generated TOTP code")

	return cmd
}
