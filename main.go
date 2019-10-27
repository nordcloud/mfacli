package main

import (
	"github.com/nordcloud/mfacli/cmd"

	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
