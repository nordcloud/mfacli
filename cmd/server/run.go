package server

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nordcloud/mfacli/config"
	"github.com/nordcloud/mfacli/pkg/vault"
)

func CreateRunCmd(cfg *config.Config) *cobra.Command {
	var logFile *os.File

	return &cobra.Command{
		Use:    config.InternalRunServerCmd,
		Hidden: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cfg.ServerLogFile == "" {
				return
			}

			var err error
			logFile, err = os.OpenFile(cfg.ServerLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err == nil {
				log.SetOutput(logFile)
				fmt.Fprintf(os.Stderr, "Set server log file to %s\n", cfg.ServerLogFile)
			} else {
				fmt.Fprintf(os.Stderr, "Failed to set server log file: %s\n", err.Error())
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Starting server...")
			return vault.RunServer(cfg)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			if logFile != nil {
				fmt.Fprintf(os.Stderr, "Closing server log file %s\n", cfg.ServerLogFile)
				logFile.Close()
			} else {
				fmt.Fprintln(os.Stderr, "Server log file not set, nothing to close")
			}
		},
	}
}
