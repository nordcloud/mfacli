package main

import (
	"bitbucket.org/nordcloud/mfacli/cmd"

	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
