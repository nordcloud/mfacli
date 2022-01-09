package main

import (
	"os"

	"github.com/nordcloud/mfacli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
