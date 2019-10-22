package config

import (
	"bitbucket.org/nordcloud/mfacli/pkg/secret"
)

const (
	CommandName          = "mfacli"
	DataDirName          = "." + CommandName
	DefaultSocketName    = CommandName + ".sock"
	DefaultVaultName     = CommandName + ".vault"
	InternalRunServerCmd = "_run_server"

	FlagServerLogFile = "server-log-file"
)

var (
	Version string = "dev" // this value is injected during build
)

type Config struct {
	SocketPath    string
	VaultPath     string
	NoCache       bool
	Password      secret.SecretValue
	ServerLogFile string
}
