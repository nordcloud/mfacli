package generate

import (
	"fmt"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"

	"os/exec"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	newLineFlag = "newline"

	xdotoolCmd = "xdotool"

	xselCmd            = "xsel"
	xselTargetsFlag    = "xsel-targets"
	defaultXselTargets = "primary,clipboard"
)

func CreateTypeCmd(cfg *config.Config) *cobra.Command {
	return createGenerateCmd(cfg, "type", "Simulate typing of the TOTP code", func(code string, newLine bool) error {
		if newLine {
			code += "\n"
		}

		args := []string{"type", code}
		log.WithField("args", args).Debug("Running " + xdotoolCmd)
		cmd := exec.Command(xdotoolCmd, args...)
		return cmd.Run()
	})
}

func CreateClipboardCmd(cfg *config.Config) *cobra.Command {
	var xselTargets string

	cmd := createGenerateCmd(cfg, "clipboard", "Copy the TOTP code to the clipboard", func(code string, newLine bool) error {
		for _, target := range strings.Split(xselTargets, ",") {
			cmd := exec.Command("xsel", "--input", "--"+target)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			clientId := args[0]

			secret, err := vault.GetSecret(clientId, cfg)
			if err != nil {
				return err
			}

			code, err := totp.GenerateCode(secret, time.Now())
			if err != nil {
				return err
			}

			return handlerFn(code, newLine)
		},
	}

	cmd.Flags().BoolVar(&newLine, newLineFlag, false, "Add a newline character to the result TOTP code string")

	return cmd
}
